package benchmark

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"replicate.ai/cli/pkg/commit"
	"replicate.ai/cli/pkg/experiment"
	"replicate.ai/cli/pkg/param"

	"replicate.ai/cli/pkg/storage"
)

func BenchmarkList(b *testing.B) {
	// Create working dir
	workingDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(b, err)
	defer os.RemoveAll(workingDir)

	// 1000 1KB files is a bit like a bit source directory
	content := []byte(strings.Repeat("a", 1000))
	for i := 1; i < 1000; i++ {
		err := ioutil.WriteFile(path.Join(workingDir, fmt.Sprintf("%d", i)), content, 0644)
		require.NoError(b, err)
	}

	// Create storage
	storageDir := path.Join(workingDir, ".replicate/storage")
	storage, err := storage.NewDiskStorage(storageDir)
	require.NoError(b, err)
	defer os.RemoveAll(storageDir)

	for i := 0; i < 100; i++ {
		exp := experiment.NewExperiment(map[string]*param.Value{
			"learning_rate": param.Float(0.001),
		})
		err := exp.Save(storage)
		require.NoError(b, err)

		for j := 0; j < 100; j++ {
			com := commit.NewCommit(*exp, map[string]*param.Value{
				"accuracy": param.Float(0.987),
			})
			err := com.Save(storage, workingDir)
			require.NoError(b, err)
		}
	}

	// So we're not timing setup
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		cmd := exec.Command("replicate", "list", "-D", workingDir)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		if err := cmd.Run(); err != nil {
			fmt.Println(stdout.String())
			fmt.Println(stderr.String())
			b.Fatal(err)
		}

		// Check the output is sensible
		stdoutStr := stdout.String()
		firstLine := strings.Split(stdoutStr, "\n")[0]
		require.Contains(b, firstLine, "experiment")
		// 100 experiments
		require.Equal(b, 102, len(strings.Split(stdoutStr, "\n")))
		// TODO: check first line is reasonable
	}

	// Stop timer before deferred cleanup
	b.StopTimer()
}
