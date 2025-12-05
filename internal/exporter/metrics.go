package exporter

import (
	"github.com/prometheus/client_golang/prometheus"

	"apatit/internal/version"
)

const (
	namespace         = "ping_admin"
	subsystemExporter = "exporter"
	subsystemMP       = "mp"
)

// Monitoring Point metrics labels
var (
	mpLabels = []string{
		LabelTaskID,
		LabelTaskName,
		LabelMPID,
		LabelMPName,
		LabelMPIP,
		LabelMPGPS,
	}
)

// Metrics starts with "A" are related to "APATIT" itself
// Metrics starts with "E" are related to "Exporter"
// Metrics starts with "MP" are related to "Monitoring Point"
var (
	AServiceInfo = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "apatit",
			Name:      "service_info",
			Help:      "Information about the APATIT service.",
			ConstLabels: prometheus.Labels{
				"language": version.Language,
				"name":     version.Name,
				"owner":    version.Owner,
				"version":  version.Version,
			},
		},
	)

	ERefreshIntervalSeconds = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystemExporter,
			Name:      "refresh_interval_seconds",
			Help:      "The configured interval for refreshing metrics.",
		},
	)

	EMaxAllowedStalenessSteps = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystemExporter,
			Name:      "max_allowed_staleness_steps",
			Help: "Configured staleness threshold in steps. " +
				"If `ping_admin_mp_data_staleness_steps` exceeds this value the MP " +
				"is considered potentially unavailable.",
		},
	)

	ERefreshDurationSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystemExporter,
			Name:      "refresh_duration_seconds",
			Help:      "The duration of the last metrics refresh cycle for a specific task.",
		},
		[]string{"task_id", "task_name"},
	)

	ELoopsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystemExporter,
			Name:      "loops_total",
			Help:      "Total number of refresh loops started for a specific task.",
		},
		[]string{"task_id", "task_name"},
	)

	EErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystemExporter,
			Name:      "errors_total",
			Help:      "Total number of errors during metrics refresh for a specific task.",
		},
		[]string{"task_id", "task_name"},
	)

	MPStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystemMP,
			Name:      "status",
			Help:      "Status of the monitoring point (1 = up/processed, 0 = stale/down).",
		},
		mpLabels,
	)

	MPConnectSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystemMP,
			Name:      "connect_seconds",
			Help:      "Time spent establishing a connection.",
		},
		mpLabels,
	)

	MPDNSLookupSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystemMP,
			Name:      "dns_lookup_seconds",
			Help:      "Time spent on DNS lookup.",
		},
		mpLabels,
	)

	MPServerProcessingSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystemMP,
			Name:      "server_processing_seconds",
			Help:      "Time the server spent processing the request.",
		},
		mpLabels,
	)

	MPTotalDurationSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystemMP,
			Name:      "total_duration_seconds",
			Help:      "Total request time.",
		},
		mpLabels,
	)

	MPSpeedBytesPerSecond = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystemMP,
			Name:      "speed_bytes_per_second",
			Help:      "Download speed in bytes per second.",
		},
		mpLabels,
	)

	MPLastSuccessTimestampSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystemMP,
			Name:      "last_success_timestamp_seconds",
			Help:      "Timestamp of the last successful data point from the API.",
		},
		mpLabels,
	)

	MPLastSuccessDeltaSeconds = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystemMP,
			Name:      "last_success_delta_seconds",
			Help:      "Time since the last successful data point was received.",
		},
		mpLabels,
	)

	MPDataStalenessSteps = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystemMP,
			Name:      "data_staleness_steps",
			Help:      "How many API data steps have been missed for this MP. 0 means the data is fresh.",
		},
		mpLabels,
	)
)

func RegisterMetrics() {
	prometheus.MustRegister(
		AServiceInfo,
		ERefreshIntervalSeconds,
		EMaxAllowedStalenessSteps,
		ERefreshDurationSeconds,
		ELoopsTotal,
		EErrorsTotal,
		MPStatus,
		MPConnectSeconds,
		MPDNSLookupSeconds,
		MPServerProcessingSeconds,
		MPTotalDurationSeconds,
		MPSpeedBytesPerSecond,
		MPLastSuccessTimestampSeconds,
		MPLastSuccessDeltaSeconds,
		MPDataStalenessSteps,
	)
}

// DeleteSeries deletes "Monitoring Point" related metrics
func DeleteSeries(labels prometheus.Labels) {
	MPStatus.Delete(labels)
	MPConnectSeconds.Delete(labels)
	MPDNSLookupSeconds.Delete(labels)
	MPServerProcessingSeconds.Delete(labels)
	MPTotalDurationSeconds.Delete(labels)
	MPSpeedBytesPerSecond.Delete(labels)
	MPLastSuccessTimestampSeconds.Delete(labels)
	MPLastSuccessDeltaSeconds.Delete(labels)
	MPDataStalenessSteps.Delete(labels)
}
