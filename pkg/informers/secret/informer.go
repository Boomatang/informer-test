package secret

import (
	"context"
	"github.com/boomatang/informers-test/pkg/utils/objref"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"sync"
)

type InformerConfig struct {
	Logger    logr.Logger
	Namespace string
	Clientset *kubernetes.Clientset
}

type Informer struct {
	ctx       context.Context
	mut       sync.RWMutex
	namespace string
	logger    logr.Logger
	clientset *kubernetes.Clientset
}

func NewsecretInformer(conf *InformerConfig) *Informer {
	return &Informer{
		namespace: conf.Namespace,
		logger:    conf.Logger,
		clientset: conf.Clientset,
	}
}

func (c *Informer) Run(ctx context.Context) error {
	c.logger.Info("Secret informer started")
	defer c.logger.Info("Secret informer stopped")

	c.ctx = ctx
	informerFactory := informers.NewSharedInformerFactoryWithOptions(c.clientset, 0, informers.WithNamespace(c.namespace))
	informer := informerFactory.Core().V1().Secrets().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	})

	informer.Run(ctx.Done())
	return nil
}

func (c *Informer) onAdd(obj interface{}) {
	f := obj.(*corev1.Secret)
	f = f.DeepCopy()

	c.logger.Info("onAdd", "secret", objref.KObj(f))
}

func (c *Informer) onUpdate(oldObj, newObj interface{}) {
	f := newObj.(*corev1.Secret)
	f = f.DeepCopy()
	c.logger.Info("onUpdate",
		"secret", objref.KObj(f),
	)
}

func (c *Informer) onDelete(obj interface{}) {
	f := obj.(*corev1.Secret)
	f = f.DeepCopy()
	c.logger.Info("onDelete",
		"secret", objref.KObj(f),
	)
}

func (c *Informer) Sync() {
	//c.logger.Info("During this time missing resource could be created on cluster")
}
