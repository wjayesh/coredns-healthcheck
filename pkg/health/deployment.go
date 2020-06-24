package health

import (
	"context"

	"github.com/sirupsen/logrus"
	mv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
)

// PodsReady checks if the new pods created after the remedial action are up and running
func PodsReady(dClient v1beta1.DeploymentInterface) bool {
	var d, err = dClient.Get(context.TODO(), deployment, mv1.GetOptions{})
	if err != nil {
		logrus.Error("Error getting deployment: ", err.Error)
		return false
	}
	// TODO add number of pods var instead of hardcoding
	if d.Status.UpdatedReplicas == 2 {
		logrus.Info("All pods are ready")
		return true
	}
	logrus.Info("Pods aren't ready yet")
	return false
}
