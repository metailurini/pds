package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"pds/internal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	apiCoreV1 "k8s.io/api/core/v1"
)

var (
	// shellScriptVar will store the shell script of user
	//
	// while the app run and detect change in a pod
	// app will execute the user's shell script with list params:
	//
	// - $1:string is name of pod
	//
	// - $2:json is status of pod
	shellScriptVar string
)

const (
	shellScriptDesc = `shell script (accept $1 is pod name - json, $2 is status - json)
to run when listened change`

	// shellScriptKey is the key to get/set value inside context
	shellScriptKey = iota
)

// singlePodCmd hold set commands for single pod controller
var singlePodCmd = &cobra.Command{
	Use:   "pod",
	Short: "for single pod",
	Run: func(cmd *cobra.Command, args []string) {
		if shellScriptVar == "" {
			logger.Fatal("the shell script can be not empty!")
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 24*time.Hour)
		ctx = context.WithValue(ctx, shellScriptKey, shellScriptVar)

		signals := make(chan os.Signal)
		signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

		app, err := internal.InitApp(logger)
		if err != nil {
			logger.Fatal(errors.Wrap(err, "InitApp").Error())
		}

		app.WatchPodChangesAllNameSpaces(ctx, taskSinglePod)

		<-signals
		cancel()
		app.Wait()
	},
}

func init() {
	singlePodCmd.PersistentFlags().StringVar(&shellScriptVar, "script", "", shellScriptDesc)
	rootCmd.AddCommand(singlePodCmd)
}

// taskSinglePod will trigger user's shell script
func taskSinglePod(ctx context.Context, pod *apiCoreV1.Pod) error {
	command, ok := ctx.Value(shellScriptKey).(string)
	if !ok {
		return fmt.Errorf("can not get shell script from context")
	}

	podName := pod.Name
	podStatus, err := json.Marshal(pod.Status)
	if err != nil {
		return errors.Wrap(err, "json.Marshal")
	}

	return exec.CommandContext(ctx, "sh", command, podName, string(podStatus)).Run()
}
