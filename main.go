package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"syscall"
	"time"

	"pds/internal"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	apiCoreV1 "k8s.io/api/core/v1"
)

var logger *zap.Logger

func showFailedPod(ctx context.Context, pod *apiCoreV1.Pod) error {
	if pod.Status.Phase != apiCoreV1.PodFailed {
		return nil
	}

	logger.Warn(
		fmt.Sprintf(
			"= pod details: %v - %v - %v",
			pod.Namespace,
			pod.Name,
			pod.Status.Phase,
		),
	)
	return bell(ctx)
}

// bell will call a system command to play a song for notify purpose
func bell(ctx context.Context) error {
	return exec.
		CommandContext(ctx,
			"paplay",
			"/usr/share/sounds/freedesktop/stereo/message-new-instant.oga",
		).Run()
}

// initLogFileName is used to init log file name, it will be in the /tmp
func initLogFileName() string {
	tmpDir := os.TempDir()
	now := time.Now().Unix()
	fileLog := path.Join(tmpDir, fmt.Sprintf("pds.%d", now))
	return fileLog
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 24*time.Hour)

	if err := bell(ctx); err != nil {
		log.Fatal(err)
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	fileLog := initLogFileName()
	if lgg, err := internal.InitLogger(fileLog); err != nil {
		log.Fatal(err)
	} else {
		logger = lgg
	}

	app, err := internal.InitApp(logger)
	if err != nil {
		logger.Fatal(errors.Wrap(err, "InitApp").Error())
	}

	app.WatchChangesAllNameSpaces(ctx, showFailedPod)

	<-signals
	cancel()
	app.Wait()
}
