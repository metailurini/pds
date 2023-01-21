package common

import (
	"fmt"
	"os"
	"path"
	"time"
)

// InitLogFileName creates a unique log file name by joining
// the system's temp directory, current day and time.
// Return string returns the log file name
func InitLogFileName() string {
	tmpDir := os.TempDir()
	now := truncateToDay(time.Now()).Unix()
	fileLog := path.Join(tmpDir, fmt.Sprintf("pds.%d", now))
	return fileLog
}
