package exporter

// Prometheus Labels
// Task is a monitoring task created in Ping-Admin.com
// MP (Monitoring Point) is a monitoring point for each task (tm in API response)
const (
	LabelErrorModule  = "error_module"
	LabelErrorType    = "error_type"
	LabelExporterType = "exporter_type"
	LabelTaskID       = "task_id"
	LabelTaskName     = "task_name"
	LabelMPID         = "mp_id"
	LabelMPName       = "mp_name"
	LabelMPIP         = "mp_ip"
	LabelMPGPS        = "mp_gps"
)
