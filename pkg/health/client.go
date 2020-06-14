// Package health has functions that help with connecting to the api-server, looking up pods and
// services, performing dns queries on them and fixing failed deployments.
package health

import (
	"github.com/sirupsen/logrus"
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
