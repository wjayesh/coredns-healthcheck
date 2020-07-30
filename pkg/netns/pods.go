package netns

import (
	"context"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	mv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ListPods will get all the pods deployed across all namespaces
func ListPods(client *kubernetes.Clientset) *[]v1.Pod {
	pods, err := client.CoreV1().Pods("").List(context.TODO(), mv1.ListOptions{})
	if err != nil {
		logrus.Error("Error retrieving pods (second phase): ", err)
		empty := make([]v1.Pod, 0)
		return &empty
	}
	logrus.Info("Pods returned: ", len(pods.Items))
	return &pods.Items
}
