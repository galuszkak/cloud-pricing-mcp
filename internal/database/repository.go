package database

import (
	"context"
	"database/sql"
	"encoding/json"
)

// Repository defines database operations for pricing data.
type Repository interface {
	UpsertService(ctx context.Context, s Service) error
	UpsertSKU(ctx context.Context, s SKU) error
	UpsertPricingInfo(ctx context.Context, p PricingInfo) error
	InsertPricingUpdate(ctx context.Context, u PricingUpdate) error
}

// SQLRepository implements Repository using an SQL database.
type SQLRepository struct {
	db *sql.DB
}

// NewRepository creates a new SQLRepository.
func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

// UpsertService inserts or updates a service.
func (r *SQLRepository) UpsertService(ctx context.Context, s Service) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO services (service_id, display_name, business_entity_name)
VALUES (?, ?, ?) ON CONFLICT(service_id) DO UPDATE SET display_name=excluded.display_name, business_entity_name=excluded.business_entity_name`, s.ServiceID, s.DisplayName, s.BusinessEntityName)
	return err
}

// UpsertSKU inserts or updates a SKU.
func (r *SQLRepository) UpsertSKU(ctx context.Context, s SKU) error {
	cat, err := json.Marshal(s.Category)
	if err != nil {
		return err
	}
	regions, err := json.Marshal(s.ServiceRegions)
	if err != nil {
		return err
	}
	geo, err := json.Marshal(s.GeoTaxonomy)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO skus (sku_id, service_id, sku_name, description, category, service_regions, geo_taxonomy)
VALUES (?, ?, ?, ?, ?, ?, ?) ON CONFLICT(sku_id) DO UPDATE SET service_id=excluded.service_id, sku_name=excluded.sku_name, description=excluded.description, category=excluded.category, service_regions=excluded.service_regions, geo_taxonomy=excluded.geo_taxonomy`, s.SKUID, s.ServiceID, s.SkuName, s.Description, cat, regions, geo)
	return err
}

// UpsertPricingInfo inserts or updates pricing info for a SKU.
func (r *SQLRepository) UpsertPricingInfo(ctx context.Context, p PricingInfo) error {
	rates, err := json.Marshal(p.TieredRates)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `INSERT INTO pricing_info (sku_id, effective_time, summary, currency_code, usage_unit, usage_unit_description, display_quantity, tiered_rates)
VALUES (?, ?, ?, ?, ?, ?, ?, ?) ON CONFLICT(sku_id, effective_time) DO UPDATE SET summary=excluded.summary, currency_code=excluded.currency_code, usage_unit=excluded.usage_unit, usage_unit_description=excluded.usage_unit_description, display_quantity=excluded.display_quantity, tiered_rates=excluded.tiered_rates`, p.SKUID, p.EffectiveTime, p.Summary, p.CurrencyCode, p.UsageUnit, p.UsageUnitDescription, p.DisplayQuantity, rates)
	return err
}

// InsertPricingUpdate records a sync run.
func (r *SQLRepository) InsertPricingUpdate(ctx context.Context, u PricingUpdate) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO pricing_updates (update_time, status, services_updated, skus_updated, log_message)
VALUES (?, ?, ?, ?, ?)`, u.UpdateTime, u.Status, u.ServicesUpdated, u.SkusUpdated, u.LogMessage)
	return err
}

var _ Repository = (*SQLRepository)(nil)
