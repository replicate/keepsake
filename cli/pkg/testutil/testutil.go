package testutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func IP(i int) *int {
	return &i
}

func FP(f float64) *float64 {
	return &f
}

func SP(s string) *string {
	return &s
}

func BP(b bool) *bool {
	return &b
}

func RequireErrContains(t *testing.T, err error, s string) {
	require.Contains(t, fmt.Sprintf("%s", err), s)
}

func Write(dir string, filename string, contents string, args ...interface{}) {
	contents = fmt.Sprintf(contents, args...)
	err := ioutil.WriteFile(path.Join(dir, filename), []byte(contents), 0644)
	if err != nil {
		// HACK (bfirsh): do this more gracefully
		fmt.Fprintf(os.Stderr, "%s", err)
		os.Exit(1)
	}
}

func RequireFilesEqual(t *testing.T, path1 string, path2 string) {
	contents1, err := ioutil.ReadFile(path1)
	require.NoError(t, err)
	contents2, err := ioutil.ReadFile(path2)
	require.NoError(t, err)

	require.Equal(t, string(contents1), string(contents2))
}

func RequireFileContentsEqual(t *testing.T, path string, expected string) {
	contents, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, expected, string(contents))
}
