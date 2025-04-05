package controller

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func CreateInClusterKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func CreateExternalKubernetesCluent(k string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.BuildConfigFromFlags("", k)
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}
