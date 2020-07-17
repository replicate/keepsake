package baseimages

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-version"
)

type BaseImage struct {
	Registry      string
	Project       string
	Version       string
	Python        Python
	Ubuntu        Ubuntu
	FrameworkMeta FrameworkMeta
	CUDA          CUDA
	CuDNN         CuDNN
}

// RepositoryName returns a fully qualified docker image repository
// name, including the version tag
func (i *BaseImage) RepositoryName() string {
	parts := []string{
		"base",
		"ubuntu" + string(i.Ubuntu),
		"python" + string(i.Python),
	}

	if i.CUDA == "" {
		parts = append(parts, "cpu")
	} else {
		parts = append(parts, "cuda"+string(i.CUDA), "cudnn"+string(i.CuDNN))
	}

	if i.FrameworkMeta != nil {
		parts = append(parts, i.FrameworkMeta.FrameworkString())
	}

	name := strings.Join(parts, "-")
	repo := i.Registry + "/" + i.Project + "/" + name + ":" + i.Version
	return repo
}

// Python version
type Python string

// CUDA version
type CUDA string

// CuDNN version
type CuDNN string

// Ubuntu version
type Ubuntu string

// Pair of CUDA / CuDNN version
type CUDACuDNN struct {
	CUDA  CUDA
	CuDNN CuDNN
}

func (c CUDACuDNN) String() string {
	return fmt.Sprintf("cuda%s-cudnn%s", c.CUDA, c.CuDNN)
}

// Triple of CUDA / CuDNN / Ubuntu version
type CUDACuDNNUbuntu struct {
	CUDA   CUDA
	CuDNN  CuDNN
	Ubuntu Ubuntu
}

// Tuple of Python / CUDA / CuDNN / Ubuntu version
type PythonCUDACuDNNUbuntu struct {
	Python Python
	CUDA   CUDA
	CuDNN  CuDNN
	Ubuntu Ubuntu
}

// Pair of Python / Ubuntu version
type PythonUbuntu struct {
	Python Python
	Ubuntu Ubuntu
}

// FrameworkMeta is a (poorly named?) interface containing
// a framework (torch/tensorflow), framework version, cuda/cudnn
// compatibility information, python packages and compatible python
// versions.
type FrameworkMeta interface {
	Name() string
	Version() string
	GetCUDA() CUDA
	GetCuDNN() CuDNN
	CPUPackages() []string
	GPUPackages() []string
	FrameworkString() string
	PythonVersions() []Python
}

type TensorflowMeta struct {
	TF           string
	TFCPUPackage string
	TFGPUPackage string
	CUDA         CUDA
	CuDNN        CuDNN
	Pythons      []Python
}

func (c TensorflowMeta) CPUPackages() []string {
	return []string{c.TFCPUPackage}
}

func (c TensorflowMeta) GPUPackages() []string {
	return []string{c.TFGPUPackage}
}

func (c TensorflowMeta) GetCUDA() CUDA {
	return c.CUDA
}

func (c TensorflowMeta) GetCuDNN() CuDNN {
	return c.CuDNN
}

func (c TensorflowMeta) FrameworkString() string {
	return fmt.Sprintf("tf%s", c.TF)
}

func (c TensorflowMeta) Name() string {
	return Tensorflow
}

func (c TensorflowMeta) PythonVersions() []Python {
	return c.Pythons
}

func (c TensorflowMeta) Version() string {
	return c.TF
}

type PyTorchMeta struct {
	Torch       string
	TorchVision string
	CUDA        CUDA
	CuDNN       CuDNN
	Pythons     []Python
}

func (c PyTorchMeta) CPUPackages() []string {
	return []string{
		fmt.Sprintf("torch==%s", c.Torch),
		fmt.Sprintf("torchvision==%s", c.TorchVision),
	}
}

func (c PyTorchMeta) GPUPackages() []string {
	return c.CPUPackages()
}

func (c PyTorchMeta) GetCUDA() CUDA {
	return c.CUDA
}

func (c PyTorchMeta) GetCuDNN() CuDNN {
	return c.CuDNN
}

func (c PyTorchMeta) FrameworkString() string {
	return fmt.Sprintf("pytorch%s", c.Torch)
}

func (c PyTorchMeta) Name() string {
	return PyTorch
}

func (c PyTorchMeta) PythonVersions() []Python {
	return c.Pythons
}

func (c PyTorchMeta) Version() string {
	return c.Torch
}

