package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	informer "k8s.io/client-go/informers"

	error "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type configMap struct {
	Name      string
	Namespace string
	Data      map[string]string
}

var configMapList []configMap

func NewConfigMap() *configMap {
	return &configMap{
		Name:      "",
		Data:      make(map[string]string),
		Namespace: "",
	}
}

func (cm *configMap) Run(client *kubernetes.Clientset, ctx context.Context) {
	// create a shared informer factory
	factory := informer.NewSharedInformerFactoryWithOptions(client, 0, informer.WithNamespace(v1.NamespaceAll),
		informer.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = "trivago.com/copy=true"
		}))
	// Create a ConfigMap informer
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
	log.Println("Added ConfigMap:", cmObj.Name)
	// Append to configMapList list
	configMapList = append(configMapList, configMap{
		Name:      cmObj.Name,
		Data:      cmObj.Data,
		Namespace: cmObj.Namespace,
	})
}

// func (cm *configMap) DeleteFunc(obj interface{}) {
// 	configMap := obj.(*v1.ConfigMap)
// 	// Handle the add event
// 	log.Println("Deleted ConfigMap: ", configMap.Name)
// 	// Add the configMap name to the map
// 	delete(cm.configMapNames, configMap.Name)
// }

func (cm *configMap) FinalConfigMaps() []configMap {
	mapList := make([]configMap, 0)
	for _, name := range configMapList {
		mapList = append(mapList, configMap{
			Name:      name.Name,
			Data:      name.Data,
			Namespace: name.Namespace,
		})
	}
	return mapList
}

func (cm *configMap) CopyConfigMaps(clientset *kubernetes.Clientset) {
	fmt.Print("Copying ConfigMaps \n")
	cmObj := &v1.ConfigMap{}
	for {

		if len(configMapList) == 0 {
			log.Println("No ConfigMaps found")
		}

		for _, name := range cm.FinalConfigMaps() {
			cmObj.Name = name.Name
			cmObj.Data = name.Data
			cmObj.Namespace = name.Namespace
			_, err := clientset.CoreV1().ConfigMaps(name.Namespace).Create(context.TODO(), cmObj, metav1.CreateOptions{})
			if error.IsAlreadyExists(err) {
				log.Println(cmObj.Name, "is already updated.")
			}
		}
		time.Sleep(5 * time.Second)

	}
}

// for _, name := range cm.FinalConfigMaps() {

func returnkubernetesclientset(filePath *string) *kubernetes.Clientset {
	config, _ := clientcmd.BuildConfigFromFlags("", *filePath)
	clientset, _ := kubernetes.NewForConfig(config)
	return clientset
}

func main() {
	sk := flag.String("sk", "", "absolute path to the sk file")
	dk := flag.String("dk", "", "absolute path to the dk file")

	flag.Parse()

	sourceClientSet := returnkubernetesclientset(sk)
	destClientSet := returnkubernetesclientset(dk)
	// create a context
	ctx, _ := context.WithCancel(context.Background())

	cm := NewConfigMap()
	fmt.Println("Starting ConfigMap controller")
	// go cm.PrintConfigMaps()
	go cm.CopyConfigMaps(destClientSet)
	cm.Run(sourceClientSet, ctx)
}
