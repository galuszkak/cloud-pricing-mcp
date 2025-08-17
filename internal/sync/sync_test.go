package sync

import (
	"context"
	"testing"

	"mcp-server/internal/database"
)

type fakeClient struct{}

func (f fakeClient) ListServices(ctx context.Context) ([]database.Service, error) {
	return []database.Service{{ServiceID: "svc", DisplayName: "Svc", BusinessEntityName: "Ent"}}, nil
}

func (f fakeClient) ListSkus(ctx context.Context, serviceID string) ([]database.SKU, []database.PricingInfo, error) {
	sku := database.SKU{
		SKUID:       "sku1",
		ServiceID:   serviceID,
		SkuName:     "SKU",
		Description: "desc",
		Category:    database.Category{ServiceDisplayName: "Svc"},
	}
	price := database.PricingInfo{SKUID: "sku1", CurrencyCode: "USD", UsageUnit: "h", TieredRates: []database.TieredRate{}}
	return []database.SKU{sku}, []database.PricingInfo{price}, nil
}

func TestJob_Run(t *testing.T) {
	cleanDB(t)
	job := NewJob(fakeClient{}, testRepo)
	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("run: %v", err)
	}
	assertCount := func(table string, want int) {
		var count int
		if err := testDB.QueryRowContext(context.Background(), "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
			t.Fatalf("count %s: %v", table, err)
		}
		if count != want {
			t.Fatalf("%s count = %d, want %d", table, count, want)
		}
	}
	assertCount("services", 1)
	assertCount("skus", 1)
	assertCount("pricing_updates", 1)
	assertCount("pricing_info", 1)
}
