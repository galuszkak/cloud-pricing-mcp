package gcp

import (
	"context"
	"strings"

	billing "cloud.google.com/go/billing/apiv1"
	billingpb "cloud.google.com/go/billing/apiv1/billingpb"
	gax "github.com/googleapis/gax-go/v2"
	"google.golang.org/api/iterator"

	"mcp-server/internal/database"
)

// ServiceIterator iterates over catalog services.
type ServiceIterator interface {
	Next() (*billingpb.Service, error)
}

// SkuIterator iterates over catalog SKUs.
type SkuIterator interface {
	Next() (*billingpb.Sku, error)
}

// CloudCatalogClient defines the subset of the Cloud Catalog client used by this package.
type CloudCatalogClient interface {
	ListServices(ctx context.Context, req *billingpb.ListServicesRequest, opts ...gax.CallOption) ServiceIterator
	ListSkus(ctx context.Context, req *billingpb.ListSkusRequest, opts ...gax.CallOption) SkuIterator
	Close() error
}

// Client wraps the Cloud Catalog API client.
type Client struct {
	catalog CloudCatalogClient
}

// NewClient creates a new Client.
func NewClient(ctx context.Context) (*Client, error) {
	c, err := billing.NewCloudCatalogClient(ctx)
	if err != nil {
		return nil, err
	}
	return &Client{catalog: cloudCatalogClient{c}}, nil
}

type cloudCatalogClient struct {
	*billing.CloudCatalogClient
}

func (c cloudCatalogClient) ListServices(ctx context.Context, req *billingpb.ListServicesRequest, opts ...gax.CallOption) ServiceIterator {
	return c.CloudCatalogClient.ListServices(ctx, req, opts...)
}

func (c cloudCatalogClient) ListSkus(ctx context.Context, req *billingpb.ListSkusRequest, opts ...gax.CallOption) SkuIterator {
	return c.CloudCatalogClient.ListSkus(ctx, req, opts...)
}

// Close releases resources held by the underlying catalog client.
func (c *Client) Close() error {
	return c.catalog.Close()
}

// ListServices retrieves all services from the Catalog API.
func (c *Client) ListServices(ctx context.Context) ([]database.Service, error) {
	it := c.catalog.ListServices(ctx, &billingpb.ListServicesRequest{})
	var services []database.Service
	for {
		svc, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		id := strings.TrimPrefix(svc.GetName(), "services/")
		services = append(services, database.Service{
			ServiceID:          id,
			DisplayName:        svc.GetDisplayName(),
			BusinessEntityName: svc.GetBusinessEntityName(),
		})
	}
	return services, nil
}

// ListSkus retrieves all SKUs and pricing info for a service.
func (c *Client) ListSkus(ctx context.Context, serviceID string) ([]database.SKU, []database.PricingInfo, error) {
	it := c.catalog.ListSkus(ctx, &billingpb.ListSkusRequest{Parent: "services/" + serviceID})
	var skus []database.SKU
	var prices []database.PricingInfo
	for {
		sku, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, nil, err
		}
		cat := database.Category{
			ServiceDisplayName: sku.GetCategory().GetServiceDisplayName(),
			ResourceFamily:     sku.GetCategory().GetResourceFamily(),
			ResourceGroup:      sku.GetCategory().GetResourceGroup(),
			UsageType:          sku.GetCategory().GetUsageType(),
		}
		geo := database.GeoTaxonomy{
			Type:    sku.GetGeoTaxonomy().GetType().String(),
			Regions: sku.GetGeoTaxonomy().GetRegions(),
		}
		id := strings.TrimPrefix(sku.GetName(), "services/"+serviceID+"/skus/")
		skus = append(skus, database.SKU{
			SKUID:          id,
			ServiceID:      serviceID,
			SkuName:        sku.GetSkuId(),
			Description:    sku.GetDescription(),
			Category:       cat,
			ServiceRegions: sku.GetServiceRegions(),
			GeoTaxonomy:    geo,
		})
		for _, pi := range sku.GetPricingInfo() {
			var rates []database.TieredRate
			for _, r := range pi.GetPricingExpression().GetTieredRates() {
				rates = append(rates, database.TieredRate{
					StartUsageAmount: r.GetStartUsageAmount(),
					UnitPrice: database.Money{
						CurrencyCode: r.GetUnitPrice().GetCurrencyCode(),
						Units:        r.GetUnitPrice().GetUnits(),
						Nanos:        r.GetUnitPrice().GetNanos(),
					},
				})
			}
			currency := ""
			if len(pi.GetPricingExpression().GetTieredRates()) > 0 {
				currency = pi.GetPricingExpression().GetTieredRates()[0].GetUnitPrice().GetCurrencyCode()
			}
			prices = append(prices, database.PricingInfo{
				SKUID:                id,
				EffectiveTime:        pi.GetEffectiveTime().AsTime(),
				Summary:              pi.GetSummary(),
				CurrencyCode:         currency,
				UsageUnit:            pi.GetPricingExpression().GetUsageUnit(),
				UsageUnitDescription: pi.GetPricingExpression().GetUsageUnitDescription(),
				DisplayQuantity:      int64(pi.GetPricingExpression().GetDisplayQuantity()),
				TieredRates:          rates,
			})
		}
	}
	return skus, prices, nil
}