var (
	PythonVersions = []Python{
		Py38,
		Py37,
		Py36,
		Py35,
		Py27,
	}

	// from https://www.tensorflow.org/install/source#tested_build_configurations,
	// though some python versions are actually missing when you
	// try to install tensorflow. e.g. py3.6.11 doesn't have a
	// pypi candidate for tensorflow==2.2.0 on linux
	TensorflowMetas = []TensorflowMeta{
		{"2.2.0", "tensorflow==2.2.0", "tensorflow==2.2.0", CUDA10_1, CuDNN7, []Python{Py35, Py37, Py38}},
		{"2.1.0", "tensorflow==2.1.0", "tensorflow==2.1.0", CUDA10_1, CuDNN7, []Python{Py37, Py27}},
		{"2.0.1", "tensorflow==2.0.1", "tensorflow==2.0.1", CUDA10_0, CuDNN7, []Python{Py37}},
		{"2.0.0", "tensorflow==2.0.0", "tensorflow==2.0.0", CUDA10_0, CuDNN7, []Python{Py37, Py27}},
		{"1.15.2", "tensorflow==1.15.2", "tensorflow-gpu==1.15.2", CUDA10_0, CuDNN7, []Python{Py37}},
		{"1.15.0", "tensorflow==1.15.0", "tensorflow-gpu==1.15.0", CUDA10_0, CuDNN7, []Python{Py37, Py27}},
		{"1.14.0", "tensorflow==1.14.0", "tensorflow-gpu==1.14.0", CUDA10_0, CuDNN7, []Python{Py37, Py36, Py35, Py27}},
		{"1.13.2", "tensorflow==1.13.2", "tensorflow-gpu==1.13.2", CUDA10_0, CuDNN7, []Python{Py37, Py36, Py35, Py27}},
		{"1.13.1", "tensorflow==1.13.1", "tensorflow-gpu==1.13.1", CUDA10_0, CuDNN7, []Python{Py37, Py36, Py35, Py27}},
		{"1.12.3", "tensorflow==1.12.3", "tensorflow-gpu==1.12.3", CUDA9_0, CuDNN7, []Python{Py36, Py35, Py27}},
		{"1.12.2", "tensorflow==1.12.2", "tensorflow-gpu==1.12.2", CUDA9_0, CuDNN7, []Python{Py36, Py35, Py27}},
		{"1.12.0", "tensorflow==1.12.0", "tensorflow-gpu==1.12.0", CUDA9_0, CuDNN7, []Python{Py36, Py35, Py27}},
		{"1.11.0", "tensorflow==1.11.0", "tensorflow-gpu==1.11.0", CUDA9_0, CuDNN7, []Python{Py36, Py35, Py27}},
		{"1.10.1", "tensorflow==1.10.1", "tensorflow-gpu==1.10.1", CUDA9_0, CuDNN7, []Python{Py36, Py35, Py27}},
		{"1.10.0", "tensorflow==1.10.0", "tensorflow-gpu==1.10.0", CUDA9_0, CuDNN7, []Python{Py36, Py35, Py27}},
		{"1.9.0", "tensorflow==1.9.0", "tensorflow-gpu==1.9.0", CUDA9_0, CuDNN7, []Python{Py36, Py35, Py27}},
		{"1.8.0", "tensorflow==1.8.0", "tensorflow-gpu==1.8.0", CUDA9_0, CuDNN7, []Python{Py36, Py35, Py27}},
		{"1.7.1", "tensorflow==1.7.1", "tensorflow-gpu==1.7.1", CUDA9_0, CuDNN7, []Python{Py36, Py35, Py27}},
		{"1.7.0", "tensorflow==1.7.0", "tensorflow-gpu==1.7.0", CUDA9_0, CuDNN7, []Python{Py36, Py35, Py27}},
		{"1.6.0", "tensorflow==1.6.0", "tensorflow-gpu==1.6.0", CUDA9_0, CuDNN7, []Python{Py36, Py35, Py27}},
		{"1.5.1", "tensorflow==1.5.1", "tensorflow-gpu==1.5.1", CUDA9_0, CuDNN7, []Python{Py36, Py35, Py27}},
		{"1.5.0", "tensorflow==1.5.0", "tensorflow-gpu==1.5.0", CUDA9_0, CuDNN7, []Python{Py36, Py35, Py27}},
		{"1.4.1", "tensorflow==1.4.1", "tensorflow-gpu==1.4.1", CUDA8_0, CuDNN6, []Python{Py36, Py35, Py27}},
		{"1.4.0", "tensorflow==1.4.0", "tensorflow-gpu==1.4.0", CUDA8_0, CuDNN6, []Python{Py36, Py35, Py27}},
		{"1.3.0", "tensorflow==1.3.0", "tensorflow-gpu==1.3.0", CUDA8_0, CuDNN6, []Python{Py36, Py35, Py27}},
		{"1.2.1", "tensorflow==1.2.1", "tensorflow-gpu==1.2.1", CUDA8_0, CuDNN5, []Python{Py36, Py35, Py27}},
		{"1.2.0", "tensorflow==1.2.0", "tensorflow-gpu==1.2.0", CUDA8_0, CuDNN5, []Python{Py36, Py35, Py27}},
		{"1.1.0", "tensorflow==1.1.0", "tensorflow-gpu==1.1.0", CUDA8_0, CuDNN5, []Python{Py36, Py35, Py27}},
		{"1.0.1", "tensorflow==1.0.1", "tensorflow-gpu==1.0.1", CUDA8_0, CuDNN5, []Python{Py36, Py35, Py27}},
		{"1.0.0", "tensorflow==1.0.0", "tensorflow-gpu==1.0.0", CUDA8_0, CuDNN5, []Python{Py36, Py35, Py27}},
		{"0.12.1", "tensorflow==0.12.1", "tensorflow-gpu==0.12.1", CUDA8_0, CuDNN5, []Python{Py36, Py35, Py27}},
		{"0.12.0", "tensorflow==0.12.0", "tensorflow-gpu==0.12.0", CUDA8_0, CuDNN5, []Python{Py35, Py27}},
	}

	PyTorchMetas = []PyTorchMeta{
		{"1.5.1", "0.6.1", CUDA10_2, CuDNN7, []Python{Py38, Py37, Py36}},
		{"1.5.0", "0.6.0", CUDA10_2, CuDNN7, []Python{Py38, Py37, Py36}},
		{"1.4.0", "0.5.0", CUDA10_1, CuDNN7, []Python{Py38, Py37, Py36, Py27}},
		{"1.2.0", "0.4.0", CUDA10_0, CuDNN7, []Python{Py37, Py36, Py27}},
		{"1.1.0", "0.3.0", CUDA10_0, CuDNN7, []Python{Py37, Py36, Py27}},
		{"1.0.1", "0.2.2", CUDA10_0, CuDNN7, []Python{Py37, Py36, Py35, Py27}},
		{"1.0.0", "0.2.1", CUDA10_0, CuDNN7, []Python{Py37, Py36, Py35, Py27}},
	}

	CUDAImages = map[CUDACuDNNUbuntu]string{
		CUDACuDNNUbuntu{CUDA10_2, CuDNN8, Ubuntu18_04}: "nvidia/cuda:10.2-cudnn8-devel-ubuntu18.04",
		CUDACuDNNUbuntu{CUDA10_2, CuDNN8, Ubuntu16_04}: "nvidia/cuda:10.2-cudnn8-devel-ubuntu16.04",
		CUDACuDNNUbuntu{CUDA10_2, CuDNN7, Ubuntu18_04}: "nvidia/cuda:10.2-cudnn7-devel-ubuntu18.04",
		CUDACuDNNUbuntu{CUDA10_2, CuDNN7, Ubuntu16_04}: "nvidia/cuda:10.2-cudnn7-devel-ubuntu16.04",
		CUDACuDNNUbuntu{CUDA10_1, CuDNN7, Ubuntu18_04}: "nvidia/cuda:10.1-cudnn7-devel-ubuntu18.04",
		CUDACuDNNUbuntu{CUDA10_1, CuDNN7, Ubuntu16_04}: "nvidia/cuda:10.1-cudnn7-devel-ubuntu16.04",
		CUDACuDNNUbuntu{CUDA10_0, CuDNN7, Ubuntu18_04}: "nvidia/cuda:10.0-cudnn7-devel-ubuntu18.04",
		CUDACuDNNUbuntu{CUDA10_0, CuDNN7, Ubuntu16_04}: "nvidia/cuda:10.0-cudnn7-devel-ubuntu16.04",
		CUDACuDNNUbuntu{CUDA9_2, CuDNN7, Ubuntu18_04}:  "nvidia/cuda:9.2-cudnn7-devel-ubuntu18.04",
		CUDACuDNNUbuntu{CUDA9_2, CuDNN7, Ubuntu16_04}:  "nvidia/cuda:9.2-cudnn7-devel-ubuntu16.04",
		CUDACuDNNUbuntu{CUDA9_1, CuDNN7, Ubuntu16_04}:  "nvidia/cuda:9.1-cudnn7-devel-ubuntu16.04",
		CUDACuDNNUbuntu{CUDA9_0, CuDNN7, Ubuntu16_04}:  "nvidia/cuda:9.0-cudnn7-devel-ubuntu16.04",
		CUDACuDNNUbuntu{CUDA8_0, CuDNN7, Ubuntu16_04}:  "nvidia/cuda:8.0-cudnn7-devel-ubuntu16.04",
		CUDACuDNNUbuntu{CUDA8_0, CuDNN6, Ubuntu16_04}:  "nvidia/cuda:8.0-cudnn6-devel-ubuntu16.04",
		CUDACuDNNUbuntu{CUDA8_0, CuDNN5, Ubuntu16_04}:  "nvidia/cuda:8.0-cudnn5-devel-ubuntu16.04",
	}

	// from https://docs.nvidia.com/deploy/cuda-compatibility/index.html#binary-compatibility__table-toolkit-driver
	CUDAMinDriverVersion = map[CUDA]string{
		CUDA11_0: "450.36.06",
		CUDA10_2: "440.33",
		CUDA10_1: "418.39",
		CUDA10_0: "410.48",
		CUDA9_2:  "396.26",
		CUDA9_1:  "390.46",
		CUDA9_0:  "384.81",
		CUDA8_0:  "375.26",
	}

	// only build a single ubuntu image per framework version,
	// there's no good reason for anyone to use an old
	// ubuntu. this map is based on the CUDAImages mapping above.
	LatestUbuntuForCUDA = map[CUDA]Ubuntu{
		CUDA10_2: Ubuntu18_04,
		CUDA10_1: Ubuntu18_04,
		CUDA10_0: Ubuntu18_04,
		CUDA9_2:  Ubuntu18_04,
		CUDA9_1:  Ubuntu16_04,
		CUDA9_0:  Ubuntu16_04,
		CUDA8_0:  Ubuntu16_04,
	}

	LatestCuDNNForCUDA = map[CUDA]CuDNN{
		CUDA10_2: CuDNN8,
		CUDA10_1: CuDNN7,
		CUDA10_0: CuDNN7,
		CUDA9_2:  CuDNN7,
		CUDA9_1:  CuDNN7,
		CUDA9_0:  CuDNN7,
		CUDA8_0:  CuDNN7,
	}

	UbuntuImages = map[Ubuntu]string{
		Ubuntu16_04: "ubuntu:16.04",
		Ubuntu18_04: "ubuntu:18.04",
	}
)

