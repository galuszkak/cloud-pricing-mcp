package sync

import (
	"context"
	"time"

	"mcp-server/internal/database"
)

// CatalogClient defines the interface for retrieving GCP catalog data.
type CatalogClient interface {
	ListServices(ctx context.Context) ([]database.Service, error)
	ListSkus(ctx context.Context, serviceID string) ([]database.SKU, []database.PricingInfo, error)
}

// Job synchronizes pricing data from GCP into the local database.
type Job struct {
	client CatalogClient
	repo   database.Repository
}

// NewJob creates a new Job.
func NewJob(client CatalogClient, repo database.Repository) *Job {
	return &Job{client: client, repo: repo}
}

// Run executes the synchronization process.
func (j *Job) Run(ctx context.Context) error {
	services, err := j.client.ListServices(ctx)
	if err != nil {
		return err
	}
	var servicesUpdated, skusUpdated int
	for _, svc := range services {
		if err := j.repo.UpsertService(ctx, svc); err != nil {
			return err
		}
		servicesUpdated++
		skus, prices, err := j.client.ListSkus(ctx, svc.ServiceID)
		if err != nil {
			return err
		}
		for _, sku := range skus {
			if err := j.repo.UpsertSKU(ctx, sku); err != nil {
				return err
			}
			skusUpdated++
		}
		for _, p := range prices {
			if err := j.repo.UpsertPricingInfo(ctx, p); err != nil {
				return err
			}
		}
	}
	update := database.PricingUpdate{
		UpdateTime:      time.Now().UTC(),
		Status:          "SUCCESS",
		ServicesUpdated: servicesUpdated,
		SkusUpdated:     skusUpdated,
		LogMessage:      "sync completed",
	}
	if err := j.repo.InsertPricingUpdate(ctx, update); err != nil {
		return err
	}
	return nil
}
