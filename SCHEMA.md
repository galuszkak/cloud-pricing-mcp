# Database Schema Documentation

This document provides a detailed overview of the database schema used by the Cloud Pricing MCP server. The schema is designed to store Google Cloud pricing information retrieved from the Cloud Billing Catalog API and optimize it for fast querying by AI agents.

## Schema Overview

The database uses **Turso (libSQL)** as the storage layer and is designed to mirror the structure of Google Cloud's Billing Catalog API while optimizing for query performance. The schema consists of several core tables that maintain pricing data, service metadata, and synchronization audit trails.

## Core Tables

### 1. `services` Table

Stores Google Cloud service metadata.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `service_id` | TEXT | PRIMARY KEY | Unique identifier from Google Cloud API (e.g., "6F81-5844-456A") |
| `display_name` | TEXT | NOT NULL | Human-readable service name (e.g., "Compute Engine") |
| `description` | TEXT | | Detailed description of the service |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Record creation timestamp |
| `updated_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Last update timestamp |

**Purpose**: Provides service-level metadata for organizing and searching SKUs.

**Example Data**:
```
service_id: "6F81-5844-456A"
display_name: "Compute Engine"
description: "Google Compute Engine provides virtual machines running in Google's data centers"
```

### 2. `skus` Table

Stores individual Stock Keeping Units (SKUs) with their metadata and categorization.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `sku_id` | TEXT | PRIMARY KEY | Unique SKU identifier from Google Cloud API |
| `service_id` | TEXT | FOREIGN KEY → services(service_id) | Reference to parent service |
| `display_name` | TEXT | NOT NULL | Human-readable SKU name |
| `description` | TEXT | | Detailed SKU description |
| `category_json` | TEXT | | JSON object containing category information |
| `resource_family` | TEXT | | Extracted resource family (e.g., "Compute") |
| `resource_group` | TEXT | | Extracted resource group (e.g., "N1") |
| `usage_type` | TEXT | | Extracted usage type (e.g., "OnDemand") |
| `geo_taxonomy_type` | TEXT | | Geographic taxonomy type |
| `geo_taxonomy_regions` | TEXT | | JSON array of applicable regions |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Record creation timestamp |
| `updated_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Last update timestamp |

**Purpose**: Core catalog of billable items with rich metadata for search and filtering.

**Category JSON Structure**:
```json
{
  "serviceDisplayName": "Compute Engine",
  "resourceFamily": "Compute",
  "resourceGroup": "N1",
  "usageType": "OnDemand"
}
```

### 3. `pricing_info` Table

Stores pricing information snapshots for each SKU over time.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | INTEGER | PRIMARY KEY AUTOINCREMENT | Unique pricing record identifier |
| `sku_id` | TEXT | FOREIGN KEY → skus(sku_id) | Reference to SKU |
| `effective_time` | DATETIME | | When this pricing becomes effective |
| `summary` | TEXT | | Human-readable pricing summary |
| `currency_code` | TEXT | DEFAULT 'USD' | Currency (limited to USD initially) |
| `pricing_expression_json` | TEXT | | JSON object with complete pricing expression |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Record creation timestamp |

**Purpose**: Time-based pricing records allowing for historical pricing queries and future price changes.

**Pricing Expression JSON Structure**:
```json
{
  "usageUnit": "hour",
  "usageUnitDescription": "hour",
  "baseUnit": "hour",
  "baseUnitDescription": "hour",
  "baseUnitConversionFactor": 1,
  "displayQuantity": 1,
  "tieredRates": [
    {
      "startUsageAmount": 0,
      "unitPrice": {
        "units": "0",
        "nanos": 24000000
      }
    }
  ]
}
```

### 4. `pricing_tiers` Table

Stores detailed tiered pricing rates extracted from pricing expressions.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | INTEGER | PRIMARY KEY AUTOINCREMENT | Unique tier identifier |
| `pricing_info_id` | INTEGER | FOREIGN KEY → pricing_info(id) | Reference to pricing info |
| `start_usage_amount` | REAL | NOT NULL | Usage threshold for this tier |
| `price_units` | TEXT | NOT NULL | Whole number part of price |
| `price_nanos` | INTEGER | NOT NULL | Fractional part of price (in billionths) |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Record creation timestamp |

**Purpose**: Enables efficient calculation of tiered pricing for usage-based billing.

**Example Data**:
```
start_usage_amount: 0.0
price_units: "0"
price_nanos: 24000000  // $0.024 per hour
```

### 5. `pricing_updates` Table

Audit log for synchronization operations.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | INTEGER | PRIMARY KEY AUTOINCREMENT | Unique update identifier |
| `started_at` | DATETIME | NOT NULL | Sync job start time |
| `completed_at` | DATETIME | | Sync job completion time (NULL if failed) |
| `status` | TEXT | NOT NULL | "running", "completed", "failed" |
| `services_synced` | INTEGER | DEFAULT 0 | Number of services processed |
| `skus_synced` | INTEGER | DEFAULT 0 | Number of SKUs processed |
| `pricing_records_synced` | INTEGER | DEFAULT 0 | Number of pricing records processed |
| `error_message` | TEXT | | Error details if sync failed |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Record creation timestamp |