func FrameworkMetaFor(frameworkName string, frameworkVersion string) (FrameworkMeta, error) {
	m := FrameworkVersionMap(frameworkName)
	if meta, ok := m[frameworkVersion]; ok {
		return meta, nil
	}
	return nil, fmt.Errorf("Unknown %s version: %s", frameworkName, frameworkVersion)
}

func FrameworkVersionMap(frameworkName string) map[string]FrameworkMeta {
	m := map[string]FrameworkMeta{}
	for _, f := range FrameworkMetas(frameworkName) {
		m[f.Version()] = f
	}
	return m
}

func FrameworkMetas(frameworkName string) []FrameworkMeta {
	metas := []FrameworkMeta{}
	if frameworkName == Tensorflow {
		for _, meta := range TensorflowMetas {
			metas = append(metas, meta)
		}
		return metas
	}
	if frameworkName == PyTorch {
		for _, meta := range PyTorchMetas {
			metas = append(metas, meta)
		}
		return metas
	}
	panic(fmt.Sprintf("Unknown framework: %s", frameworkName))
}

func CUDADriverIsCompatible(cuda CUDA, driverVersion string) (bool, error) {
	minDriverVersion := CUDAMinDriverVersion[cuda]
	driverVer, err := version.NewVersion(driverVersion)
	if err != nil {
		return false, err
	}
	minDriverVer, err := version.NewVersion(minDriverVersion)
	if err != nil {
		return false, err
	}
	return driverVer.GreaterThanOrEqual(minDriverVer), nil
}

func LatestCUDAForDriverVersion(driverVersion string) (CUDA, error) {
	for _, cuda := range []CUDA{
		CUDA11_0,
		CUDA10_2,
		CUDA10_1,
		CUDA10_0,
		CUDA9_2,
		CUDA9_1,
		CUDA9_0,
		CUDA8_0,
	} {
		compatible, err := CUDADriverIsCompatible(cuda, driverVersion)
		if err != nil {
			return "", err
		}
		if compatible {
			return cuda, nil
		}
	}
	return "", fmt.Errorf("No compatible CUDA version found for CUDA driver version %s. Please upgrade to a more recent CUDA driver", driverVersion)
}

func LatestFrameworkVersion(framework string) string {
	metas := FrameworkMetas(framework)
	return metas[0].Version()
}

func LatestUbuntu() Ubuntu {
	return Ubuntu18_04
}
