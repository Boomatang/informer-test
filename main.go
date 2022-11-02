package main

import (
	"context"
	"github.com/boomatang/informers-test/pkg/utils/signals"
	"github.com/go-logr/zapr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/boomatang/informers-test/pkg/informers"
	"go.uber.org/zap"
)

var (
	namespace = "default"
)

func main() {

	logConfig := zap.NewDevelopmentConfig()
	zapLog, err := logConfig.Build()
	if err != nil {
		os.Exit(1)
	}
	log := zapr.NewLogger(zapLog)

	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{CurrentContext: ""}).ClientConfig()
	if err != nil {
		log.Error(err, "failed to create kubernetes client")
		os.Exit(1)
	}

	log.Info("Starting the informers")
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Error(err, "failed to create k8s clientset")
		os.Exit(1)
	}

	informer := informers.NewInformer(&informers.InformerConfig{
		Logger:    log.WithName("informer"),
		Namespace: namespace,
		Clientset: clientset,
	})

	stopCh := signals.SetupNotifySignalHandler()
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		<-stopCh
		cancel()
	}()

	err = informer.Run(ctx)
	if err != nil {
		log.Error(err, "unable to start main controller")
		os.Exit(1)
	}

}
