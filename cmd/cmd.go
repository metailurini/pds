package cmd

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"pds/internal"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var logger *zap.Logger
var rootCmd = &cobra.Command{
	Use:  "pds",
	Long: `pds is used for listen changes of pods in a given namespace k8s to fire notification in time`,
	Run: func(cmd *cobra.Command, args []string) {
		// TODO
		log.Println("rootCmd")
	},
}

func init() {
	// init logger
	fileLog := initLogFileName()
	if lgg, err := internal.InitLogger(fileLog); err != nil {
		log.Fatal(err)
	} else {
		logger = lgg
	}
}

// truncateToDay round time to start moment of this time
func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// initLogFileName is used to init log file name, it will be `/tmp` in linux
func initLogFileName() string {
	tmpDir := os.TempDir()
	now := truncateToDay(time.Now()).Unix()
	fileLog := path.Join(tmpDir, fmt.Sprintf("pds.%d", now))
	return fileLog
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
