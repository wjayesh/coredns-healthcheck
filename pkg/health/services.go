package health

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	mv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetService returns a Service struct using the given parameters
func GetService() (*v1.Service, error) {

	logrus.Info("Client received: ", client.LegacyPrefix)
	var svc, err = client.CoreV1().Services(namespace).Get(context.TODO(), svcName, mv1.GetOptions{})
	if err != nil {
		// exit
		logrus.Fatal(err)
	}
	if svc != nil {
		// no errors
		return svc, nil
	}
	return nil, errors.New("In-Cluster service not found")
}

// GetServiceByPort uses the port in the params to find
// external services exposing that port.
func GetServiceByPort(port int32,
	client *kubernetes.Clientset) (*v1.Service, error) {

	svcs, err := client.CoreV1().Services(namespace).List(context.TODO(), mv1.ListOptions{})
	if err != nil {
		logrus.Fatal(err)
	}
	for _, svc := range svcs.Items {
		// search among all services
		for _, svcPort := range svc.Spec.Ports {
			// only select services that serve the port and are exposed
			// to the outside world
			if svcPort.Port == port && svc.Spec.ExternalIPs != nil {
				return &svc, nil
			}
		}
	}
	return nil, errors.New("No external services available")
}
