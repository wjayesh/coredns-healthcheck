package health

import (
	"context"
	"strconv"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	mv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"
)

// GetMemory returns the memory limit of the container in the pod specified by the name param
func GetMemory(name string) int64 {

	var podMetrics, err = mClient.MetricsV1alpha1().PodMetricses(namespace).Get(context.TODO(), name, mv1.GetOptions{})
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

// AddMemory multiplies the existing memory limit of deployment by memFactor
func AddMemory(memFactor int, name string) {

	if memFactor < 1 {
		memFactor = 2
	}
	currMem := GetMemory(name)
	newMem := int(currMem) * memFactor

	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {

		result, err := dClient.Get(context.TODO(), deployment, mv1.GetOptions{})
		if err != nil {
			logrus.Error("Error getting deployment :", err)
		}
		var i = 0
		var updateErr error
		for i < 2 {
			result.Spec.Template.Spec.Containers[i].Resources.Limits =
				make(map[v1.ResourceName]resource.Quantity)

			result.Spec.Template.Spec.Containers[i].Resources.Limits[v1.ResourceMemory] =
				resource.MustParse(strconv.Itoa(newMem))

			_, updateErr = dClient.Update(context.TODO(), result, mv1.UpdateOptions{})
			i = i + 1
		}
		return updateErr
	})

	if retryErr != nil {
		logrus.Error("Retry on conflict fails: ", retryErr.Error)
	}

	// TODO
	// Sleep till all pods are running again (same in RestartPod)
}
