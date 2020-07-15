package main

import (
	"github.com/spf13/cobra"
	"replicate.ai/cli/pkg/baseimages"

	"replicate.ai/image-builder/pkg/gcp"
	"time"
)

func main() {
	cmd := newRootCommand()
	cmd.Execute()
}

func newRootCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:    "image-builder",
		Short:  "Build Replicate base images.",
		RunE:   buildBaseImages,
		Args:   cobra.ExactArgs(0),
		Hidden: true,
	}

	cmd.Flags().String("version", "", "Base image version")
	cmd.Flags().String("project-id", "replicate", "GCP project ID")
	cmd.Flags().String("registry", "us.gcr.io", "Docker registry")
	cmd.Flags().IntP("wait-seconds", "w", 20, "Wait time between starting builds")
	cmd.MarkFlagRequired("version")

	return cmd
}

func buildBaseImages(cmd *cobra.Command, args []string) error {
	projectID, err := cmd.Flags().GetString("project-id")
	if err != nil {
		return err
	}

	cloudProvider, err := gcp.NewProvider(projectID)
	if err != nil {
		return err
	}
	registry, err := cmd.Flags().GetString("registry")
	if err != nil {
		return err
	}
	version, err := cmd.Flags().GetString("version")
	if err != nil {
		return err
	}
	waitSeconds, err := cmd.Flags().GetInt("wait-seconds")
	if err != nil {
		return err
	}
	waitTime := time.Duration(waitSeconds) * time.Second

	return baseimages.BuildBaseImages(cloudProvider, projectID, registry, version, waitTime)
}
