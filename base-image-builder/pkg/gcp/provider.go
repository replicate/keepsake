package gcp

import (
	"fmt"
	"context"

	"cloud.google.com/go/logging/logadmin"
	monitoring "cloud.google.com/go/monitoring/apiv3"
	"cloud.google.com/go/storage"
	"google.golang.org/api/cloudbuild/v1"
)

type Provider struct {
	build        *cloudbuild.Service
	storage      *storage.Client
	logadmin     *logadmin.Client
	monitoring   *monitoring.MetricClient
	projectID    string
	zone         string
	network      string
	subnet       string
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
