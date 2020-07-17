package build

import (
	"bytes"
	"fmt"
	"path"
	"strings"

	"text/template"

	"replicate.ai/cli/pkg/assets"
	"replicate.ai/cli/pkg/baseimages"
	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/files"
	"replicate.ai/cli/pkg/global"
)

type DockerfileParams struct {
	HasPythonRequirements bool
	PythonRequirements    string
	Install               []string
}

// GenerateDockerfile generates a Dockerfile from a template,
// based on a replicate config
func GenerateDockerfile(conf *config.Config, sourceDir string) (string, error) {
	hasPythonRequirements, err := files.FileExists(path.Join(sourceDir, conf.PythonRequirements))
	if err != nil {
		return "", err
	}

	params := &DockerfileParams{
		HasPythonRequirements: hasPythonRequirements,
		PythonRequirements:    conf.PythonRequirements,
		Install:               conf.Install,
	}
	contents := assets.MustAsset("Dockerfile")
	tmpl, err := template.New("Dockerfile").Parse(string(contents))
	if err != nil {
		panic(fmt.Errorf("Failed to parse Dockerfile template got error: %s", err))
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, params); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// GetBaseImage returns a base image by first searching
// requirements.txt for a framework (torch/tensorflow[-gpu]), and then
// adding a matching cuda version if the host has a cuda driver
func GetBaseImage(conf *config.Config, sourceDir string, hostCUDADriverVersion string) (baseImage *baseimages.BaseImage, err error) {
	frameworkMeta, err := getRequiredFramework(conf, sourceDir)
	if err != nil {
		return nil, err
	}

	if hostCUDADriverVersion != "" {
		cuda, cuDNN, err := getCUDAVersion(conf, frameworkMeta, hostCUDADriverVersion)
		if err != nil {
			return nil, err
		}
		ubuntu := baseimages.LatestUbuntuForCUDA[cuda]
		return &baseimages.BaseImage{
			Registry:      baseimages.DefaultRegistry,
			Project:       baseimages.DefaultProject,
			Version:       baseimages.DefaultVersion,
			Python:        baseimages.Python(conf.Python),
			Ubuntu:        ubuntu,
			FrameworkMeta: frameworkMeta,
			CUDA:          cuda,
			CuDNN:         cuDNN,
		}, nil
	}

	return &baseimages.BaseImage{
		Registry:      baseimages.DefaultRegistry,
		Project:       baseimages.DefaultProject,
		Version:       baseimages.DefaultVersion,
		Python:        baseimages.Python(conf.Python),
		Ubuntu:        baseimages.LatestUbuntu(),
		FrameworkMeta: frameworkMeta,
	}, nil
}

func getCUDAVersion(conf *config.Config, frameworkMeta baseimages.FrameworkMeta, hostCUDADriverVersion string) (baseimages.CUDA, baseimages.CuDNN, error) {
	if frameworkMeta == nil {
		// if no framework is specified, use cuda from config
		// or the latest cuda version compatible with the host

		if conf.CUDA != "" {
			cuda := baseimages.CUDA(conf.CUDA)
			driverCompatible, err := baseimages.CUDADriverIsCompatible(cuda, hostCUDADriverVersion)
			if err != nil {
				return "", "", err
			}
			if !driverCompatible {
				return "", "", cudaCompatibilityError(cuda, hostCUDADriverVersion)
			}
			cuDNN := baseimages.LatestCuDNNForCUDA[cuda]

			return cuda, cuDNN, nil
		}

		cuda, err := baseimages.LatestCUDAForDriverVersion(hostCUDADriverVersion)
		if err != nil {
			return "", "", err
		}
		cuDNN := baseimages.LatestCuDNNForCUDA[cuda]

		console.Info("No CUDA version specified in %s, using CUDA %s and CuDNN %s", global.ConfigFilename, cuda, cuDNN)

		return cuda, cuDNN, nil
	}
	// if a framework is specified, automatically use the matching
	// cuda version

	cuda := frameworkMeta.GetCUDA()
	cuDNN := frameworkMeta.GetCuDNN()

	// ensure cuda version in config matches the
	// framework's compatible cuda version
	if conf.CUDA != "" && cuda != baseimages.CUDA(conf.CUDA) {
		return "", "", fmt.Errorf("%s==%s requires CUDA %s", frameworkMeta.Name(), frameworkMeta.Version(), cuda)
	}

	// ensure cuda version is compatible with host
	driverCompatible, err := baseimages.CUDADriverIsCompatible(cuda, hostCUDADriverVersion)
	if err != nil {
		return "", "", err
	}
	if !driverCompatible {
		return "", "", cudaCompatibilityError(cuda, hostCUDADriverVersion)
	}

	return cuda, cuDNN, err
}

// return the *first* framework (torch or tensorflow[-gpu]) found in
// requirements.txt, or nil if no framework is found
func getRequiredFramework(conf *config.Config, sourceDir string) (baseimages.FrameworkMeta, error) {
	requirementsLines, err := conf.ReadPythonRequirements(sourceDir)
	if err != nil {
		return nil, err
	}
	for _, line := range requirementsLines {
		parts := strings.SplitN(line, "==", 2)
		name := parts[0]
		version := ""
		if len(parts) == 2 {
			version = parts[1]
		}
		switch name {
		case "tensorflow":
			fallthrough
		case "tensorflow-gpu":
			return getFrameworkMeta(baseimages.Tensorflow, version)
		case "torch":
			return getFrameworkMeta(baseimages.PyTorch, version)
		}
	}

	return nil, nil
}

func getFrameworkMeta(frameworkName string, version string) (baseimages.FrameworkMeta, error) {
	// if no framework version is specified, use the latest
	if version == "" {
		version = baseimages.LatestFrameworkVersion(baseimages.Tensorflow)
	}
	meta, err := baseimages.FrameworkMetaFor(frameworkName, version)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func cudaCompatibilityError(requestedCUDA baseimages.CUDA, hostCUDADriverVersion string) error {
	return fmt.Errorf("CUDA %s is not compatible with your host's CUDA driver version %s.\nPlease refer to https://docs.nvidia.com/deploy/cuda-compatibility/index.html#binary-compatibility__table-toolkit-driver for the correct driver version.", requestedCUDA, hostCUDADriverVersion)
}
