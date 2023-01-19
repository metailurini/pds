package internal

import (
	"context"
	"flag"
	"fmt"
	"k8s.io/apimachinery/pkg/watch"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	apiCoreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typeCoreV1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// clientSetInstanceLock is used by getClientSetInstance to init one time,
// it's the implementation of single ton pattern
var clientSetInstanceLock = &sync.Mutex{}

var clientSetInstance typeCoreV1.CoreV1Interface

// callback is used to execute tasks for a Pod when App listen changes in a namespace
type callback func(context.Context, *apiCoreV1.Pod) error

// App struct holds list dependencies for managing the watcher app
type App struct {
	clientSet typeCoreV1.CoreV1Interface
	logger    *zap.Logger
	group     *sync.WaitGroup
}

// getK8sConfig is used to init k8s client config based on local environment vars
func getK8sConfig() (*rest.Config, error) {
	var kubeConfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeConfig = flag.String("kubeConfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeConfig file")
	} else {
		kubeConfig = flag.String("kubeConfig", "", "absolute path to the kubeConfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeConfig)
	if err != nil {
		return nil, errors.Wrap(err, "BuildConfigFromFlags")
	}
	return config, nil
}

// getClientSetInstance is used to init one time k8s client set
func getClientSetInstance() (typeCoreV1.CoreV1Interface, error) {
	if clientSetInstance == nil {
		clientSetInstanceLock.Lock()
		defer clientSetInstanceLock.Unlock()
		if clientSetInstance == nil {
			config, err := getK8sConfig()
			if err != nil {
				return nil, errors.Wrap(err, "getK8sConfig")
			}

			clientSet, err := kubernetes.NewForConfig(config)
			if err != nil {
				return nil, errors.Wrap(err, "NewForConfig")
			}
			clientSetInstance = clientSet.CoreV1()
		}
	}
	return clientSetInstance, nil
}

// InitApp is used to construct app
func InitApp(logger *zap.Logger) (*App, error) {
	instance, err := getClientSetInstance()
	if err != nil {
		return nil, err
	}

	return &App{
		clientSet: instance,
		logger:    logger,
		group:     new(sync.WaitGroup),
	}, nil
}

// createWatcher will use k8s clientSet to create watcher instance for a namespace
func createWatcher(ctx context.Context, clientSet typeCoreV1.CoreV1Interface, namespace string, options metaV1.ListOptions) (watch.Interface, error) {
	return clientSet.Pods(namespace).Watch(ctx, options)
}

// WatchPodChanges is used to watch changes all pods of a given namespace,
// and execute list callback function
func (app *App) WatchPodChanges(ctx context.Context, namespace string, callbacks ...callback) error {
	watcher, err := createWatcher(ctx, app.clientSet, namespace, metaV1.ListOptions{})
	if err != nil {
		return errors.Wrap(err, "createWatcher")
	}

	for event := range watcher.ResultChan() {
		pod, ok := event.Object.(*apiCoreV1.Pod)
		if !ok {
			continue
		}

		for idx, fn := range callbacks {
			err := fn(ctx, pod)
			if err != nil {
				return errors.Wrapf(err, "fn[%d]", idx)
			}
		}
	}
	return nil
}

// getNameSpaces will use k8s clientSet to list all existed namespaces
func getNameSpaces(ctx context.Context, clientSet typeCoreV1.CoreV1Interface, options metaV1.ListOptions) ([]apiCoreV1.Namespace, error) {
	list, err := clientSet.Namespaces().List(ctx, options)
	if err != nil {
		return nil, errors.Wrap(err, "Namespaces.List")
	}
	return list.Items, nil
}

// WatchPodChangesAllNameSpaces is async function to watch all pod for all existed namespaces
// and execute list callback function
//
// For the safe usage it should be controlled by Context, after context was done
// use function Wait to wait all runner done their tasks
func (app *App) WatchPodChangesAllNameSpaces(ctx context.Context, callbacks ...callback) {
	namespaces, err := getNameSpaces(ctx, app.clientSet, metaV1.ListOptions{})
	if err != nil {
		app.logger.Fatal(errors.Wrap(err, "getNameSpaces").Error())
	}

	// distribute all callback functions for all namespaces
	app.group.Add(len(namespaces) * len(callbacks))

	for _, namespace := range namespaces {
		for _, fn := range callbacks {
			go func(namespace apiCoreV1.Namespace, fn callback) {
				defer app.group.Done()
				app.logger.Info(fmt.Sprintf("start watching %s", namespace.Name))

				if err := app.WatchPodChanges(ctx, namespace.Name, fn); err != nil {
					app.logger.Error(fmt.Sprintf("error while watching %s: %v", namespace.Name, err))
				}

				app.logger.Info(fmt.Sprintf("stop watching %s", namespace.Name))
			}(namespace, fn)
		}
	}
}

// Wait is used for wait all runner done their tasks
func (app *App) Wait() {
	app.group.Wait()
}
