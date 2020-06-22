// Package health has functions that help with connecting to the api-server, looking up pods and
// services, performing dns queries on them and fixing failed deployments.
package health

import (
	"context"

	"github.com/sirupsen/logrus"
	mv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

var mc *metrics.Clientset

// GetClient returns a clientset and initializes the metrics client
func GetClient(pathToCfg string) (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error
	if pathToCfg == "" {
		logrus.Info("Using in cluster config; deployed as a pod")
		config, err = rest.InClusterConfig()
	} else {
		logrus.Info("Using out of cluster config; deployed externally")
		config, err = clientcmd.BuildConfigFromFlags("", pathToCfg)
	}
	if err != nil {
		return nil, err
	}
	mc, err = metrics.NewForConfig(config)
	if err != nil {
		logrus.Error("Error getting metrics client", err)
	} else {
		logrus.Info("Metrics client found: ", mc.LegacyPrefix)
	}
	return kubernetes.NewForConfig(config)
}

// GetMemory returns the memory limit of the container in the pod specified by the name param
func GetMemory(name string) int64 {

	var podMetrics, err = mc.MetricsV1alpha1().PodMetricses(namespace).Get(context.TODO(), name, mv1.GetOptions{})
	if err != nil {
		logrus.Error("Error getting metrics for pod: ", name)
		return -1
	}
	for _, container := range podMetrics.Containers {
		memory, ok := container.Usage.Memory().AsInt64()
		if !ok {
			logrus.Error("Error getting the memory usage of container")
		} else {
			return memory
		}
	}
	return -1
}
