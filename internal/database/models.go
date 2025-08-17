package database

import "time"

// Service represents a cloud service.
type Service struct {
	ServiceID          string
	DisplayName        string
	BusinessEntityName string
}

// Category describes a SKU category.
type Category struct {
	ServiceDisplayName string `json:"serviceDisplayName"`
	ResourceFamily     string `json:"resourceFamily"`
	ResourceGroup      string `json:"resourceGroup"`
	UsageType          string `json:"usageType"`
}

// GeoTaxonomy specifies the geographic taxonomy for a SKU.
type GeoTaxonomy struct {
	Type    string   `json:"type"`
	Regions []string `json:"regions"`
}

// SKU represents a stock-keeping unit for a service.
type SKU struct {
	SKUID          string
	ServiceID      string
	SkuName        string
	Description    string
	Category       Category
	ServiceRegions []string
	GeoTaxonomy    GeoTaxonomy
}

// Money represents a currency amount.
type Money struct {
	CurrencyCode string `json:"currencyCode"`
	Units        int64  `json:"units"`
	Nanos        int32  `json:"nanos"`
}

// TieredRate defines a usage-based price tier.
type TieredRate struct {
	StartUsageAmount float64 `json:"startUsageAmount"`
	UnitPrice        Money   `json:"unitPrice"`
}

// PricingInfo holds pricing details for a SKU.
type PricingInfo struct {
	PricingInfoID        int64
	SKUID                string
	EffectiveTime        time.Time
	Summary              string
	CurrencyCode         string
	UsageUnit            string
	UsageUnitDescription string
	DisplayQuantity      int64
	TieredRates          []TieredRate
}

// PricingUpdate records a synchronization run.
type PricingUpdate struct {
	UpdateID        int64
	UpdateTime      time.Time
	Status          string
	ServicesUpdated int
	SkusUpdated     int
	LogMessage      string
}
