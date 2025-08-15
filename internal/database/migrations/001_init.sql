CREATE TABLE IF NOT EXISTS services (
    service_id TEXT PRIMARY KEY,
    display_name TEXT NOT NULL,
    business_entity_name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS skus (
    sku_id TEXT PRIMARY KEY,
    service_id TEXT NOT NULL,
    sku_name TEXT NOT NULL,
    description TEXT NOT NULL,
    category BLOB NOT NULL,
    service_regions BLOB NOT NULL,
    geo_taxonomy BLOB NOT NULL,
    FOREIGN KEY (service_id) REFERENCES services(service_id)
);

CREATE TABLE IF NOT EXISTS pricing_info (
    pricing_info_id INTEGER PRIMARY KEY AUTOINCREMENT,
    sku_id TEXT NOT NULL,
    effective_time TIMESTAMP NOT NULL,
    summary TEXT,
    currency_code TEXT NOT NULL,
    usage_unit TEXT NOT NULL,
    usage_unit_description TEXT,
    display_quantity INTEGER,
    tiered_rates BLOB NOT NULL,
    FOREIGN KEY (sku_id) REFERENCES skus(sku_id)
);

CREATE TABLE IF NOT EXISTS pricing_updates (
    update_id INTEGER PRIMARY KEY AUTOINCREMENT,
    update_time TIMESTAMP NOT NULL,
    status TEXT NOT NULL,
    services_updated INTEGER,
    skus_updated INTEGER,
    log_message TEXT
);
