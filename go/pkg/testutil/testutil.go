package testutil

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func IP(i int) *int64 {
	i64 := int64(i)
	return &i64
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

func TrimRightLines(s string) string {
	lines := []string{}
	for _, line := range strings.Split(s, "\n") {
		lines = append(lines, strings.TrimRight(line, " "))
	}
	return strings.Join(lines, "\n")
}
