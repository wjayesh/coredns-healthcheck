package utils

import (
	"errors"
	"log"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	mv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetClient returns a clientset
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
	return kubernetes.NewForConfig(config)
}

// GetService returns the Service with the given parameters
// If the port is not nil, we look for any service exposing the port
// to the outside world.
func GetService(name string, namespace string, port int32,
	client *kubernetes.Clientset) (*v1.Service, error) {
	if port == -1 {
		var svc, err = client.CoreV1().Services(namespace).Get(name, mv1.GetOptions{})
		if err != nil {
			log.Fatal(err)
		}
		if svc != nil {
			return svc, nil
		} else {
			return nil, errors.New("In-Cluster service not found")
		}
	} else {
		// find services that expose the given port
		svcs, err := client.CoreV1().Services(namespace).List(mv1.ListOptions{})
		if err != nil {
			log.Fatal(err)
		}
		for _, svc := range svcs.Items {
			for _, svcPort := range svc.Spec.Ports {
				if svcPort.Port == port {
					return &svc, nil
				}
			}
		}
		return nil, errors.New("No external services available")
	}
}
