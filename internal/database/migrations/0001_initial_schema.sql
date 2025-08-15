-- 0001_initial_schema.sql

-- This table tracks which migrations have been applied.
CREATE TABLE IF NOT EXISTS schema_migrations (
    version TEXT PRIMARY KEY
);

-- Table for services as defined in ARCHITECTURE.md
CREATE TABLE IF NOT EXISTS services (
    service_id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    business_entity_name TEXT
);

-- Table for SKUs as defined in ARCHITECTURE.md
CREATE TABLE IF NOT EXISTS skus (
    sku_id TEXT PRIMARY KEY,
    service_id TEXT NOT NULL,
    sku_name TEXT NOT NULL,
    description TEXT NOT NULL,
    category BLOB, -- Stored as JSONB
    service_regions BLOB, -- Stored as JSONB
    geo_taxonomy BLOB, -- Stored as JSONB
    FOREIGN KEY (service_id) REFERENCES services(service_id)
);

-- Table for pricing information as defined in ARCHITECTURE.md
CREATE TABLE IF NOT EXISTS pricing_info (
    pricing_info_id INTEGER PRIMARY KEY AUTOINCREMENT,
    sku_id TEXT NOT NULL,
    effective_time TIMESTAMP NOT NULL,
    summary TEXT,
    currency_code TEXT NOT NULL,
    usage_unit TEXT NOT NULL,
    usage_unit_description TEXT,
    display_quantity INTEGER,
    tiered_rates BLOB, -- Stored as JSONB
    FOREIGN KEY (sku_id) REFERENCES skus(sku_id)
);

-- Table for tracking pricing updates as defined in ARCHITECTURE.md
CREATE TABLE IF NOT EXISTS pricing_updates (
    update_id INTEGER PRIMARY KEY AUTOINCREMENT,
    update_time TIMESTAMP NOT NULL,
    status TEXT NOT NULL,
    services_updated INTEGER,
    skus_updated INTEGER,
    log_message TEXT
);
