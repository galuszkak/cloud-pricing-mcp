# Product Requirements Document (PRD)

**Project:** MCP Server for Google Cloud Pricing Access
**Status:** Concept / Pet Project

## Goals and Background Context

The project aims to create a Model Context Protocol (MCP) server — a standardized interface that enables AI agents to interact with tools and structured data — which will provide programmatic access to Google Cloud pricing information. Acting as a bridge between the Google Cloud Pricing Calculator’s regularly updated datasets and MCP-compatible AI agents, the server will eliminate manual lookup and enable automated, intelligent cost analysis.

### Primary goals:

- **AI-Native Cloud Pricing Access** – Allow AI agents to query, search, and retrieve structured Google Cloud pricing data without human mediation.
- **Tool-Based Interaction** – Provide MCP-compliant tools for:
  - Searching services/SKUs
  - Retrieving details for a given service/SKU
  - Calculating prices for specific configurations
- **Future-Proof Extensibility** – Architect the solution to support additional cloud providers (AWS, Azure) in a later phase.
- **Accuracy and Optimization** – Enable precise, automated cost forecasting and optimization, reducing errors and improving decision-making.
- **Multi-Role Value** – Serve FinOps teams, developers, procurement analysts, and cloud architects with AI-driven cost insights.

## Requirements

### Functional Requirements

#### Data Ingestion

- **Primary:** Retrieve Google Cloud pricing and SKU metadata via the Cloud Billing Catalog API.
- Implement a local caching layer for faster queries and offline resilience.
- **Optional fallback:** ingest and normalize CSV datasets if API is unavailable.

#### MCP Tool Interfaces

- **Search Tool:** Locate services or SKUs based on keywords, categories, or filters.
- **Details Tool:** Retrieve metadata and pricing details for a given service or SKU.
- **Calculation Tool:** Accept configuration parameters (e.g., region, usage volume) and return estimated costs.

#### Extensibility Hooks

- Abstracted data processing layer to support integration of AWS, Azure, or other cloud providers in future phases.
- Configuration-driven data ingestion for adding new providers without core code changes.

#### Access & Security

- **Authentication:** Use GitHub as the Identity Provider via OAuth 2.1, following the MCP Authorization specification.
- **Flow:** Authorization Code with PKCE for public clients (CLI/desktop) and Authorization Code for confidential server clients.
- **Scopes:** Start with `read:user` and `user:email`; expand only if future features require more.
- **Token Handling:** Short-lived access tokens; refresh tokens where applicable; secure storage and rotation.
- **MCP Compliance:** Publish OAuth 2.0 Protected Resource Metadata so MCP clients can discover authentication details.
- No RBAC in initial release — all authenticated users have the same access level.

#### Monitoring & Logging

- Record MCP tool calls for debugging and performance monitoring.
- Basic health-check and status endpoints.

### Non-Functional Requirements

#### Performance

- **Search** and **Details** tools respond in under 500 ms for typical queries.
- **Calculate** tool returns results within 1 second for standard configurations.

#### Scalability

- Handle concurrent requests from multiple AI agents without performance degradation.

#### Reliability

- Target 99.9% uptime.
- Graceful degradation if partial data is unavailable.

#### Maintainability

- Modular code for easy updates to data schemas and MCP tool definitions.
- Automated tests for ingestion, search, and calculation.

#### Compliance

- Respect licensing terms of source datasets.
- Follow MCP protocol specifications.

## User Interface Design Goals

Although the MCP server does not present a traditional graphical UI, its “interface” is the structured set of MCP tools that AI agents interact with.

### Design goals:

- **MCP Tool Consistency** – Consistent schemas and minimal parameters.
- **Natural Language Agent Integration** – Clear tool descriptions and example calls for LLM understanding.
- **Error Transparency** – Standard error codes and soft-fail handling for partial data.
- **Versioned Tool Contracts** – Versioned interfaces for safe evolution.
- **Future Extensibility** – Schemas adaptable for multi-cloud support.
