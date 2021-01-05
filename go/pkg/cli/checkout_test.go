package cli

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/replicate/replicate/go/pkg/config"
	"github.com/replicate/replicate/go/pkg/files"
	"github.com/replicate/replicate/go/pkg/hash"
	"github.com/replicate/replicate/go/pkg/project"
	"github.com/replicate/replicate/go/pkg/repository"
)

func TestCheckout(t *testing.T) {
	repoDir, err := files.TempDir("test-checkout")
	require.NoError(t, err)
	defer os.RemoveAll(repoDir)

	repo, err := repository.NewDiskRepository(repoDir)
	require.NoError(t, err)

	fixedTime, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	require.NoError(t, err)

	rand1 := hash.Random()
	rand2 := hash.Random()
	rand3 := hash.Random()

	experiment := &project.Experiment{
		ID:      "1eeeeeeeee",
		Created: fixedTime.Add(-10 * time.Minute),
		Config:  &config.Config{},
		Path:    rand1,
		Checkpoints: []*project.Checkpoint{
			{
				ID:      "1ccccccccc",
				Created: fixedTime.Add(-5 * time.Minute),
				Path:    rand2,
			},
			{
				ID:      "2aaaaaaaaa",
				Created: fixedTime.Add(-5 * time.Minute),
				Path:    "",
			},
		},
	}
	require.NoError(t, experiment.Save(repo))

	codeDir, err := files.TempDir("test-checkout-code")
	require.NoError(t, err)
	defer os.RemoveAll(codeDir)

	//  codeDir
	//  |-- rand1
	//  |-- rand2
	//  |--subdir
	//     |--rand3
	err = ioutil.WriteFile(path.Join(codeDir, rand1), []byte(rand1), 0644)
	require.NoError(t, err)

	err = ioutil.WriteFile(path.Join(codeDir, rand2), []byte(rand2), 0644)
	require.NoError(t, err)

	err = os.MkdirAll(path.Join(codeDir, "subdir"), 0755)
	require.NoError(t, err)

	err = ioutil.WriteFile(path.Join(codeDir, "subdir", rand3), []byte(rand3), 0644)
	require.NoError(t, err)

	err = repo.PutPathTar(codeDir, "experiments/1eeeeeeeee.tar.gz", rand1)
	require.NoError(t, err)

	err = repo.PutPathTar(codeDir, "checkpoints/1ccccccccc.tar.gz", rand2)
	require.NoError(t, err)

	err = repo.PutPathTar(codeDir, "checkpoints/2aaaaaaaaa.tar.gz", "")
	require.NoError(t, err)

	outputDir, err := files.TempDir("test-checkout-output")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir)

	// checkout to output directory
	err = checkoutCheckpoint(checkoutOpts{
		outputDirectory: outputDir,
		checkoutPath:    "",
		force:           true,
		repositoryURL:   "file://" + repoDir,
	}, []string{"1cc"})
	require.NoError(t, err)

	contents, err := ioutil.ReadFile(path.Join(outputDir, rand1))
	require.NoError(t, err)
	require.Equal(t, rand1, string(contents))

	contents, err = ioutil.ReadFile(path.Join(outputDir, rand2))
	require.NoError(t, err)
	require.Equal(t, rand2, string(contents))

	outputDir2, err := files.TempDir("test-checkout-output-2")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir2)

	cwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(outputDir2))
	defer func() { require.NoError(t, os.Chdir(cwd)) }()

	// checkout to working directory without replicate.yaml
	err = checkoutCheckpoint(checkoutOpts{
		outputDirectory: "",
		checkoutPath:    "",
		force:           true,
		repositoryURL:   "file://" + repoDir,
	}, []string{"1cc"})

	// no replicate.yaml, should error
	require.Error(t, err)

	require.NoError(t, ioutil.WriteFile("replicate.yaml", []byte("repository: file://"+repoDir), 0644))

	// checkout to working directory with replicate.yaml
	err = checkoutCheckpoint(checkoutOpts{
		outputDirectory: "",
		checkoutPath:    "",
		force:           true,
		repositoryURL:   "",
	}, []string{"1cc"})

	require.NoError(t, err)

	contents, err = ioutil.ReadFile(path.Join(outputDir2, rand1))
	require.NoError(t, err)
	require.Equal(t, rand1, string(contents))

	contents, err = ioutil.ReadFile(path.Join(outputDir2, rand2))
	require.NoError(t, err)
	require.Equal(t, rand2, string(contents))

	outputDir3, err := files.TempDir("test-checkout-output-3")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir3)

	// checkout a single file to output directory
	err = checkoutCheckpoint(checkoutOpts{
		outputDirectory: outputDir3,
		checkoutPath:    rand2,
		force:           true,
		repositoryURL:   "file://" + repoDir,
	}, []string{"1cc"})
	require.NoError(t, err)

	_, err = ioutil.ReadFile(path.Join(outputDir3, rand1))
	// only checking out rand2, should error
	require.Error(t, err)

	contents, err = ioutil.ReadFile(path.Join(outputDir3, rand2))
	require.NoError(t, err)
	require.Equal(t, rand2, string(contents))

	outputDir4, err := files.TempDir("test-checkout-output-4")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir4)

	// checkout a single directory
	err = checkoutCheckpoint(checkoutOpts{
		outputDirectory: outputDir4,
		checkoutPath:    "subdir",
		force:           true,
		repositoryURL:   "file://" + repoDir,
	}, []string{"2aa"})
	require.NoError(t, err)

	_, err = ioutil.ReadFile(path.Join(outputDir4, rand1))
	// checking out "subdir", should error
	require.Error(t, err)

	_, err = ioutil.ReadFile(path.Join(outputDir4, rand2))
	// checking out "subdir", should error
	require.Error(t, err)

	contents, err = ioutil.ReadFile(path.Join(outputDir4, "subdir", rand3))
	require.NoError(t, err)
	require.Equal(t, rand3, string(contents))

	outputDir5, err := files.TempDir("test-checkout-output-5")
	require.NoError(t, err)
	defer os.RemoveAll(outputDir5)

	// checkout a single file from a subdirectory
	err = checkoutCheckpoint(checkoutOpts{
		outputDirectory: outputDir5,
		checkoutPath:    "subdir/" + rand3,
		force:           true,
		repositoryURL:   "file://" + repoDir,
	}, []string{"2aa"})
	require.NoError(t, err)

	_, err = ioutil.ReadFile(path.Join(outputDir5, rand1))
	// checking out "subdir", should error
	require.Error(t, err)

	_, err = ioutil.ReadFile(path.Join(outputDir5, rand2))
	// checking out "subdir", should error
	require.Error(t, err)

	contents, err = ioutil.ReadFile(path.Join(outputDir5, "subdir", rand3))
	require.NoError(t, err)
	require.Equal(t, rand3, string(contents))
}
