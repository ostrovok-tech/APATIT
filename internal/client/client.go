package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"apatit/internal/utils"
	"apatit/internal/version"
)

const (
	defaultEndpoint = "https://ping-admin.com"
)

var apiKeyMasker = regexp.MustCompile(`(api_key=)(\w+)`)

// Client is a client to connect with Ping-Admin API.
type Client struct {
	httpClient     *http.Client
	apiKey         string
	endpoint       string
	requestDelay   time.Duration
	requestRetries int

	// Rate limiting: configurable max requests per second
	maxRequestsPerSecond int
	rateLimitMu          sync.Mutex
	lastRequestTimes     []time.Time // Track recent request times
}

// New creates a new API client entity.
func New(apiKey string, httpClient *http.Client, requestDelay time.Duration, requestRetries int, maxRequestsPerSecond int) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	if maxRequestsPerSecond <= 0 {
		maxRequestsPerSecond = 2
	}
	return &Client{
		httpClient:           httpClient,
		apiKey:               apiKey,
		endpoint:             defaultEndpoint,
		requestDelay:         requestDelay,
		requestRetries:       requestRetries,
		maxRequestsPerSecond: maxRequestsPerSecond,
	}
}

// waitForRateLimit ensures we don't exceed maxRequestsPerSecond requests per second.
// It waits if necessary before allowing the next request.
// This method is thread-safe and can be called from multiple goroutines.
func (c *Client) waitForRateLimit() {
	c.rateLimitMu.Lock()
	defer c.rateLimitMu.Unlock()

	now := time.Now()
	oneSecondAgo := now.Add(-time.Second)

	// Remove request times older than 1 second
	validRequests := make([]time.Time, 0, c.maxRequestsPerSecond)
	for _, t := range c.lastRequestTimes {
		if t.After(oneSecondAgo) {
			validRequests = append(validRequests, t)
		}
	}
	c.lastRequestTimes = validRequests

	// If we have maxRequestsPerSecond requests in the last second, wait until the oldest one expires
	if len(c.lastRequestTimes) >= c.maxRequestsPerSecond {
		oldestRequest := c.lastRequestTimes[0]
		// Wait 1.1 seconds instead of 1 second to add a small buffer and avoid exact 1 second API restriction
		waitDuration := 1100*time.Millisecond - now.Sub(oldestRequest)
		if waitDuration > 0 {
			logrus.WithField("wait_seconds", waitDuration.Seconds()).
				Debug("Rate limit: waiting before next API request")
			// Sleep while holding the lock - this ensures other goroutines wait
			// and see the updated state after we're done
			time.Sleep(waitDuration)
			now = time.Now() // Update now after waiting

			// Recalculate valid requests after sleep (time has passed)
			oneSecondAgo = now.Add(-time.Second)
			validRequests = make([]time.Time, 0, 2)
			for _, t := range c.lastRequestTimes {
				if t.After(oneSecondAgo) {
					validRequests = append(validRequests, t)
				}
			}
			c.lastRequestTimes = validRequests
		}
	}
	c.lastRequestTimes = append(c.lastRequestTimes, now)
}

// getAPI make a request to Ping-Admin API.
// Request could be delayed to avoid "Server Unavailable" error.
func (c *Client) getAPI(path string, result interface{}, delayed bool) error {
	log := logrus.WithField("component", "api_client")

	// Enforce rate limit: max 2 requests per second
	c.waitForRateLimit()

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %s", maskAPIKey(err.Error()))
	}
	req.Header.Set("User-Agent", fmt.Sprintf("%s/%s", version.Name, version.Version))

	var resp *http.Response
	for i := 1; i < c.requestRetries+1; i++ {

		if delayed {
			utils.RandomizedPause(c.requestDelay)
		}
		resp, err = c.httpClient.Do(req)

		log.WithField("url", maskAPIKey(req.URL.String())).Debug("Sending API request")

		if err != nil {
			log.WithFields(logrus.Fields{
				"url":   maskAPIKey(req.URL.String()),
				"error": maskAPIKey(err.Error()),
			}).Warn("Failed to send API request")

			if i < c.requestRetries {
				log.WithField("url", maskAPIKey(req.URL.String())).Info("Trying to send this request again..")
				utils.RandomizedPause(c.requestDelay)
			} else {
				return fmt.Errorf("request failed: %s", maskAPIKey(err.Error()))
			}

		} else {
			break
		}
	}

	if resp == nil {
		return fmt.Errorf("no response after \"%s\" request", maskAPIKey(req.URL.String()))
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if err := json.Unmarshal(body, result); err != nil {
		return fmt.Errorf("failed to unmarshal JSON response: %w", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to close response body: %w", err)
	}

	return nil
}

// GetTaskGraphStat get task statistics using sa=task_graph_stat request.
func (c *Client) GetTaskGraphStat(taskID int) ([]*MonitoringPointEntry, error) {
	u := fmt.Sprintf(
		"%s/?a=api&sa=task_graph_stat&enc=utf8&api_key=%s&id=%d&notnull=1&limit=1",
		c.endpoint, c.apiKey, taskID,
	)

	var resultsRaw []*EntryRaw
	if err := c.getAPI(u, &resultsRaw, false); err != nil {
		return nil, err
	}

	results := make([]*MonitoringPointEntry, len(resultsRaw))
	for i, r := range resultsRaw {
		results[i] = r.ProcessMonitoringPointEntry()
	}

	return results, nil
}

// GetTaskStat get task status using sa=task_stat request.
func (c *Client) GetTaskStat(taskID int) (*TaskStatEntry, error) {
	u := fmt.Sprintf(
		"%s/?a=api&sa=task_stat&enc=utf8&api_key=%s&id=%d&limit=100",
		c.endpoint, c.apiKey, taskID,
	)

	var resultsRaw []*TaskStatRaw
	if err := c.getAPI(u, &resultsRaw, false); err != nil {
		return nil, err
	}

	if len(resultsRaw) == 0 {
		return nil, fmt.Errorf("no task stat entries returned for task %d", taskID)
	}

	processedResult := resultsRaw[0].ProcessTaskEntry()

	return processedResult, nil
}

// GetMPs get monitoring points info by sa=tm request.
func (c *Client) GetMPs() ([]*MonitoringPointInfo, error) {
	u := fmt.Sprintf("%s/?a=api&sa=tm&enc=utf8&api_key=%s", c.endpoint, c.apiKey)

	var mps []*MonitoringPointRaw
	if err := c.getAPI(u, &mps, true); err != nil {
		return nil, err
	}

	processedMonitoringPointsInfo := make([]*MonitoringPointInfo, 0, len(mps))
	for _, mp := range mps {
		processedMonitoringPointsInfo = append(processedMonitoringPointsInfo, mp.ProcessMonitoringPointInfo())
	}

	return processedMonitoringPointsInfo, nil
}

// GetAllTasks get all tasks list.
func (c *Client) GetAllTasks() ([]*TaskInfo, error) {
	u := fmt.Sprintf("%s/?a=api&sa=tasks&enc=utf8&api_key=%s", c.endpoint, c.apiKey)

	var tasks []*TaskRaw
	if err := c.getAPI(u, &tasks, true); err != nil {
		return nil, err
	}

	processedTasks := make([]*TaskInfo, 0, len(tasks))
	for _, task := range tasks {
		processedTasks = append(processedTasks, task.ProcessTaskInfo())
	}

	return processedTasks, nil
}

// maskAPIKey change api_key in string on '***' for safe logging.
func maskAPIKey(str string) string {
	return apiKeyMasker.ReplaceAllString(str, "${1}***")
}
