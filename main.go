package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	informer "k8s.io/client-go/informers"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type configMap struct {
	Name string
	Data map[string]string
}

var configMapList []configMap

func NewConfigMap() *configMap {
	return &configMap{
		Name: "",
		Data: make(map[string]string),
	}
}

func (cm *configMap) Run(client *kubernetes.Clientset, ctx context.Context) {
	// create a shared informer factory
	//factory := informer.NewSharedInformerFactory(client, 0)
	factory := informer.NewSharedInformerFactoryWithOptions(client, 0, informer.WithNamespace(v1.NamespaceAll),
		informer.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = "trivago.com/copy=true"
		}))
	// Create a ConfigMap informer
	// This will create a ConfigMap informer that watches all namespaces and caches the ConfigMaps in memory.
	cmInformer := factory.Core().V1().ConfigMaps().Informer()

	// Add event handlers to the informer on which we should act.
	cmInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: cm.AddFunc,
		//	DeleteFunc: cm.DeleteFunc,
	})
	// Start the informer
	factory.Start(ctx.Done())
	// Wait for the caches to be synced before starting the controller
	if !cache.WaitForCacheSync(ctx.Done(), cmInformer.HasSynced) {
		panic("failed to sync caches")
	} else {
		log.Println("Caches synced")
	}
	// Wait for the context to be done
	<-ctx.Done()
}

func (cm *configMap) AddFunc(obj interface{}) {
	cmObj := obj.(*v1.ConfigMap)
	// Handle the add event
	log.Println("Added ConfigMap: ", cmObj.Name)
	// Append to configMapList list
	configMapList = append(configMapList, configMap{
		Name: cmObj.Name,
		Data: cmObj.Data,
	})
}

// func (cm *configMap) DeleteFunc(obj interface{}) {
// 	configMap := obj.(*v1.ConfigMap)
// 	// Handle the add event
// 	log.Println("Deleted ConfigMap: ", configMap.Name)
// 	// Add the configMap name to the map
// 	delete(cm.configMapNames, configMap.Name)
// }

func (cm *configMap) ReadConfigMapNames() []configMap {
	mapList := make([]configMap, 0)
	for _, name := range configMapList {
		mapList = append(mapList, configMap{
			Name: name.Name,
			Data: name.Data,
		})
	}
	return mapList
}

func (cm *configMap) PrintConfigMaps() {
	for {
		if len(configMapList) != 0 {
			for _, name := range cm.ReadConfigMapNames() {
				log.Println("Name ", name.Name)
				log.Println("Data: ", name.Data)
			}
		} else {
			log.Println("No ConfigMaps found")
		}
		time.Sleep(5 * time.Second)
	}

}

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	// create a context
	ctx, _ := context.WithCancel(context.Background())

	cm := NewConfigMap()
	fmt.Println("Starting ConfigMap controller")
	go cm.PrintConfigMaps()
	cm.Run(clientset, ctx)

}
