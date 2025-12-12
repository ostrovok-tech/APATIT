<div align="center">

# APATIT
## Advanced Ping-Admin Task Indicators Transducer

</div>

<div
style="text-align: center;">
  <img src="branding/banner.png" style="width: 100%; display: block; margin: 0 auto;">

</div>

---

<div
style="text-align: justify">

## About

**APATIT** is a set of exporters for the Website and Server Monitoring Service [Ping-Admin.com](https://ping-admin.com/?lang=en). **APATIT** collects, processes and publishes the tasks monitoring metrics and statistics.

## Features

- 🔄 **Automatic Metrics Collection**: Periodically fetches metrics from Ping-Admin API for multiple tasks
- 📊 **Prometheus Integration**: Exposes metrics in standard Prometheus format at `/metrics`
- 📈 **JSON Stats API**: Provides additional JSON endpoints for task statistics
- 🌍 **Location Translation**: Supports translation of location names via `locations.json`
- 🚀 **Concurrent Processing**: Efficiently processes multiple tasks in parallel
- 🔁 **Automatic Cleanup**: Removes stale metrics when monitoring points are no longer available
- 🐳 **Docker Support**: Ready-to-use Docker image

## Installation

### Using Docker

```bash
docker run --rm -d \
  --name apatit \
  -p 8080:8080 \
  -e API_KEY=your-api-key \
  -e TASK_IDS=1,2,3 \
  ghcr.io/ostrovok-tech/apatit:latest
```

### From Source

1. Clone the repository:
```bash
git clone https://github.com/ostrovok-tech/apatit.git
cd apatit
```

2. Build the binary:
```bash
go build -o apatit ./cmd/apatit
```

3. Run the exporter:
```bash
./apatit --api-key=your-api-key --task-ids=1,2,3
```

## Configuration

The exporter can be configured via command-line flags or environment variables.

### Required Parameters

| Flag | Environment Variable | Description | Default |
|------|---------------------|-------------|---------|
| `--api-key` | `API_KEY` | Ping-Admin API key | *required* |
| `--task-ids` | `TASK_IDS` | Comma-separated list of task IDs | *required* |

### Optional Parameters

| Flag | Environment Variable | Description | Default |
|------|---------------------|-------------|---------|
| `--listen-address` | `LISTEN_ADDRESS` | HTTP server listen address | `:8080` |
| `--log-level` | `LOG_LEVEL` | Log level (debug, info, warn, error) | `info` |
| `--locations-file` | `LOCATIONS_FILE` | Path to locations.json file | `locations.json` |
| `--eng-mp-names` | `ENG_MP_NAMES` | Translate MP names to English | `true` |
| `--refresh-interval` | `REFRESH_INTERVAL` | Metrics refresh interval | `3m` |
| `--api-update-delay` | `API_UPDATE_DELAY` | Ping-Admin API data update delay | `4m` |
| `--api-data-time-step` | `API_DATA_TIME_STEP` | Time between API data points | `3m` |
| `--max-allowed-staleness-steps` | `MAX_ALLOWED_STALENESS_STEPS` | Max staleness steps before marking MP as unavailable | `3` |
| `--max-requests-per-second` | `MAX_REQUESTS_PER_SECOND` | Maximum number of API requests allowed per second | `2` |
| `--request-delay` | `REQUEST_DELAY` | Minimum delay before API request (randomized) | `2s` |
| `--request-retries` | `REQUEST_RETRIES` | Maximum number of retries for API requests | `3` |

### Example Configuration

```bash
./apatit \
  --api-key=your-api-key \
  --task-ids=1,2,3 \
  --listen-address=:9090 \
  --refresh-interval=5m \
  --log-level=debug
```

Or using environment variables:

```bash
export API_KEY=your-api-key
export TASK_IDS=1,2,3
export REFRESH_INTERVAL=5m
export LOG_LEVEL=debug
./apatit
```

## Usage

### HTTP Endpoints

- **`/`** - Home page with links to metrics and stats
- **`/metrics`** - Prometheus metrics endpoint
- **`/stats?type=task`** - JSON endpoint for task statistics
- **`/stats?type=all`** - JSON endpoint for all tasks information

### Prometheus Configuration

Add the following to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'apatit'
    static_configs:
      - targets: ['localhost:8080']
```

## Metrics

The exporter exposes the following Prometheus metrics:

### Service Metrics

- `apatit_service_info` - Information about the APATIT service (version, name, owner)

### Exporter Metrics

- `apatit_exporter_refresh_interval_seconds` - Configured refresh interval
- `apatit_exporter_max_allowed_staleness_steps` - Configured staleness threshold
- `apatit_exporter_refresh_duration_seconds{task_id, task_name}` - Duration of last refresh cycle
- `apatit_exporter_loops_total{exporter_type}` - Total number of refresh loops for each exporter type
- `apatit_exporter_errors_total{error_module, error_type, task_id, task_name}` - Total number of errors

### Monitoring Point Metrics

All MP metrics include labels: `task_id`, `task_name`, `mp_id`, `mp_name`, `mp_ip`, `mp_gps`

- `apatit_mp_status` - Status of monitoring point (1 = up, 0 = down/stale)
- `apatit_mp_data_status` - Status of the data for the monitoring point (1 = has data, 0 = no data)
- `apatit_mp_connect_seconds` - Connection establishment time
- `apatit_mp_dns_lookup_seconds` - DNS lookup time
- `apatit_mp_server_processing_seconds` - Server processing time
- `apatit_mp_total_duration_seconds` - Total request duration
- `apatit_mp_speed_bytes_per_second` - Download speed
- `apatit_mp_last_success_timestamp_seconds` - Timestamp of last successful data point
- `apatit_mp_last_success_delta_seconds` - Time since last successful data point
- `apatit_mp_data_staleness_steps` - Number of missed API data steps (0 = fresh)

## Project Structure

```
apatit/
├── cmd/
│   └── apatit/
│       └── main.go              # Application entry point
├── internal/
│   ├── cache/                   # Cache implementation
│   ├── client/                  # Ping-Admin API client
│   ├── config/                  # Configuration management
│   ├── exporter/                # Metrics and stats exporters logic
│   ├── log/                     # Logging setup
│   ├── scheduler/               # Metrics and stats schedulers
│   ├── server/                  # HTTP server
│   ├── translator/              # Location name translation
│   ├── utils/                   # Utility functions
│   └── version/                 # Version information
├── deploy/
│   └── docker-compose.yaml      # Docker Compose configuration
├── Dockerfile                   # Container image definition
├── locations.json               # Location translation mappings
└── go.mod                       # Go module definition
```

## Support

For issues and feature requests, please use the [GitHub Issues](https://github.com/ostrovok-tech/apatit/issues) page.

</div>