package gcp

import (
	"context"
	"testing"
	"time"

	billingpb "cloud.google.com/go/billing/apiv1/billingpb"
	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"
	money "google.golang.org/genproto/googleapis/type/money"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"mcp-server/internal/database"
)

type fakeCatalog struct {
	services []*billingpb.Service
	skus     []*billingpb.Sku
}

func (f fakeCatalog) ListServices(ctx context.Context, req *billingpb.ListServicesRequest, opts ...gax.CallOption) ServiceIterator {
	return &fakeServiceIterator{services: f.services}
}

func (f fakeCatalog) ListSkus(ctx context.Context, req *billingpb.ListSkusRequest, opts ...gax.CallOption) SkuIterator {
	return &fakeSkuIterator{skus: f.skus}
}

func (f fakeCatalog) Close() error { return nil }

type fakeServiceIterator struct {
	services []*billingpb.Service
	idx      int
}

func (it *fakeServiceIterator) Next() (*billingpb.Service, error) {
	if it.idx >= len(it.services) {
		return nil, iterator.Done
	}
	svc := it.services[it.idx]
	it.idx++
	return svc, nil
}

type fakeSkuIterator struct {
	skus []*billingpb.Sku
	idx  int
}

func (it *fakeSkuIterator) Next() (*billingpb.Sku, error) {
	if it.idx >= len(it.skus) {
		return nil, iterator.Done
	}
	sku := it.skus[it.idx]
	it.idx++
	return sku, nil
}

func TestClient_ListServices(t *testing.T) {
	svc := &billingpb.Service{
		Name:               "services/abc",
		DisplayName:        "ABC",
		BusinessEntityName: "Google",
	}
	c := Client{catalog: fakeCatalog{services: []*billingpb.Service{svc}}}
	got, err := c.ListServices(context.Background())
	if err != nil {
		t.Fatalf("ListServices: %v", err)
	}
	want := []database.Service{{ServiceID: "abc", DisplayName: "ABC", BusinessEntityName: "Google"}}
	if len(got) != len(want) {
		t.Fatalf("got %d services, want %d", len(got), len(want))
	}
	if got[0] != want[0] {
		t.Fatalf("service mismatch: got %+v want %+v", got[0], want[0])
	}
}

func TestClient_ListSkus(t *testing.T) {
	sku := &billingpb.Sku{
		Name:        "services/svc/skus/sku1",
		SkuId:       "sku1",
		Description: "desc",
		Category: &billingpb.Category{
			ServiceDisplayName: "Svc",
			ResourceFamily:     "fam",
			ResourceGroup:      "group",
			UsageType:          "usage",
		},
		ServiceRegions: []string{"us"},
		GeoTaxonomy: &billingpb.GeoTaxonomy{
			Type:    billingpb.GeoTaxonomy_REGIONAL,
			Regions: []string{"us-east1"},
		},
		PricingInfo: []*billingpb.PricingInfo{{
			EffectiveTime: timestamppb.New(time.Unix(0, 0)),
			Summary:       "sum",
			PricingExpression: &billingpb.PricingExpression{
				UsageUnit:            "h",
				UsageUnitDescription: "hour",
				DisplayQuantity:      1,
				TieredRates: []*billingpb.PricingExpression_TierRate{{
					StartUsageAmount: 0,
					UnitPrice:        &money.Money{CurrencyCode: "USD", Units: 1, Nanos: 500000000},
				}},
			},
		}},
	}
	c := Client{catalog: fakeCatalog{skus: []*billingpb.Sku{sku}}}
	gotSkus, gotPrices, err := c.ListSkus(context.Background(), "svc")
	if err != nil {
		t.Fatalf("ListSkus: %v", err)
	}
	if len(gotSkus) != 1 {
		t.Fatalf("got %d skus, want 1", len(gotSkus))
	}
	gs := gotSkus[0]
	if gs.SKUID != "sku1" || gs.ServiceID != "svc" || gs.SkuName != "sku1" {
		t.Fatalf("sku fields not mapped: %+v", gs)
	}
	if len(gotPrices) != 1 {
		t.Fatalf("got %d prices, want 1", len(gotPrices))
	}
	gp := gotPrices[0]
	if gp.SKUID != "sku1" || gp.CurrencyCode != "USD" || gp.UsageUnit != "h" {
		t.Fatalf("pricing info not mapped: %+v", gp)
	}
	if len(gp.TieredRates) != 1 {
		t.Fatalf("got %d tiered rates, want 1", len(gp.TieredRates))
	}
	tr := gp.TieredRates[0]
	if tr.StartUsageAmount != 0 || tr.UnitPrice.CurrencyCode != "USD" || tr.UnitPrice.Units != 1 || tr.UnitPrice.Nanos != 500000000 {
		t.Fatalf("tiered rate mismatch: %+v", tr)
	}
}
