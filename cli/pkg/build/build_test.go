package build

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"

	"replicate.ai/cli/pkg/baseimages"
	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/files"
)

func TestGenerateDockerfile(t *testing.T) {
	conf := &config.Config{
		Python:             "3.7",
		PythonRequirements: "requirements.txt",
		Install: []string{
			"apt-get update",
			"apt-get install -y ffmpeg",
		},
	}

	tmpDir, err := files.TempDir("test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = ioutil.WriteFile(path.Join(tmpDir, "requirements.txt"), []byte("tensorflow==2.2.0"), 0644)
	require.NoError(t, err)

	dockerfile, err := GenerateDockerfile(conf, tmpDir)
	require.NoError(t, err)

	expected := `ARG BASE_IMAGE
FROM $BASE_IMAGE

ARG HAS_GPU
ENV HAS_GPU=$HAS_GPU

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update
RUN apt-get install -y ffmpeg

COPY "requirements.txt" /tmp/requirements.txt
RUN pip install -r /tmp/requirements.txt

# FIXME: temporary, until this is on pypi or we find a better temporary spot
RUN pip install https://storage.googleapis.com/replicate-python-dev/replicate-0.0.4.tar.gz

COPY . /code
WORKDIR /code
`
	require.Equal(t, expected, dockerfile)
}

func TestGetRequiredFramework(t *testing.T) {
	noFramework := `foo
#torch
bar`
	tfNoVersion := `foo
tensorflow
bar`
	tfWithVersion := `tensorflow==2.1.0`
	torchWithVersion := `#foo
torch==1.2.0

`
	torchWithBadVersion := `torch==bad`

	tf220 := baseimages.TensorflowMeta{
		TF:           "2.2.0",
		TFCPUPackage: "tensorflow==2.2.0",
		TFGPUPackage: "tensorflow==2.2.0",
		CUDA:         baseimages.CUDA10_1,
		CuDNN:        baseimages.CuDNN7,
		Pythons:      []baseimages.Python{baseimages.Py35, baseimages.Py37, baseimages.Py38},
	}
	tf210 := baseimages.TensorflowMeta{
		TF:           "2.1.0",
		TFCPUPackage: "tensorflow==2.1.0",
		TFGPUPackage: "tensorflow==2.1.0",
		CUDA:         baseimages.CUDA10_1,
		CuDNN:        baseimages.CuDNN7,
		Pythons:      []baseimages.Python{baseimages.Py37, baseimages.Py27},
	}
	torch120 := baseimages.PyTorchMeta{
		Torch:       "1.2.0",
		TorchVision: "0.4.0",
		CUDA:        baseimages.CUDA10_0,
		CuDNN:       baseimages.CuDNN7,
		Pythons:     []baseimages.Python{baseimages.Py37, baseimages.Py36, baseimages.Py27},
	}

	for _, tt := range []struct {
		requirements string
		expected     baseimages.FrameworkMeta
		isError      bool
	}{
		{noFramework, nil, false},
		{tfNoVersion, tf220, false},
		{tfWithVersion, tf210, false},
		{torchWithVersion, torch120, false},
		{torchWithBadVersion, nil, true},
	} {
		tmpDir, err := files.TempDir("test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		err = ioutil.WriteFile(path.Join(tmpDir, "requirements.txt"), []byte(tt.requirements), 0644)
		require.NoError(t, err)

		conf := &config.Config{
			PythonRequirements: "requirements.txt",
		}

		frameworkMeta, err := getRequiredFramework(conf, tmpDir)
		if tt.isError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}

		require.Equal(t, tt.expected, frameworkMeta)
	}
}

func TestGetCUDAVersion(t *testing.T) {
	for _, tt := range []struct {
		confCUDA              string
		framework             string
		frameworkVersion      string
		hostCUDADriverVersion string
		expectedCUDA          baseimages.CUDA
		expectedCuDNN         baseimages.CuDNN
		isError               bool
	}{
		{"", "", "", "450.0", baseimages.CUDA10_2, baseimages.CuDNN8, false},
		{"", "", "", "250.0", "", "", true},
		{"", baseimages.PyTorch, "1.4.0", "450.0", baseimages.CUDA10_1, baseimages.CuDNN7, false},
		{"", baseimages.PyTorch, "1.4.0", "400.0", "", "", true},
		{"10.1", baseimages.PyTorch, "1.4.0", "450.0", baseimages.CUDA10_1, baseimages.CuDNN7, false},
		{"10.2", baseimages.PyTorch, "1.4.0", "450.0", "", "", true},
		{"10.1", "", "", "450.0", baseimages.CUDA10_1, baseimages.CuDNN7, false},
		{"10.1", "", "", "400.0", "", "", true},
	} {
		conf := &config.Config{
			CUDA: tt.confCUDA,
		}
		var frameworkMeta baseimages.FrameworkMeta
		var err error
		if tt.framework != "" {
			frameworkMeta, err = baseimages.FrameworkMetaFor(tt.framework, tt.frameworkVersion)
			require.NoError(t, err)
		}

		cuda, cuDNN, err := getCUDAVersion(conf, frameworkMeta, tt.hostCUDADriverVersion)
		if tt.isError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}

		require.Equal(t, tt.expectedCUDA, cuda)
		require.Equal(t, tt.expectedCuDNN, cuDNN)
	}
}

func TestGetBaseImage(t *testing.T) {
	requirementsTf := `foo
tensorflow==2.1.0
bar
`
	requirementsTorch := `
torch==1.5.0
`
	requirementsEmpty := ""

	for _, tt := range []struct {
		requirements          string
		confCUDA              string
		confPython            string
		hostCUDADriverVersion string
		expected              string
		isError               bool
	}{
		{requirementsTf, "", "3.5", "450.0", "us.gcr.io/replicate/base-ubuntu18.04-python3.5-cuda10.1-cudnn7-tf2.1.0:0.3", false},
		{requirementsTf, "10.1", "3.5", "450.0", "us.gcr.io/replicate/base-ubuntu18.04-python3.5-cuda10.1-cudnn7-tf2.1.0:0.3", false},
		{requirementsTf, "10.2", "3.5", "450.0", "", true},
		{requirementsTf, "10.1", "3.5", "350.0", "", true},
		{requirementsEmpty, "10.1", "3.7", "450.0", "us.gcr.io/replicate/base-ubuntu18.04-python3.7-cuda10.1-cudnn7:0.3", false},
		{requirementsTorch, "", "3.8", "450.0", "us.gcr.io/replicate/base-ubuntu18.04-python3.8-cuda10.2-cudnn7-pytorch1.5.0:0.3", false},
		{requirementsTorch, "", "3.8", "", "us.gcr.io/replicate/base-ubuntu18.04-python3.8-cpu-pytorch1.5.0:0.3", false},
	} {
		conf := &config.Config{
			Python:             tt.confPython,
			PythonRequirements: "requirements.txt",
			CUDA:               tt.confCUDA,
		}

		tmpDir, err := files.TempDir("test")
		require.NoError(t, err)
		defer os.RemoveAll(tmpDir)

		err = ioutil.WriteFile(path.Join(tmpDir, "requirements.txt"), []byte(tt.requirements), 0644)
		require.NoError(t, err)

		baseImage, err := GetBaseImage(conf, tmpDir, tt.hostCUDADriverVersion)
		if tt.isError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.expected, baseImage.RepositoryName())
		}
	}
}
