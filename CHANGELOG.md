# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [v1.0.0] - 2025-12-03

### Added
- Prometheus metrics exporter for Ping-Admin monitoring service
- HTTP server with `/metrics` endpoint for Prometheus scraping
- JSON stats API endpoints (`/stats?type=task` and `/stats?type=all`)
- Support for multiple task monitoring with concurrent processing
- Automatic metrics collection with configurable refresh intervals
- Location name translation via `locations.json` file
- Automatic cleanup of stale metrics when monitoring points are removed
- Comprehensive Prometheus metrics:
  - Exporter metrics (service info, refresh intervals, loops, errors)
  - Monitoring point metrics (status, connection time, DNS lookup, server processing, total duration, speed, timestamps, staleness)
- Configuration via command-line flags and environment variables
- Docker image support with multi-stage build
- Graceful shutdown handling with signal support
- Request retry mechanism with configurable retry count
- Randomized request delays to prevent API throttling
- Rate limiting with configurable maximum requests per second (default: 2 requests/second)
- Staleness detection and reporting for monitoring points
- Support for English MP name translation
- Docker Compose configuration for easy deployment
- CI/CD workflow with linting and building
