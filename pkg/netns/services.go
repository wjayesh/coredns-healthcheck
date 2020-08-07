package netns

import (
	"context"

	"github.com/sirupsen/logrus"
	mv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetServiceIP returns the ip address associated with the svc name
// namespace is the k8s namespace
func GetServiceIP(name string, namespace string, client *kubernetes.Clientset) string {
	logrus.Info("Client received: ", client.LegacyPrefix)
	var svc, err = client.CoreV1().Services(namespace).Get(context.TODO(), svcName, mv1.GetOptions{})
	if err != nil {
		// exit
		logrus.Fatal(err)
	}
	if svc != nil {
		// no errors
		return svc.Spec.ClusterIP
	}
	return ""
}
