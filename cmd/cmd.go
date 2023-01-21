package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/metailurini/pds/common"
	"github.com/metailurini/pds/internal"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"k8s.io/client-go/util/homedir"
)

var kubeConfigVar string

var logger *zap.Logger

var rootCmd = &cobra.Command{
	Use:  "pds",
	Long: `pds is used for listen changes of pods in a given namespace k8s to fire notification in time`,
}

func init() {
	initKubeConfig()
	if err := initLog(); err != nil {
		log.Panic(err)
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initLog() error {
	// init logger
	fileLog := common.InitLogFileName()
	lgg, err := internal.InitLogger(fileLog)
	if err != nil {
		return errors.Wrap(err, "InitLogger")
	}
	logger = lgg
	return nil
}

func initKubeConfig() {
	kubeConfigValue := ""
	if home := homedir.HomeDir(); home != "" {
		kubeConfigValue = filepath.Join(home, ".kube", "config")
	}
	rootCmd.PersistentFlags().StringVar(&kubeConfigVar, "kubeConfigVar", kubeConfigValue, "set kube config")
}
