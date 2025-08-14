# Architecture Design Document

## 1. Overview
The goal of this project is to build a Model Context Protocol (MCP) server that enables AI agents to access Google Cloud's publicly available pricing information through standardized tool interfaces: search, details, and calculate. These AI-native tools will facilitate discovery, metadata retrieval, and cost estimation. Pricing data will be synchronized from the Cloud Billing Catalog API into a Turso (libSQL) database on a daily schedule to ensure low-latency access, all underpinned by secure authentication using GitHub's OAuth 2.1 flow with PKCE, following the MCP authorization specification.

## 2. High-Level System Architecture
The system consists of three major components:

The MCP Server, hosted on Cloud Run, which handles tool requests from AI agents.

A Cloud Run Job, triggered daily by Cloud Scheduler to sync pricing data.

Turso, the database that persists normalized pricing data and supports efficient querying.

The MCP Server listens for agent requests over HTTP. It is responsible for authenticating these requests using GitHub OAuth 2.1 with PKCE, in compliance with MCP spec requirements. Once authenticated, agents may invoke any of the three tools, each implemented using the official MCP Go SDK.

Separately, the sync job periodically queries the Cloud Billing Catalog API through Google’s official Go client (cloud.google.com/go/billing/apiv1) to retrieve service and SKU data. Upon retrieving this data, the job transforms it into the Turso schema and persists it. This separation of responsibilities ensures that the API’s operational flow is never impacted by batch updates, and vice versa.

## 3. GCP Integration and Tooling
The Cloud Run Job is utilized to schedule batch pricing updates. This avoids maintaining a continuously running service for that purpose, and Cloud Run Jobs are designed for task execution to completion. Documentation confirms that Cloud Run Jobs can be executed via Cloud Scheduler, enabling a simple cron-like trigger. The job executes, writes logs to Cloud Logging, and provides monitoring signals without affecting the MCP server’s availability.

To retrieve pricing data programmatically, the architecture uses Google’s Cloud Catalog API. The official Go client, cloud.google.com/go/billing/apiv1, supports listing all services and associated SKUs, including detailed pricing expressions. The system will leverage this client to fetch and paginate through SKU listings.

## 4. Data Model in Turso

### 4.1 Schema Overview
The database uses **Turso (libSQL)** as the storage layer and is designed to mirror the structure of Google Cloud's Billing Catalog API while optimizing for query performance. The schema consists of several core tables that maintain pricing data, service metadata, and synchronization audit trails.

We use the natural identifiers provided by the API—service_id for services and sku_id for SKUs—as primary keys. This eliminates the need for synthetic surrogate keys and simplifies synchronization logic.

### 4.2 Core Tables

#### 4.2.1 `services` Table
Stores Google Cloud service metadata.

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `service_id` | TEXT | PRIMARY KEY | Unique identifier from Google Cloud API (e.g., "6F81-5844-456A") |
| `display_name` | TEXT | NOT NULL | Human-readable service name (e.g., "Compute Engine") |
| `description` | TEXT | | Detailed description of the service |
| `created_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Record creation timestamp |
| `updated_at` | DATETIME | DEFAULT CURRENT_TIMESTAMP | Last update timestamp |

**Purpose**: Provides service-level metadata for organizing and searching SKUs.

#### 4.2.2 `skus` Table
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

The skus table includes a category_json field that stores the entire category object (resourceFamily, resourceGroup, usageType, serviceDisplayName) for each SKU. This approach simplifies initial development while preserving flexibility. Additional columns (resource_family, resource_group, usage_type) are extracted for efficient filtering.

**Category JSON Structure**:
```json
{
  "serviceDisplayName": "Compute Engine",
  "resourceFamily": "Compute",
  "resourceGroup": "N1",
  "usageType": "OnDemand"
}
```

#### 4.2.3 `pricing_info` Table
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

The model tracks pricing history: each pricing_info record captures a timestamp and summary, with child tables storing pricing expressions and tiered rates, precisely mirroring the nested structure of the API. The architecture omits currency conversion overhead by limiting prices to USD only, using the API's units+nanos representation to avoid float inaccuracies.

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

#### 4.2.4 `pricing_tiers` Table
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

#### 4.2.5 `pricing_updates` Table
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

### 4.3 Data Flow and Relationships

```
services (1) ←→ (many) skus
    ↓
skus (1) ←→ (many) pricing_info
    ↓
pricing_info (1) ←→ (many) pricing_tiers
```

### 4.4 Indexes and Performance
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
```

### 4.5 Usage Patterns for MCP Tools

#### Search Tool Queries
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

#### Details Tool Queries
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

#### Calculate Tool Queries
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
## 5. Authentication and Security
Authentication is handled via GitHub using OAuth 2.1 with PKCE. This ensures secure sign-in for both public and confidential clients per MCP specifications. The server will offer OAuth metadata endpoints so MCP-compliant clients can discover necessary auth info dynamically, as required by the compliance draft.

Initially, all authenticated users will have the same level of access—there is no role-based access control (RBAC) yet. Personalization and per-user features are planned, but those will be introduced after the MVP stage.

Secrets, including GitHub credentials and database connection strings, are securely stored in GCP’s Secret Manager and injected into Cloud Run service environments as environment variables, following best practices.

## 6. Sync Job Mechanics
The sync functionality runs as a Cloud Run Job, scheduled through Cloud Scheduler. This design ensures that pricing synchronization runs independently and reliably. The job uses CloudCatalogClient to fetch services and SKUs, handles pagination, transforms data, and writes to the Turso database in structured tables. Each run is logged in Turso’s pricing_updates table for auditability.

This setup promotes decoupling between batch processing and request handling and avoids needless idle infrastructure.

## 7. Available Tool Interfaces
The MCP server exposes three tools:

search: returns lists of services and SKUs based on query criteria.

details: retrieves complete metadata and latest pricing information for a specified SKU.

calculate: computes cost estimates using SKU pricing tiers, region, and usage parameters.

These tools follow a versioned contract, facilitating backward compatibility as new features or providers (e.g., AWS/Azure) are added.

## 8. Scalability & Performance
Cloud Run automatically scales the MCP server based on request load, which, combined with fast Turso queries, ensures responsive tool execution. The system targets latency under 500 milliseconds for search and detail operations, and up to one second for price calculations.

The sync job runs off-hours and can scale in parallel (by splitting tasks) if needed, but initial implementation will run serially for simplicity.
