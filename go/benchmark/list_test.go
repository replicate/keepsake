package benchmark

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/replicate/keepsake/go/pkg/concurrency"
	"github.com/replicate/keepsake/go/pkg/global"
	"github.com/replicate/keepsake/go/pkg/hash"
	"github.com/replicate/keepsake/go/pkg/param"
	"github.com/replicate/keepsake/go/pkg/project"
	"github.com/replicate/keepsake/go/pkg/repository"
)

// run a command and return stdout. If there is an error, print stdout/err and fail test
func keepsake(b *testing.B, arg ...string) string {
	// Get absolute path to built binary
	_, currentFilename, _, _ := runtime.Caller(0)
	binPath, err := filepath.Abs(path.Join(path.Dir(currentFilename), "../release", runtime.GOOS, runtime.GOARCH, "keepsake"))
	require.NoError(b, err)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(binPath, arg...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "KEEPSAKE_NO_ANALYTICS=1")
	if err := cmd.Run(); err != nil {
		fmt.Println(stdout.String())
		fmt.Println(stderr.String())
		b.Fatal(err)
	}
	return stdout.String()

}

func keepsakeList(b *testing.B, workingDir string, numExperiments int) {
	out := keepsake(b, "list", "-D", workingDir)

	// Check the output is sensible
	firstLine := strings.Split(out, "\n")[0]
	require.Contains(b, firstLine, "EXPERIMENT")
	// numExperiments + heading + trailing \n
	require.Equal(b, numExperiments+2, len(strings.Split(out, "\n")))
	// TODO: check first line is reasonable
}

func removeCache(b *testing.B, workingDir string) {
	cachePath := path.Join(workingDir, ".keepsake", "metadata-cache")
	require.NoError(b, os.RemoveAll(cachePath))
}

// Create lots of files in a working dir
func createLotsOfFiles(b *testing.B, dir string) {
	// Some 1KB files is a bit like a bit source directory
	content := []byte(strings.Repeat("a", 1000))
	for i := 1; i < 10; i++ {
		err := ioutil.WriteFile(path.Join(dir, fmt.Sprintf("%d", i)), content, 0644)
		require.NoError(b, err)
	}
}

// Create lots of experiments and checkpoints
func createLotsOfExperiments(workingDir string, repository repository.Repository, numExperiments int) error {
	numCheckpoints := 50

	maxWorkers := 25
	queue := concurrency.NewWorkerQueue(context.Background(), maxWorkers)

	for i := 0; i < numExperiments; i++ {
		err := queue.Go(func() error {
			exp := project.NewExperiment(param.ValueMap{
				"learning_rate": param.Float(0.001),
			})
			if err := exp.Save(repository); err != nil {
				return fmt.Errorf("Error saving experiment: %w", err)
			}

			if err := project.CreateHeartbeat(repository, exp.ID, time.Now().Add(-24*time.Hour)); err != nil {
				return fmt.Errorf("Error creating heartbeat: %w", err)
			}

			exp.Checkpoints = []*project.Checkpoint{}
			for j := 0; j < numCheckpoints; j++ {
				exp.Checkpoints = append(exp.Checkpoints, project.NewCheckpoint(param.ValueMap{
					"accuracy": param.Float(0.987),
				}))
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return queue.Wait()
}

func BenchmarkKeepsakeDisk(b *testing.B) {
	// Create working dir
	workingDir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(b, err)
	defer os.RemoveAll(workingDir)

	createLotsOfFiles(b, workingDir)

	// Create repository
	repositoryDir := path.Join(workingDir, ".keepsake/repository")
	repository, err := repository.NewDiskRepository(repositoryDir)
	require.NoError(b, err)
	defer os.RemoveAll(repositoryDir)

	err = createLotsOfExperiments(workingDir, repository, 10)
	require.NoError(b, err)

	b.Run("list first run with 10 experiments", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			keepsakeList(b, workingDir, 10)
		}
	})

	err = createLotsOfExperiments(workingDir, repository, 10)
	require.NoError(b, err)

	b.Run("list first run with 20 experiments", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			keepsakeList(b, workingDir, 20)
		}
	})

	err = createLotsOfExperiments(workingDir, repository, 10)
	require.NoError(b, err)

	b.Run("list first run with 30 experiments", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			keepsakeList(b, workingDir, 30)
		}
	})
}

func BenchmarkKeepsakeS3(b *testing.B) {
	// Create working dir
	workingDir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(b, err)
	defer os.RemoveAll(workingDir)

	// Disable filling working directory with files. This makes these benchmarks real slow,
	// and files in working directory now doesn't affect speed of list (and hopefully will
	// not regress...)
	// createLotsOfFiles(b, workingDir)

	// Create a bucket
	bucketName := "keepsake-test-benchmark-" + hash.Random()[0:10]
	err = repository.CreateS3Bucket(global.S3Region, bucketName)
	require.NoError(b, err)
	defer func() {
		require.NoError(b, repository.DeleteS3Bucket(global.S3Region, bucketName))
	}()
	// Even though CreateS3Bucket is supposed to wait until it exists, sometimes it doesn't
	time.Sleep(1 * time.Second)

	// keepsake.yaml
	err = ioutil.WriteFile(
		path.Join(workingDir, "keepsake.yaml"),
		[]byte(fmt.Sprintf("repository: s3://%s", bucketName)), 0644)
	require.NoError(b, err)

	// Create repository
	repository, err := repository.NewS3Repository(bucketName, "root")
	require.NoError(b, err)

	err = createLotsOfExperiments(workingDir, repository, 5)
	require.NoError(b, err)

	b.Run("list first run with 5 experiments", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			keepsakeList(b, workingDir, 5)
			removeCache(b, workingDir)
		}
	})

	keepsakeList(b, workingDir, 5)
	b.Run("list second run with 5 experiments", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			keepsakeList(b, workingDir, 5)
		}
	})

	err = createLotsOfExperiments(workingDir, repository, 5)
	require.NoError(b, err)

	b.Run("list first run with 10 experiments", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			keepsakeList(b, workingDir, 10)
			removeCache(b, workingDir)
		}
	})

	keepsakeList(b, workingDir, 10)
	b.Run("list second run with 10 experiments", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			keepsakeList(b, workingDir, 10)
		}
	})

	err = createLotsOfExperiments(workingDir, repository, 5)
	require.NoError(b, err)

	b.Run("list first run with 15 experiments", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			keepsakeList(b, workingDir, 15)
			removeCache(b, workingDir)
		}
	})

	keepsakeList(b, workingDir, 15)
	b.Run("list second run with 15 experiments", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			keepsakeList(b, workingDir, 15)
		}
	})
}

func BenchmarkKeepsakeHelp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		out := keepsake(b, "--help")
		require.Contains(b, out, "Usage:")
	}
}
