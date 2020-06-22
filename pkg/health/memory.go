package health

import (
	"context"

	"github.com/sirupsen/logrus"
	mv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

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
