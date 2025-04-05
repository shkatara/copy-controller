package controller

import (
	"context"
	"time"

	types "github.com/shkatara/copy-controller/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

func RunController(clientset *kubernetes.Clientset, ctx context.Context, cc types.ControllerConfig) {
	for {
		select {
		case <-time.After(cc.Interval):
			configMapList, err := clientset.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{
				LabelSelector: "trivago.com/copy=true",
			})

			if err != nil {
				klog.Infof("Did not find any configmap %s", err)
			}
			for _, configMap := range configMapList.Items {
				klog.Infof("Found configmap %s", configMap.Name)
			}
		case <-ctx.Done():
			return
		}
	}
}
