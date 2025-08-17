package database

import (
	"context"
	"testing"
	"time"
)

func TestSQLRepository_UpsertService(t *testing.T) {
	repo := setupTestRepo(t)
	svc := Service{ServiceID: "service1", DisplayName: "Service One", BusinessEntityName: "Entity"}
	ctx := context.Background()
	if err := repo.UpsertService(ctx, svc); err != nil {
		t.Fatalf("insert: %v", err)
	}
	// update
	svc.DisplayName = "Service 1"
	if err := repo.UpsertService(ctx, svc); err != nil {
		t.Fatalf("update: %v", err)
	}
}

func TestSQLRepository_UpsertSKU(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()
	// insert service first for FK
	if err := repo.UpsertService(ctx, Service{ServiceID: "svc", DisplayName: "Svc", BusinessEntityName: "Ent"}); err != nil {
		t.Fatalf("service: %v", err)
	}
	sku := SKU{
		SKUID:       "sku1",
		ServiceID:   "svc",
		SkuName:     "SKU One",
		Description: "desc",
		Category: Category{
			ServiceDisplayName: "Svc",
			ResourceFamily:     "compute",
			ResourceGroup:      "cpu",
			UsageType:          "onDemand",
		},
		ServiceRegions: []string{"us-east1"},
		GeoTaxonomy:    GeoTaxonomy{Type: "REGIONAL", Regions: []string{"us-east1"}},
	}
	if err := repo.UpsertSKU(ctx, sku); err != nil {
		t.Fatalf("insert: %v", err)
	}
	sku.Description = "desc2"
	if err := repo.UpsertSKU(ctx, sku); err != nil {
		t.Fatalf("update: %v", err)
	}
}

func TestSQLRepository_UpsertPricingInfo(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()
	if err := repo.UpsertService(ctx, Service{ServiceID: "svc", DisplayName: "Svc", BusinessEntityName: "Ent"}); err != nil {
		t.Fatalf("service: %v", err)
	}
	sku := SKU{SKUID: "sku1", ServiceID: "svc", SkuName: "SKU One", Description: "desc"}
	if err := repo.UpsertSKU(ctx, sku); err != nil {
		t.Fatalf("sku: %v", err)
	}
	pi := PricingInfo{
		SKUID:           "sku1",
		EffectiveTime:   time.Unix(0, 0),
		Summary:         "test",
		CurrencyCode:    "USD",
		UsageUnit:       "h",
		DisplayQuantity: 1,
		TieredRates: []TieredRate{{
			StartUsageAmount: 0,
			UnitPrice:        Money{CurrencyCode: "USD", Units: 0, Nanos: 1000000},
		}},
	}
	if err := repo.UpsertPricingInfo(ctx, pi); err != nil {
		t.Fatalf("insert: %v", err)
	}
	pi.Summary = "updated"
	if err := repo.UpsertPricingInfo(ctx, pi); err != nil {
		t.Fatalf("update: %v", err)
	}
}
