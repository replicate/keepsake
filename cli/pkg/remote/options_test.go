package remote

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseHost(t *testing.T) {
	options, err := ParseHost("1.2.3.4")
	require.NoError(t, err)
	require.Equal(t, &Options{Username: "", Host: "1.2.3.4", Port: 0}, options)

	options, err = ParseHost("ben@1.2.3.4")
	require.NoError(t, err)
	require.Equal(t, &Options{Username: "ben", Host: "1.2.3.4", Port: 0}, options)

	options, err = ParseHost("ben@1.2.3.4:5678")
	require.NoError(t, err)
	require.Equal(t, &Options{Username: "ben", Host: "1.2.3.4", Port: 5678}, options)
}
