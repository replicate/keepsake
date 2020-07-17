package baseimages

// TODO: move this file to base-image-builder

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"
	"text/template"
	"time"

	"replicate.ai/cli/pkg/assets"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/files"
)

type Builder interface {
	Build(dir string, image string) error
	ImageExists(image string) bool
	Verbose() bool
}

type DockerfileParams struct {
	BaseImage      string
	PythonVersion  Python
	PythonPackages string
}

func BuildBaseImages(builder Builder, project string, registry string, version string, waitTime time.Duration) error {

	// build everything in parallel using waitgroups
	wg := sync.WaitGroup{}
	gpuBaseImages := map[PythonCUDACuDNNUbuntu]BaseImage{}
	cpuBaseImages := map[PythonUbuntu]BaseImage{}

	// build gpu base images without framework
	for _, py := range PythonVersions {
		for cuVer, cudaImage := range CUDAImages {
			image := BaseImage{
				Registry:      registry,
				Project:       project,
				Version:       version,
				Python:        py,
				Ubuntu:        cuVer.Ubuntu,
				FrameworkMeta: nil,
				CUDA:          cuVer.CUDA,
				CuDNN:         cuVer.CuDNN,
			}
			gpuBaseImages[PythonCUDACuDNNUbuntu{
				Python: py,
				CUDA:   cuVer.CUDA,
				CuDNN:  cuVer.CuDNN,
				Ubuntu: cuVer.Ubuntu,
			}] = image
			if !builder.ImageExists(image.RepositoryName()) {
				go buildImageWithoutPythonPackages(builder, &wg, cudaImage, py, image.RepositoryName())
				wg.Add(1)
				time.Sleep(waitTime)
			}
		}
	}

	// build cpu base images without framework
	for _, py := range PythonVersions {
		for ubuntu, ubuntuImage := range UbuntuImages {
			image := BaseImage{
				Registry:      registry,
				Project:       project,
				Version:       version,
				Python:        py,
				Ubuntu:        ubuntu,
				FrameworkMeta: nil,
			}
			cpuBaseImages[PythonUbuntu{py, ubuntu}] = image
			if !builder.ImageExists(image.RepositoryName()) {
				go buildImageWithoutPythonPackages(builder, &wg, ubuntuImage, py, image.RepositoryName())
				wg.Add(1)
				time.Sleep(waitTime)
			}
		}
	}

	// wait for no-framework base images to finish building before
	// we build base images with frameworks
	wg.Wait()
	wg = sync.WaitGroup{}

	metas := []FrameworkMeta{}
	for _, meta := range TensorflowMetas {
		metas = append(metas, meta)
	}
	for _, meta := range PyTorchMetas {
		metas = append(metas, meta)
	}

	// build images with frameworks (torch/tf) on top of the
	// images we built above
	for _, meta := range metas {
		for _, py := range meta.PythonVersions() {
			cuda := meta.GetCUDA()
			cuDNN := meta.GetCuDNN()
			ubuntu := LatestUbuntuForCUDA[cuda]

			// first, build gpu image
			pythonCUDACuDNNUbuntu := PythonCUDACuDNNUbuntu{
				Python: py,
				CUDA:   cuda,
				CuDNN:  cuDNN,
				Ubuntu: ubuntu,
			}
			gpuBaseImage, ok := gpuBaseImages[pythonCUDACuDNNUbuntu]
			if !ok {
				return fmt.Errorf("No base image for %v", pythonCUDACuDNNUbuntu)
			}

			gpuImage := BaseImage{
				Registry:      registry,
				Project:       project,
				Version:       version,
				Python:        py,
				Ubuntu:        ubuntu,
				FrameworkMeta: meta,
				CUDA:          cuda,
				CuDNN:         cuDNN,
			}
			if !builder.ImageExists(gpuImage.RepositoryName()) {
				go buildImageWithPythonPackages(builder, &wg, gpuBaseImage.RepositoryName(), meta.GPUPackages(), py, gpuImage.RepositoryName())
				wg.Add(1)
				time.Sleep(waitTime)
			}

			// then, build cpu image
			pythonUbuntu := PythonUbuntu{py, ubuntu}
			cpuBaseImage, ok := cpuBaseImages[pythonUbuntu]
			if !ok {
				return fmt.Errorf("No base image for %v", pythonUbuntu)
			}
			cpuImage := BaseImage{
				Registry:      registry,
				Project:       project,
				Version:       version,
				Python:        py,
				Ubuntu:        ubuntu,
				FrameworkMeta: meta,
			}
			if !builder.ImageExists(cpuImage.RepositoryName()) {
				go buildImageWithPythonPackages(builder, &wg, cpuBaseImage.RepositoryName(), meta.CPUPackages(), py, cpuImage.RepositoryName())
				wg.Add(1)
				time.Sleep(waitTime)
			}
		}
	}
	wg.Wait()

	return nil
}

func buildImageWithoutPythonPackages(builder Builder, wg *sync.WaitGroup, osImage string, python Python, image string) {
	if builder.Verbose() {
		console.Info("Building image: %s", image)
	}

	params := DockerfileParams{
		BaseImage:     osImage,
		PythonVersion: python,
	}
	if err := buildImage(builder, "baseimages-base.Dockerfile", params, image); err != nil {
		console.Warn(err.Error())
	}
	wg.Done()
}

func buildImageWithPythonPackages(builder Builder, wg *sync.WaitGroup, baseImage string, packages []string, py Python, image string) {
	if builder.Verbose() {
		console.Info("Building image: %s", image)
	}
	params := DockerfileParams{
		BaseImage:      baseImage,
		PythonPackages: strings.Join(packages, " "),
	}
	if err := buildImage(builder, "baseimages-packages.Dockerfile", params, image); err != nil {
		console.Warn(err.Error())
	}
	wg.Done()
}

func buildImage(builder Builder, assetName string, params DockerfileParams, image string) error {
	tmpDir, err := files.TempDir("base-image-builder")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	// need the "source" part for archival reasons
	tmpDir = path.Join(tmpDir, "source")
	if err := os.Mkdir(tmpDir, 0755); err != nil {
		return err
	}

	contents := assets.MustAsset(assetName)
	tmpl, err := template.New("Dockerfile").Parse(string(contents))
	if err != nil {
		return fmt.Errorf("Failed to parse template: %s", err)
	}

	f, err := os.Create(path.Join(tmpDir, "Dockerfile"))
	if err != nil {
		return fmt.Errorf("Failed create Dockerfile: %s", err)
	}

	err = tmpl.Execute(f, params)
	if err != nil {
		return fmt.Errorf("Failed execute template: %s", err)
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("Failed to close file: %s", err)
	}

	c, err := ioutil.ReadFile(path.Join(tmpDir, "Dockerfile"))
	if err != nil {
		return err
	}
	s := string(c)
	if builder.Verbose() {
		fmt.Printf("%-90s (%s)\n%s\n", image, params.BaseImage, s)
	}

	return builder.Build(tmpDir, string(image))
}
