package scheduler

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"ping-admin-exporter/internal/cache"
	"ping-admin-exporter/internal/client"
	"ping-admin-exporter/internal/config"
	"ping-admin-exporter/internal/exporter"
	"ping-admin-exporter/internal/utils"
)

// runStatsScheduler starts a loop that periodically updates task stats and publish them
func RunStatsScheduler(exporters []*exporter.Exporter, cfg *config.Config, stop <-chan struct{}) {
	statsLog := logrus.WithField("component", "stats_scheduler")

	runCycle := func() {
		cycleStartTime := time.Now()
		statsLog.Info("Starting new stats refresh cycle...")

		var wg sync.WaitGroup
		var mu sync.Mutex

		// All Task Stats will be here
		allStats := make([]*client.TaskStatEntry, 0, len(exporters))

		wg.Add(len(exporters))
		for _, exp := range exporters {

			utils.RandomizedPause(cfg.RequestDelay)

			go func(e *exporter.Exporter) {
				defer wg.Done()

				// perform request about all tasks only once
				if e == exporters[0] {
					allTasksInfo, err := e.UpdateAllTasksInfo()
					if err != nil {
						statsLog.WithFields(logrus.Fields{
							"task_id": e.Config.TaskID,
							"error":   err,
						}).Error("All Tasks info refresh failed")
						return
					}

					cache.AllTasksInfoCache, err = json.Marshal(allTasksInfo)
					if err != nil {
						statsLog.Errorf("Failed to marshal tasks info to JSON: %v", err)
						return
					}

				}

				stats, err := e.UpdateTaskStats()
				if err != nil {
					statsLog.WithFields(logrus.Fields{
						"task_id": e.Config.TaskID,
						"error":   err,
					}).Error("Stats refresh failed")
					return
				}

				mu.Lock()
				allStats = append(allStats, stats)
				mu.Unlock()

			}(exp)
		}

		wg.Wait()
		statsLog.Infof("All exporters finished stats refresh cycle in %s.", time.Since(cycleStartTime))

		//// transpose stats
		//transposedStats := make([]*client.TransposedTaskStatEntry, 0, len(allStats))
		//for _, originalStat := range allStats {
		//	transposedStats = append(transposedStats, originalStat.Transpose())
		//}

		finalJSON, err := json.Marshal(allStats)
		if err != nil {
			statsLog.Errorf("Failed to marshal aggregated transposed stats to JSON: %v", err)
			return
		}

		// safely update cache
		cache.TaskDataCache.UpdateCache(finalJSON)
		statsLog.Info("Successfully updated tasks JSON cache.")
	}

	ticker := time.NewTicker(cfg.RefreshInterval)
	defer ticker.Stop()

	runCycle() // first run starts without ticker

	for {
		select {
		case <-ticker.C:
			runCycle()
		case <-stop:
			logrus.Infof("Stopping stats scheduler...")
			return
		}
	}
}
