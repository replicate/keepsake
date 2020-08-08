package gcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/apex/log"
	"github.com/dchest/uniuri"
	"github.com/mholt/archiver"
	"google.golang.org/api/cloudbuild/v1"
)

const topLevelSourceDir = "source"

var tagRe = regexp.MustCompile("^([^:]+):([^-]+)(-.+)?$")

func (p *Provider) ImageExists(image string) bool {
	// TODO(andreas): error handling
	if exec.Command("gcloud", "container", "images", "describe", string(image)).Run() == nil {
		log.Infof("Image already exists %s", image)
		return true
	}
	return false
}

func (p *Provider) Verbose() bool {
	return true
}

func (p *Provider) Build(dir string, image string) error {
	log.Debugf("Uploading source dir %s for %s to GCS...", dir, image)
	obj, err := p.uploadSource(dir)
	if err != nil {
		return err
	}

	// TODO: clean up builds at some point, but doing it here can lead to builds failing in case of rate limits.
	//log.Infof("Deleting temporary source for %s", image)
	//defer obj.Delete(context.TODO())

	log.Debugf("Submitting %s to CloudBuild...", image)
	err = p.cloudBuild(image, obj, log.Debug)
	if err != nil {
		return err
	}
	return nil
}

func (p *Provider) cloudBuild(dockerTag string, obj *storage.ObjectHandle, logger func(msg string)) error {
	dockerTagLatest := latestTag(dockerTag)

	build := makeBuild(obj.BucketName(), obj.ObjectName(), dockerTag, dockerTagLatest)
	op, err := p.build.Projects.Builds.Create(p.projectID, build).Do()
	if err != nil {
		return fmt.Errorf("Build failed: %s", err)
	}
	metadata, err := getBuildMetadata(op)
	if err != nil {
		return fmt.Errorf("Failed to get build metadata: %s", err)
	}
	buildID := metadata.Build.Id
	logURL := metadata.Build.LogUrl
	log.Infof("Submitted to Cloud Build, log url: %s", logURL) // TODO: hide in saas mode

	err = waitForOperation(context.TODO(), func() (bool, error) {
		build, err := p.build.Projects.Builds.Get(p.projectID, buildID).Do()
		if err != nil {
			return false, err
		}
		if s := build.Status; s != "WORKING" && s != "QUEUED" {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("Failed to get operation: %s", err)
	}
	build, err = p.build.Projects.Builds.Get(p.projectID, buildID).Do()
	if err != nil {
		return fmt.Errorf("Failed to get build: %s", err)
	}
	if build.Status == "FAILURE" {
		return fmt.Errorf("Build failed: %s (%s)", dockerTag, logURL)
	}
	log.Infof("Successfully pushed to %s and %s", dockerTag, dockerTagLatest)
	return nil
}

func makeBuild(gcsBucket string, gcsPath string, dockerTag string, dockerTagLatest string) *cloudbuild.Build {
	return &cloudbuild.Build{
		Source: &cloudbuild.Source{
			StorageSource: &cloudbuild.StorageSource{
				Bucket: gcsBucket,
				Object: gcsPath,
			},
		},
		Images: []string{dockerTag, dockerTagLatest},
		Steps: []*cloudbuild.BuildStep{
			{
				Name:       "gcr.io/cloud-builders/docker",
				Entrypoint: "bash",
				Args: []string{
					"-c", fmt.Sprintf("docker pull %s || exit 0", dockerTagLatest),
				},
			},
			{
				Name: "gcr.io/cloud-builders/docker",
				Args: []string{
					"build",
					"-t", dockerTag,
					"--cache-from", dockerTagLatest,
					".",
				},
				Dir: topLevelSourceDir,
			},
			{
				Name: "gcr.io/cloud-builders/docker",
				Args: []string{
					"build",
					"-t", dockerTagLatest,
					"--cache-from", dockerTag,
					".",
				},
				Dir: topLevelSourceDir,
			},
		},
		Timeout: "1200s",
	}
}

func (p *Provider) uploadSource(sourceDir string) (obj *storage.ObjectHandle, err error) {
	tarPath, tarDir, err := compressSource(sourceDir)
	defer os.RemoveAll(tarDir)

	bucketName := strings.TrimPrefix(p.GetTempBucket(), "gs://")
	bucket := p.storage.Bucket(bucketName)
	obj = bucket.Object(path.Join("build", uniuri.NewLen(50), filepath.Base(tarPath)))
	w := obj.NewWriter(context.TODO())
	r, err := os.Open(tarPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read compressed model from %s, got error: %s", tarPath, err)
	}
	if _, err := io.Copy(w, r); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}

	return obj, nil
}

func latestTag(dockerTag string) string {
	return tagRe.ReplaceAllString(dockerTag, "$1:latest")
}

func getBuildMetadata(op *cloudbuild.Operation) (*cloudbuild.BuildOperationMetadata, error) {
	if op.Metadata == nil {
		return nil, fmt.Errorf("missing Metadata in operation")
	}
	var metadata cloudbuild.BuildOperationMetadata
	if err := json.Unmarshal([]byte(op.Metadata), &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

func compressSource(sourceDir string) (archivePath string, tempDir string, err error) {
	archivePath = path.Join(sourceDir, "source.tar.gz")
	if err != nil {
		return "", tempDir, err
	}
	err = archiver.Archive([]string{sourceDir}, archivePath)
	if err != nil {
		return "", tempDir, err
	}

	return archivePath, tempDir, nil
}