**Purpose**: Provides audit trail and monitoring for data synchronization jobs.

## Indexes

To optimize query performance for the MCP tools, the following indexes are recommended:

```sql
-- Search optimization
CREATE INDEX idx_skus_display_name ON skus(display_name);
CREATE INDEX idx_skus_description ON skus(description);
CREATE INDEX idx_skus_resource_family ON skus(resource_family);
CREATE INDEX idx_skus_usage_type ON skus(usage_type);

-- Service relationships
CREATE INDEX idx_skus_service_id ON skus(service_id);
CREATE INDEX idx_pricing_info_sku_id ON pricing_info(sku_id);
CREATE INDEX idx_pricing_tiers_pricing_info_id ON pricing_tiers(pricing_info_id);

-- Time-based queries
CREATE INDEX idx_pricing_info_effective_time ON pricing_info(effective_time);
CREATE INDEX idx_pricing_updates_started_at ON pricing_updates(started_at);

-- Full-text search (if supported by Turso)
CREATE VIRTUAL TABLE skus_fts USING fts5(sku_id, display_name, description);
```

## Data Flow and Relationships

```
services (1) ←→ (many) skus
    ↓
skus (1) ←→ (many) pricing_info
    ↓
pricing_info (1) ←→ (many) pricing_tiers
```

## Key Design Decisions

### 1. Natural Primary Keys
- Uses Google Cloud API identifiers (`service_id`, `sku_id`) as primary keys
- Eliminates need for synthetic keys and simplifies synchronization logic
- Ensures referential integrity with external API

### 2. JSON Storage for Flexibility
- `category_json` stores complete category object for future extensibility
- `pricing_expression_json` preserves full API response structure
- Extracted fields (`resource_family`, `resource_group`, etc.) for efficient filtering

### 3. Price Precision
- Uses Google's `units` + `nanos` representation to avoid floating-point precision issues
- Stores prices as separate `price_units` (string) and `price_nanos` (integer) fields
- Enables accurate financial calculations

### 4. Time-based Pricing
- `pricing_info` table supports historical pricing and future price changes
- `effective_time` allows querying prices for specific dates
- Supports price change notifications and trend analysis

### 5. Audit and Monitoring
- `pricing_updates` table provides complete audit trail
- Enables monitoring of sync job health and performance
- Supports troubleshooting and data quality assurance

## Usage Patterns

### Search Tool Queries
```sql
-- Find SKUs by keyword
SELECT s.sku_id, s.display_name, s.description 
FROM skus s 
WHERE s.display_name LIKE '%compute%' 
   OR s.description LIKE '%compute%'
LIMIT 50;

-- Filter by resource family and region
SELECT s.sku_id, s.display_name, s.resource_family
FROM skus s
WHERE s.resource_family = 'Compute'
  AND s.geo_taxonomy_regions LIKE '%us-central1%';
```

### Details Tool Queries
```sql
-- Get complete SKU information with latest pricing
SELECT 
    s.*,
    pi.summary,
    pi.pricing_expression_json,
    pi.effective_time
FROM skus s
LEFT JOIN pricing_info pi ON s.sku_id = pi.sku_id
WHERE s.sku_id = 'ABCD-1234-5678'
  AND pi.effective_time <= CURRENT_TIMESTAMP
ORDER BY pi.effective_time DESC
LIMIT 1;
```

### Calculate Tool Queries
```sql
-- Get tiered pricing for calculations
SELECT 
    pt.start_usage_amount,
    pt.price_units,
    pt.price_nanos
FROM pricing_tiers pt
JOIN pricing_info pi ON pt.pricing_info_id = pi.id
WHERE pi.sku_id = 'ABCD-1234-5678'
  AND pi.effective_time <= CURRENT_TIMESTAMP
ORDER BY pi.effective_time DESC, pt.start_usage_amount ASC;
```

## Future Extensions

The schema is designed to support future enhancements:

1. **Multi-cloud Support**: Additional tables for AWS and Azure pricing
2. **User Preferences**: Tables for storing user-specific settings and favorites
3. **Price Alerts**: Tables for tracking price change notifications
4. **Usage Estimates**: Tables for storing usage patterns and projections
5. **Currency Support**: Extension to support multiple currencies with exchange rates

## Migration Strategy

When implementing this schema:

1. **Phase 1**: Core tables (`services`, `skus`, `pricing_info`, `pricing_tiers`)
2. **Phase 2**: Audit table (`pricing_updates`) and basic indexes
3. **Phase 3**: Full-text search capabilities and advanced indexes
4. **Phase 4**: Performance optimization based on usage patterns

This phased approach allows for incremental development while maintaining data integrity and performance.