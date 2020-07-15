package baseimages

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCUDADriverIsCompatible(t *testing.T) {
	for _, tt := range []struct {
		cuda          CUDA
		driverVersion string
		expected      bool
	}{
		{CUDA11_0, "460.0", true},
		{CUDA11_0, "460", true},
		{CUDA11_0, "450.20.30", false},
		{CUDA9_2, "396.26", true},
		{CUDA9_2, "396.25", false},
	} {
		compatible, err := CUDADriverIsCompatible(tt.cuda, tt.driverVersion)
		require.NoError(t, err)
		require.Equal(t, tt.expected, compatible)
	}
}

func TestLatestCUDAForDriverVersion(t *testing.T) {
	for _, tt := range []struct {
		driverVersion string
		expected      CUDA
	}{
		{"460.0", CUDA11_0},
		{"460", CUDA11_0},
		{"450.20.30", CUDA10_2},
		{"396.26", CUDA9_2},
		{"396.25", CUDA9_1},
	} {
		cuda, err := LatestCUDAForDriverVersion(tt.driverVersion)
		require.NoError(t, err)
		require.Equal(t, tt.expected, cuda, tt.driverVersion)
	}
}
