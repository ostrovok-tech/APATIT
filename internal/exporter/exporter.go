package exporter

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"

	"apatit/internal/client"
	"apatit/internal/translator"
)

// Exporter collects metrics for a single task.
type Exporter struct {
	Config    *Config
	apiClient *client.Client
	log       *logrus.Entry

	taskInfo         *client.TaskInfo
	monitoringPoints []*client.MonitoringPointInfo
}

// Config contains the configuration for a specific Exporter instance.
type Config struct {
	TaskID          int
	EngMPNames      bool
	ApiUpdateDelay  time.Duration
	ApiDataTimeStep time.Duration
}

// New creates a new Exporter instance.
// Metadata (tasks, monitoring_points) is passed to avoid repeated API requests.
// allTasks map[string]*client.TaskRaw
func New(conf *Config, apiClient *client.Client, allTasks []*client.TaskInfo, mps []*client.MonitoringPointInfo) (*Exporter, error) {

	var taskInfo *client.TaskInfo
	for _, task := range allTasks {
		if task.ID == conf.TaskID {
			taskInfo = task
			break
		}
	}
	if taskInfo == nil {
		return nil, fmt.Errorf("task with ID %d not found in provided metadata", conf.TaskID)
	}

	log := logrus.WithFields(logrus.Fields{
		"component": "exporter",
		"task_id":   conf.TaskID,
		"task_name": taskInfo.ServiceName,
	})

	log.Debug("Exporter instance created")

	return &Exporter{
		Config:           conf,
		apiClient:        apiClient,
		log:              log,
		taskInfo:         taskInfo,
		monitoringPoints: mps,
	}, nil
}

func (e *Exporter) UpdateAllTasksInfo() ([]*client.TaskInfo, error) {
	e.log.Info("Updating all tasks info...")

	allTasks, err := e.apiClient.GetAllTasks()
	if err != nil {
		EErrorsTotal.WithLabelValues(
			"api_client", "get_all_tasks",
			strconv.Itoa(e.taskInfo.ID),
			e.taskInfo.ServiceName).Inc()
		return nil, fmt.Errorf("error getting all tasks info: %w", err)
	}

	return allTasks, nil
}

// UpdateTaskStats get task_stat data from the API and converts it to JSON.
func (e *Exporter) UpdateTaskStats() (*client.TaskStatEntry, error) {

	e.log.Info("Updating task stats...")

	// Get and process task stats
	taskStatResults, err := e.apiClient.GetTaskStat(e.Config.TaskID)
	if err != nil {
		EErrorsTotal.WithLabelValues(
			"api_client", "get_task_stat",
			strconv.Itoa(e.taskInfo.ID),
			e.taskInfo.ServiceName).Inc()
		return nil, fmt.Errorf("failed to get task stat: %w", err)
	}

	e.processTaskStatResults(taskStatResults)
	ELoopsTotal.WithLabelValues("stats").Inc()

	return taskStatResults, nil
}

// RefreshMetrics requests new data from the API, updates Prometheus metrics
// and returns a list of labels for the processed series.
func (e *Exporter) RefreshMetrics() ([]prometheus.Labels, error) {
	startTime := time.Now()
	e.log.Info("Refreshing metrics...")

	// Updating the exporter's metrics
	defer func() {
		duration := time.Since(startTime).Seconds()
		ELoopsTotal.WithLabelValues("metrics").Inc()
		ERefreshDurationSeconds.WithLabelValues(strconv.Itoa(e.taskInfo.ID), e.taskInfo.ServiceName).Set(duration)
		e.log.WithField("duration_s", duration).Info("Refresh finished")
	}()

	// Get and process task graph stats (metrics)
	taskStatGraphResults, err := e.apiClient.GetTaskGraphStat(e.Config.TaskID)
	if err != nil {
		EErrorsTotal.WithLabelValues(
			"api_client",
			"get_task_graph_stat",
			strconv.Itoa(e.taskInfo.ID),
			e.taskInfo.ServiceName).Inc()
		return nil, fmt.Errorf("failed to get task graph stat: %w", err)
	}

	if len(taskStatGraphResults) == 0 {
		EErrorsTotal.WithLabelValues(
			"api_client",
			"get_task_graph_stat",
			strconv.Itoa(e.taskInfo.ID),
			e.taskInfo.ServiceName).Inc()
		e.log.Error("No MP data from API.")
	} else {
		e.log.Debugf("Received %d data items from API", len(taskStatGraphResults))
	}

	mpsInfo, err := e.apiClient.GetMPs()
	if err != nil {
		EErrorsTotal.WithLabelValues(
			"api_client", "get_mps",
			strconv.Itoa(e.taskInfo.ID),
			e.taskInfo.ServiceName).Inc()
		return nil, fmt.Errorf("failed to get monitoring points info: %w", err)
	}

	processedLabels := make([]prometheus.Labels, 0)

	for _, item := range taskStatGraphResults {
		for _, mp := range mpsInfo {
			if mp.ID == item.ID {
				item.Status = int(mp.Status)
				break
			}
		}
		// check monitoring point status and set it as ZERO if it was incorrect
		if item.Status > 1 || item.Status < 0 {
			e.log.WithFields(
				logrus.Fields{
					"mp_id": item.ID, 
					"mp_name": item.Name,
					"status": item.Status,
				}).Errorf("incorrect monitoring points status: %d", item.Status)
			item.Status = 0
		}

		labels := e.processTaskStatGraphResultItem(item, startTime)
		if labels != nil {
			processedLabels = append(processedLabels, labels...)
		}
	}

	return processedLabels, nil
}

