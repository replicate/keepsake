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

func TestLatestSupportedCUDA(t *testing.T) {
	tf210 := TensorflowMeta{
		TF:           "2.1.0",
		TFCPUPackage: "tensorflow==2.1.0",
		TFGPUPackage: "tensorflow==2.1.0",
		CUDA:         CUDA10_1,
		CuDNN:        CuDNN7,
		Pythons:      []Python{Py27, Py35, Py36, Py37},
	}
	torch120 := PyTorchMeta{
		Torch:             "1.2.0",
		TorchVision:       "0.4.0",
		DefaultCUDA:       CUDA10_0,
		CuDNN:             CuDNN7,
		Pythons:           []Python{Py27, Py36, Py37},
		OtherCUDASuffixes: map[CUDA]string{CUDA9_2: "+cu92"},
	}

	for _, tt := range []struct {
		frameworkMeta FrameworkMeta
		driverVersion string
		expected      CUDA
		isError       bool
	}{
		{tf210, "419.0", CUDA10_1, false},
		{tf210, "400.0", "", true},
		{torch120, "411.0", CUDA10_0, false},
		{torch120, "410.0", CUDA9_2, false},
		{torch120, "390.0", "", true},
	} {
		actual, err := LatestSupportedCUDA(tt.frameworkMeta, tt.driverVersion)
		if tt.isError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.expected, actual)
		}
	}
}

func TestPyTorchGPUPackages(t *testing.T) {
	torch120 := PyTorchMeta{
		Torch:             "1.2.0",
		TorchVision:       "0.4.0",
		DefaultCUDA:       CUDA10_0,
		CuDNN:             CuDNN7,
		Pythons:           []Python{Py27, Py36, Py37},
		OtherCUDASuffixes: map[CUDA]string{CUDA9_2: "+cu92"},
	}

	for _, tt := range []struct {
		cuda     CUDA
		expected []string
	}{
		{CUDA10_0, []string{"torch==1.2.0", "torchvision==0.4.0"}},
		{CUDA9_2, []string{"torch==1.2.0+cu92", "torchvision==0.4.0+cu92"}},
	} {
		actual := torch120.GPUPackages(tt.cuda)
		require.Equal(t, tt.expected, actual)
	}
}
