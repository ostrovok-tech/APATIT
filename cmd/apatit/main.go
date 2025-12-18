package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"apatit/internal/client"
	"apatit/internal/config"
	"apatit/internal/exporter"
	"apatit/internal/log"
	"apatit/internal/scheduler"
	"apatit/internal/server"
	"apatit/internal/translator"
)

// createExporters creates and returns a list of exporters for the specified tasks.
func createExporters(apiClient *client.Client, cfg *config.Config) ([]*exporter.Exporter, error) {
	exportersLog := logrus.WithField("component", "initializer")
	exportersLog.Infof("Creating exporters for %d tasks...", len(cfg.TaskIDs))

	exporters := make([]*exporter.Exporter, 0, len(cfg.TaskIDs))

	// get account tasks
	tasks, err := apiClient.GetAllTasks()
	if err != nil {
		return nil, fmt.Errorf("failed to get tasks metadata: %w", err)
	}

	// get all available monitoring points
	mps, err := apiClient.GetMPs()
	if err != nil {
		return nil, fmt.Errorf("failed to get MPs metadata: %w", err)
	}

	// create exporter for each task
	for _, taskID := range cfg.TaskIDs {

		expConfig := &exporter.Config{
			TaskID:          taskID,
			EngMPNames:      cfg.EngMPNames,
			ApiUpdateDelay:  cfg.ApiUpdateDelay,
			ApiDataTimeStep: cfg.ApiDataTimeStep,
		}

		exp, err := exporter.New(expConfig, apiClient, tasks, mps)
		if err != nil {
			exportersLog.Errorf("Unable to create exporter for TaskID %d: %v", taskID, err)
			continue
		}
		exporters = append(exporters, exp)
	}

	if len(exporters) == 0 {
		return nil, fmt.Errorf("no exporters were created, check task IDs and API key")
	}

	exportersLog.Infof("Successfully created %d exporters.", len(exporters))
	return exporters, nil
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.New()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	app, err := newApp(cfg)
	if err != nil {
		logrus.Fatalf("Application initialization failed: %v", err)
	}

	if err := app.Run(ctx); err != nil {
		logrus.Fatalf("Application terminated with error: %v", err)
	}
}

type application struct {
	cfg       *config.Config
	exporters []*exporter.Exporter
	stop      chan struct{}
}

func newApp(cfg *config.Config) (*application, error) {
	// Set logger for components
	log.Init(cfg.LogLevel)

	// Set translator
	if err := translator.Init(cfg.LocationsFilePath); err != nil {
		logrus.Warnf("Failed to initialize translator, location names will not be translated: %v", err)
	}

	// Create API client
	apiClient := client.New(cfg.APIKey, nil, cfg.RequestDelay, cfg.RequestRetries, cfg.MaxRequestsPerSecond)

	// Register metrics, set ServiceInfo metric
	exporter.RegisterMetrics()
	exporter.AServiceInfo.Set(1)

	// Create Exporters for each task
	exporters, err := createExporters(apiClient, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create exporters: %w", err)
	}

	return &application{
		cfg:       cfg,
		exporters: exporters,
		stop:      make(chan struct{}),
	}, nil
}

func (a *application) Run(ctx context.Context) error {
	// Run HTTP server
	go server.StartServer(a.cfg.ListenAddress)

	// Task Statistic Loop
	go scheduler.RunStatsScheduler(a.exporters, a.cfg, a.stop)

	// Exporter Metrics Loop
	go scheduler.RunMetricsScheduler(a.exporters, a.cfg, a.stop)

	logrus.Info("Exporters are running. Press Ctrl+C to exit.")

	// Stop goroutines once context is canceled
	<-ctx.Done()

	logrus.Info("Shutdown signal received. Stopping schedulers...")
	close(a.stop)

	time.Sleep(2 * time.Second)

	logrus.Info("Shutdown complete. Bye!")
	return nil
}
