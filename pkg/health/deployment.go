package health

import (
	"context"

	"github.com/sirupsen/logrus"
	mv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodsReady checks if the new pods created after the remedial action are up and running
func PodsReady() bool {
	logrus.Info("DClient in PodsReady: ", dClient)
	var d, err = dClient.Get(context.TODO(), deployment, mv1.GetOptions{})
	if err != nil {
		logrus.Error("Error getting deployment: ", err)
		return false
	}
	// the number of updated pods and the total no of pods both should be two.
	// this solves the problem when >replicas pods (some with status other than ready)
	// are present and the func still returns true.
	if d.Status.UpdatedReplicas == int32(replicas) &&
		d.Status.Replicas == int32(replicas) {
		logrus.Info("All pods are ready")
		return true
	}
	logrus.Info("Pods aren't ready yet")
	return false
}
