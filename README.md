# cloud-pricing-mcp
Google Cloud Pricing Calculator MCP (Model Context Protocol)

## Overview

This project provides a Model Context Protocol (MCP) server that enables AI agents to access Google Cloud pricing information through standardized tool interfaces. The server acts as a bridge between Google Cloud's Billing Catalog API and MCP-compatible AI agents, enabling automated cost analysis and optimization.

## Documentation

- **[Product Requirements Document (PRD)](PRD.md)** - Project goals, requirements, and specifications
- **[Architecture Document](ARCHITECTURE.md)** - System design and component overview  
- **[Database Schema](SCHEMA.md)** - Detailed database schema documentation

## Key Features

- **AI-Native Access**: Programmatic access to Google Cloud pricing data for AI agents
- **MCP Tool Interfaces**: Search, details, and calculation tools following MCP specifications
- **Real-time Data**: Daily synchronization with Google Cloud Billing Catalog API
- **Secure Authentication**: GitHub OAuth 2.1 with PKCE for secure access
- **High Performance**: Optimized queries with sub-500ms response times

## Architecture

The system consists of:
- **MCP Server**: Hosted on Google Cloud Run, handles tool requests from AI agents
- **Sync Job**: Daily Cloud Run Job to update pricing data from Google's API
- **Database**: Turso (libSQL) for fast, reliable data storage

For detailed architecture information, see [ARCHITECTURE.md](ARCHITECTURE.md).

## Database Schema

The system uses a normalized schema optimized for search and calculation operations. Key tables include:
- `services` - Google Cloud service metadata
- `skus` - Individual pricing units with categorization
- `pricing_info` - Time-based pricing records
- `pricing_tiers` - Detailed tiered pricing rates

For complete schema documentation, see [SCHEMA.md](SCHEMA.md).
