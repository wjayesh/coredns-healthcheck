// Package health has functions that help with connecting to the api-server, looking up pods and
// services, performing dns queries on them and fixing failed deployments.
package health

import (
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/typed/extensions/v1beta1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

var (
	mClient *metrics.Clientset
	dClient v1beta1.DeploymentInterface
)

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

	// metrics client
	mClient, err = metrics.NewForConfig(config)
	if err != nil {
		logrus.Error("Error getting metrics client", err)
	} else {
		logrus.Info("Metrics client found: ", mClient.LegacyPrefix)
	}

	return kubernetes.NewForConfig(config)
}

// InitDClient initializes the deployment client for future use.
func InitDClient(client *kubernetes.Clientset, ns string) {
	// deployment client
	logrus.Info("namespace used for dClient: ", ns)
	dClient = client.ExtensionsV1beta1().Deployments(ns)
	logrus.Info("Deployment Client: ", dClient)
	namespace = ns
}
