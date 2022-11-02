package informers

import (
	"context"
	"github.com/boomatang/informers-test/pkg/informers/configmap"
	"github.com/boomatang/informers-test/pkg/informers/secret"
	"github.com/boomatang/informers-test/pkg/utils/trybuffer"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"time"
)

type Informer struct {
	ctx               context.Context
	logger            logr.Logger
	namespace         string
	configmapInformer *configmap.Informer
	secretInformer    *secret.Informer
	clientset         *kubernetes.Clientset
	try               *trybuffer.TryBuffer
}

type InformerConfig struct {
	Logger    logr.Logger
	Namespace string
	Clientset *kubernetes.Clientset
}

func NewInformer(conf *InformerConfig) *Informer {
	return &Informer{
		logger:    conf.Logger,
		namespace: conf.Namespace,
		clientset: conf.Clientset,
	}
}

func (c *Informer) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	c.ctx = ctx

	r, err := configmap.LoadData()
	if err != nil {
		c.logger.Error(err, "loading configmap data")
	}

	configMapInformer := configmap.NewConfigMapInformer(&configmap.InformerConfig{
		Logger:    c.logger.WithName("configmap"),
		Namespace: c.namespace,
		Clientset: c.clientset,
		Required:  r,
	})
	c.configmapInformer = configMapInformer

	secretInformer := secret.NewsecretInformer(&secret.InformerConfig{
		Logger:    c.logger.WithName("secret"),
		Namespace: c.namespace,
		Clientset: c.clientset,
	})
	c.secretInformer = secretInformer

	go c.startResourceInformer(configMapInformer, cancel, "Run ConfigMap Informer")
	go c.startResourceInformer(secretInformer, cancel, "Run Secret Informer")

	select {
	case <-c.ctx.Done():
		c.try.Close()
		return c.ctx.Err()
	case <-time.After(5 * time.Second):
		c.try.Try()
	}

	wait.Until(c.sync, time.Minute, c.ctx.Done())
	c.try.Close()
	return c.ctx.Err()
}

func (c *Informer) startResourceInformer(informer ResourceInformer, cancel context.CancelFunc, msg string) {
	err := informer.Run(c.ctx)
	if err != nil {
		c.logger.Error(err, msg)
	}
	cancel()
}

func (c *Informer) sync() {
	c.secretInformer.Sync()
	c.configmapInformer.Sync()

}
