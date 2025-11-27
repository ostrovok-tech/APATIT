package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"ping-admin-exporter/internal/config"
	"ping-admin-exporter/internal/exporter"
	"ping-admin-exporter/internal/utils"
)

// runMetricsScheduler starts a loop that periodically updates metrics and clears old ones.
func RunMetricsScheduler(exporters []*exporter.Exporter, cfg *config.Config, stop <-chan struct{}) {
	var lastRunMPSeries = make(map[string]prometheus.Labels)

	runCycle := func() {
		currentRunMPSeries := make(map[string]prometheus.Labels)
		var mu sync.Mutex
		var wg sync.WaitGroup

		metricsLog := logrus.WithField("component", "scheduler")
		cycleStartTime := time.Now()
		metricsLog.Info("Starting new metrics refresh cycle...")
		exporter.ERefreshIntervalSeconds.Set(cfg.RefreshInterval.Seconds())
		exporter.EMaxAllowedStalenessSteps.Set(float64(cfg.MaxAllowedStalenessSteps))

		wg.Add(len(exporters))
		for _, exp := range exporters {

			utils.RandomizedPause(cfg.RequestDelay)

			go func(e *exporter.Exporter) {
				defer wg.Done()

				processedLabels, err := e.RefreshMetrics()
				if err != nil {
					metricsLog.WithFields(logrus.Fields{
						"task_id": e.Config.TaskID,
						"error":   err,
					}).Error("Exporter refresh failed")
					return
				}

				mu.Lock()
				for _, labels := range processedLabels {
					seriesKey := fmt.Sprintf("%s:%s", labels["task_id"], labels["mp_id"])
					currentRunMPSeries[seriesKey] = labels
				}
				mu.Unlock()
			}(exp)
		}

		wg.Wait()
		metricsLog.Infof("All exporters finished refresh cycle in %s.", time.Since(cycleStartTime))

		// cleaning up absent metrics: comparing series from the previous run with the current ones
		for seriesKey, labels := range lastRunMPSeries {
			if _, exists := currentRunMPSeries[seriesKey]; !exists {
				metricsLog.WithField("series", seriesKey).Info("Deleting stale series")
				exporter.DeleteSeries(labels)
			}
		}
		lastRunMPSeries = currentRunMPSeries
		metricsLog.Info("Metrics cleanup finished. Waiting for the next cycle.")
	}

	ticker := time.NewTicker(cfg.RefreshInterval)
	defer ticker.Stop()

	runCycle() // first run starts without ticker

	for {
		select {
		case <-ticker.C:
			runCycle()
		case <-stop:
			logrus.Infof("Stopping metrics scheduler...")
			return
		}
	}
}