func (e *Exporter) processTaskStatResults(taskStatResults *client.TaskStatEntry) {

	taskStatResults.TaskID = strconv.Itoa(e.taskInfo.ID)
	taskStatResults.TaskName = e.taskInfo.ServiceName
	taskStatResults.Timestamp = time.Now()

	for _, entry := range taskStatResults.TaskLogs {
		entry.Traceroute = strings.ReplaceAll(entry.Traceroute, "\\n", "\n")
		entry.MPName = translator.GetEngLocation(entry.MPName)
	}
}

// processTaskStatGraphResultItem processes one record (monitoring point) and updates metrics.
func (e *Exporter) processTaskStatGraphResultItem(item *client.MonitoringPointEntry, refreshStartTime time.Time) []prometheus.Labels {
	if len(item.Result) == 0 {
		locationName := item.Name
		if e.Config.EngMPNames {
			locationName = translator.GetEngLocation(item.Name)
		}
		MPDataStatus.WithLabelValues(
			strconv.Itoa(e.taskInfo.ID), 
			e.taskInfo.ServiceName,
			item.ID,
			locationName,
			).Set(0)

		e.log.WithFields(
			logrus.Fields{
				"mp_id": item.ID, 
				"mp_name": item.Name}).Warn("No results found for MP")
		return nil
	}

	// Usually there is only one element in the MPResult in the response, but just in case we go through them all.
	processedLabels := make([]prometheus.Labels, 0, len(item.Result))
	for _, res := range item.Result {
		labels := e.buildLabels(item)
		e.updateMetrics(res, labels, item.Status, refreshStartTime)
		processedLabels = append(processedLabels, labels)
	}

	MPDataStatus.WithLabelValues(
		strconv.Itoa(e.taskInfo.ID), 
		e.taskInfo.ServiceName,
		item.ID,
		processedLabels[0][LabelMPName],
		).Set(1)

	return processedLabels
}

// buildLabels creates a set of Prometheus labels for a monitoring point.
func (e *Exporter) buildLabels(item *client.MonitoringPointEntry) prometheus.Labels {
	locationName := item.Name
	if e.Config.EngMPNames {
		locationName = translator.GetEngLocation(item.Name)
	}

	ipAddress := "unknown"
	gpsCoordinates := "unknown"
	for _, mp := range e.monitoringPoints {
		if mp.ID == item.ID {
			ipAddress = mp.IP
			gpsCoordinates = mp.GPS
			break
		}
	}

	return prometheus.Labels{
		LabelTaskID:   strconv.Itoa(e.taskInfo.ID),
		LabelTaskName: e.taskInfo.ServiceName,
		LabelMPID:     item.ID,
		LabelMPName:   locationName,
		LabelMPIP:     ipAddress,
		LabelMPGPS:    gpsCoordinates,
	}
}

// updateMetrics sets values for all metrics based on data.
func (e *Exporter) updateMetrics(res *client.MonitoringPointConnectionResult, labels prometheus.Labels, mpStatus int, refreshStartTime time.Time) {
	ts := time.Unix(res.Timestamp, 0)
	lastCheckDelta := refreshStartTime.Sub(ts)

	// remove time related metrics and set MPStatus as ZERO if monitoring point was unavailable according to 'mp' API
	if mpStatus == 0 {
		e.log.WithFields(logrus.Fields{"mp_id": labels["mp_id"], "mp_name": labels["mp_name"]}).
			Warn("Monitoring point is unavailable")
		DeleteSeries(labels)
		MPStatus.With(labels).Set(0)
		return
	}

	// remove time related metrics and set MPStatus as ZERO if MP data is older than 24 hours
	if lastCheckDelta >= 24*time.Hour {
		e.log.WithFields(logrus.Fields{"mp_id": labels["mp_id"], "mp_name": labels["mp_name"]}).
			Warn("Data for MP is older than 24 hours")
		DeleteSeries(labels)
		MPStatus.With(labels).Set(0)
		return
	}

	// Calculate the latency in "steps" (how many API intervals have passed since the data was received)
	// This helps us understand how "old" the data is. 0 is the most recent.

	delayInSteps := math.Floor(math.Abs(lastCheckDelta.Seconds()-e.Config.ApiUpdateDelay.Seconds()) / e.Config.ApiDataTimeStep.Seconds())

	MPConnectSeconds.With(labels).Set(res.Connect)
	MPDNSLookupSeconds.With(labels).Set(res.DNS)
	MPServerProcessingSeconds.With(labels).Set(res.Server)
	MPTotalDurationSeconds.With(labels).Set(res.Total)
	MPSpeedBytesPerSecond.With(labels).Set(float64(res.Speed))
	MPLastSuccessTimestampSeconds.With(labels).Set(float64(res.Timestamp))
	MPLastSuccessDeltaSeconds.With(labels).Set(lastCheckDelta.Seconds())
	MPDataStalenessSteps.With(labels).Set(delayInSteps)
	MPStatus.With(labels).Set(1)

	e.log.WithFields(logrus.Fields{
		"mp_id":   labels["mp_id"],
		"mp_name": labels["mp_name"],
		"delta":   lastCheckDelta,
		"steps":   delayInSteps,
	}).Debug("Metrics updated for MP")

}
