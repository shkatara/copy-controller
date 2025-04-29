package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	v1 "k8s.io/api/core/v1"
	errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	informer "k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

type configMap struct {
	Name      string
	Namespace string
	State     string
	Data      map[string]string
}

var configMapList []configMap

func NewConfigMap() *configMap {
	return &configMap{
		Name:      "",
		Data:      make(map[string]string),
		Namespace: "",
		State:     "",
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
		AddFunc:    cm.AddFunc,
		DeleteFunc: cm.DeleteFunc,
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
		State:     "add",
	})
}

func (cm *configMap) DeleteFunc(obj interface{}) {
	cmObj := obj.(*v1.ConfigMap)
	// Handle the add event
	log.Println("Deleted ConfigMap: ", cmObj.Name)
	// Change the state of the configMap to deleted
	for i, name := range configMapList {
		if name.Name == cmObj.Name {
			configMapList[i].State = "delete"
		}
	}
}

func (cm *configMap) FinalConfigMaps() []configMap {
	mapList := make([]configMap, 0)
	for _, name := range configMapList {
		mapList = append(mapList, configMap{
			Name:      name.Name,
			Data:      name.Data,
			Namespace: name.Namespace,
			State:     name.State,
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
			//fmt.Println("state is: ", name.State)
			// log.Println("Adding ConfigMap: ", cmObj.Name)
			// _, err := clientset.CoreV1().ConfigMaps(name.Namespace).Create(context.TODO(), cmObj, metav1.CreateOptions{})
			// if error.IsAlreadyExists(err) {
			// 	log.Println(cmObj.Name, "is already updated.")
			// }

			if name.State == "add" {
				log.Println("Adding ConfigMap: ", cmObj.Name)
				_, err := clientset.CoreV1().ConfigMaps(name.Namespace).Create(context.TODO(), cmObj, metav1.CreateOptions{})
				if errors.IsAlreadyExists(err) {
					log.Println(cmObj.Name, "is already updated.")
				}
			} else if name.State == "delete" {
				log.Println("Deleting ConfigMap: ", cmObj.Name)
				err := clientset.CoreV1().ConfigMaps(name.Namespace).Delete(context.TODO(), cmObj.Name, metav1.DeleteOptions{})
				if errors.IsNotFound(err) {
					log.Println(cmObj.Name, "is already deleted.")
				}
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
	go cm.CopyConfigMaps(destClientSet)
	cm.Run(sourceClientSet, ctx)
}
