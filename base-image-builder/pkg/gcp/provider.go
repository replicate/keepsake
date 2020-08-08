package gcp

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"google.golang.org/api/cloudbuild/v1"
)

type Provider struct {
	build     *cloudbuild.Service
	storage   *storage.Client
	projectID string
}

// TODO: automatically run gcloud services enable container.googleapis.com

func NewProvider(projectID string) (p *Provider, err error) {
	p = &Provider{
		projectID: projectID,
	}
	p.build, err = cloudbuild.NewService(context.TODO())
	if err != nil {
		return nil, err
	}
	p.storage, err = storage.NewClient(context.TODO())
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *Provider) GetTempBucket() string {
	return fmt.Sprintf("gs://%s-replicate-temp", p.projectID)
}
