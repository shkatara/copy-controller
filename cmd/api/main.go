package main

import (
	"context"
	"flag"
	"time"

	controller "github.com/shkatara/copy-controller/internal/server/controller"
	types "github.com/shkatara/copy-controller/types"

	"k8s.io/klog"
)

func main() {
	ctx := context.Background()
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")

	flag.StringVar(&types.Interval, "interval", "30s", "Interval to scrape directories (e.g. 30s, 1m, 1h)")
	flag.Parse()

	parsedInterval, err := time.ParseDuration(types.Interval)

	clientset, err := controller.CreateExternalKubernetesCluent(*kubeconfig)
	if err != nil {
		klog.Fatalf("Failed to create cient: %s", err)
	}

	cc := types.ControllerConfig{
		Interval: parsedInterval,
	}
	go controller.RunController(clientset, ctx, cc)

}
