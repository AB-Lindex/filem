package metrics

import (
	"time"
)

// Target is the abstract interface a metrics-receiver
type Target interface {
	Set(name string, value interface{}, keys map[string]interface{}) error
	Send() error
}

// All the metric-names
const (
	VersionInfo       = "filem_version_info"
	MetricsStart      = "filem_start"
	MetricsEnd        = "filem_end"
	FilesMatched      = "filem_files_match"
	FilesUnmatched    = "filem_files_match_failures"
	FilesDone         = "filem_files_done"
	FilesDeletes      = "filem_files_deleted"
	FilesDeleteFailed = "filem_files_delete_failures"
	NamegenFailed     = "filem_name_generator_failures"
	SaveFailed        = "filem_storage_failures"
	SendFailed        = "filem_message_failures"
	SendMissingID     = "filem_message_missing_id_failure"
)

// Now return now as seconds with microsecond granularity
func Now() float64 {
	return float64(time.Now().UnixMicro()) / 1000000.0
}
