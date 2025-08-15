# Architecture Design Document

**Project:** MCP Server for Google Cloud Pricing Access
**Stack:** Go language, GCP Cloud Run (service + Job), Turso (libSQL), GitHub OAuth 2.1, GCP Cloud Billing Catalog API, MCP Go SDK

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
The database schema in Turso reflects the structure of pricing data exposed by the Catalog API. We will use the natural identifiers provided by the API—service_id for services and sku_id for SKUs—as primary keys. This eliminates the need for synthetic surrogate keys and simplifies synchronization logic.

The skus table includes a `category` field that stores the entire category object (resourceFamily, resourceGroup, usageType, serviceDisplayName) for each SKU as JSONB. This approach simplifies initial development while preserving flexibility. Additional columns (resource_family, resource_group, usage_type) can be added later if filtering becomes necessary.

The model also tracks pricing history. To simplify the schema, each `pricing_info` record captures a timestamp, summary, and the full pricing expression. Tiered rates are stored as a JSON array within the same record, which closely mirrors the nested structure of the API response. The architecture omits currency conversion overhead by limiting prices to USD only, using the API’s units+nanos representation to avoid float inaccuracies.

Below is a sketch of the proposed database schema:

**`services`**

| Column | Type | Description |
| --- | --- | --- |
| service_id | TEXT | Primary Key, from Catalog API |
| display_name | TEXT | Human-readable name of the service |
| business_entity_name | TEXT | The business entity providing the service |

**`skus`**

| Column | Type | Description |
| --- | --- | --- |
| sku_id | TEXT | Primary Key, from Catalog API |
| service_id | TEXT | Foreign Key to `services` table |
| sku_name | TEXT | Human-readable name of the SKU |
| description | TEXT | Description of the SKU |
| category | JSONB | Full category object from API |
| service_regions | JSONB | List of regions where SKU is available |
| geo_taxonomy | JSONB | Geographic taxonomy information |

**`pricing_info`**

| Column | Type | Description |
| --- | --- | --- |
| pricing_info_id | INTEGER | Primary Key, auto-incrementing |
| sku_id | TEXT | Foreign Key to `skus` table |
| effective_time | TIMESTAMP | When this pricing became effective |
| summary | TEXT | A summary of the pricing |
| currency_code | TEXT | Currency of the price (e.g., "USD") |
| usage_unit | TEXT | The unit of usage (e.g., "GIBI.H") |
| usage_unit_description| TEXT | Description of the usage unit |
| display_quantity | INTEGER | The quantity the price is for |
| tiered_rates | JSONB | JSON array of tiered rate objects. Each object contains `start_usage_amount`, `units`, and `nanos`. |

**`pricing_updates`**

| Column | Type | Description |
| --- | --- | --- |
| update_id | INTEGER | Primary Key, auto-incrementing |
| update_time | TIMESTAMP | Timestamp of the sync job run |
| status | TEXT | 'SUCCESS' or 'FAILURE' |
| services_updated | INTEGER | Number of services updated |
| skus_updated | INTEGER | Number of SKUs updated |
| log_message | TEXT | Log message from the sync job |

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
