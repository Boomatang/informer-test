package configmap

import (
	"context"
	"fmt"
	"github.com/boomatang/informers-test/pkg/utils"
	"github.com/boomatang/informers-test/pkg/utils/objref"
	"github.com/go-logr/logr"
	"io/fs"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"path/filepath"
	"strings"
	"sync"
)

type InformerConfig struct {
	Logger    logr.Logger
	Namespace string
	Clientset *kubernetes.Clientset
	Required  map[string]string
}

type Informer struct {
	ctx       context.Context
	mut       sync.RWMutex
	namespace string
	logger    logr.Logger
	clientset *kubernetes.Clientset
	required  map[string]string
	has       []string
}

func NewConfigMapInformer(conf *InformerConfig) *Informer {
	return &Informer{
		namespace: conf.Namespace,
		logger:    conf.Logger,
		clientset: conf.Clientset,
		required:  conf.Required,
	}
}

func (c *Informer) Run(ctx context.Context) error {
	c.logger.Info("ConfigMap informer started")
	defer c.logger.Info("ConfigMap informer stopped")

	c.ctx = ctx
	informerFactory := informers.NewSharedInformerFactoryWithOptions(c.clientset, 0, informers.WithNamespace(c.namespace))
	informer := informerFactory.Core().V1().ConfigMaps().Informer()
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	})

	informer.Run(ctx.Done())
	return nil
}

func (c *Informer) onAdd(obj interface{}) {
	f := obj.(*corev1.ConfigMap)
	f = f.DeepCopy()

	c.logger.Info("OnAdd", "configmap", objref.KObj(f))
	a := fmt.Sprintf("%s", objref.KObj(f))
	c.has = append(c.has, a)

	//If the resource was in c.required then
	//	A validation step would be added here to ensure the resource is in the correct state
	// 	The obj is a reference to the resource on cluster. There is no need to re-fetch the resource
	//	If not in correct state then
	//		Update resource to correct state
}

func (c *Informer) onUpdate(oldObj, newObj interface{}) {
	f := newObj.(*corev1.ConfigMap)
	f = f.DeepCopy()
	c.logger.Info("onUpdate",
		"configmap", objref.KObj(f),
	)

	//If the resource was in c.required then
	//	A validation step would be added here to ensure the resource is in the correct state
	// 	The newObj is a reference to the resource on cluster. There is no need to re-fetch the resource
	//	If not in correct state then
	//		Update resource to correct state
}

func (c *Informer) onDelete(obj interface{}) {
	f := obj.(*corev1.ConfigMap)
	f = f.DeepCopy()
	configmap := fmt.Sprint(objref.KObj(f))
	c.logger.Info("onDelete",
		"configmap", configmap,
	)

	s := make([]string, len(c.has))
	copy(s, c.has)

	for index, value := range s {
		if value == configmap {
			c.has = append(s[:index], s[index+1:]...)
		}
	}

	_, ok := c.required[configmap]
	if ok {
		c.create(configmap)
	}

}

func (c *Informer) Sync() {
	for key := range c.required {
		if !utils.ElementExist(c.has, key) {
			c.logger.Info("Missing ConfigMap from cluster", "configmap", key)
			c.create(key)
		}
	}

}

func (c *Informer) create(key string) {
	c.logger.Info("creating", "configmap", key)

	d, err := ioutil.ReadFile(c.required[key])
	if err != nil {
		c.logger.Error(err, "not able to read file")
	}

	a := &corev1.ConfigMap{}
	err = yaml.Unmarshal(d, a)
	if err != nil {
		c.logger.Error(err, "not able to parse yaml")
	}

	_, err = c.clientset.CoreV1().ConfigMaps(c.namespace).Create(c.ctx, a, metav1.CreateOptions{})
	if err != nil {
		return
	}

}

func LoadData() (map[string]string, error) {
	data := map[string]string{}

	err := filepath.Walk("assets", func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), ".yaml") {
			d, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			a := &corev1.ConfigMap{}
			err = yaml.Unmarshal(d, a)
			if err != nil {
				return err
			}
			if a.TypeMeta.Kind == "ConfigMap" {
				data[fmt.Sprintf("default/%s", a.Name)] = path
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return data, nil
}
