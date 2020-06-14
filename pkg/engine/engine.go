// Package engine uses pkg/health to provide a quick way to start a health check
package engine

import (
	"strconv"

	"github.com/WJayesh/health-check/pkg/health"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

// Engine is a structure to hold the details about the deployment mode
// and the pods to be tested.
type Engine struct {
	path        string                // path to the kubeconfig file
	podsAllowed bool                  // true, if creating pods allowed inside cluster
	port        int                   // the port of the service to be tested. default: 53
	namespace   string                // the namespace of the resource
	svcName     string                // the name of the service
	client      *kubernetes.Clientset // the clientset
}

// New returns an Engine instance initialized with the
// supplied preferences
func New(prefs map[string]string) Engine {
	var e Engine
	podsAllowed, err := strconv.ParseBool(prefs["podsAllowed"])
	if err == nil {
		e.podsAllowed = podsAllowed
	}
	port, err := strconv.Atoi(prefs["port"])
	if err == nil {
		e.port = port
	}
	e.svcName = prefs["svcName"]
	e.namespace = prefs["namespace"]
	return e
}

// Init connects the application to the cluster's api-server
func (e Engine) Init(path string) {
	var err error
	e.client, err = health.GetClient(e.path)
	if e.client == nil {
		logrus.Error("Client not found: ", err)
	}
	logrus.Info("Client received: ", e.client.LegacyPrefix)
}

// Start runs the health check and checks for failures.
// It also attempts to fix any terminated pods.
func (e Engine) Start() {
	var IPs = health.FindIPs(e.namespace, e.svcName, e.client)
	logrus.Info("Service IPs: ", IPs["Service IPs"])
	logrus.Info("Pod IPs: ", IPs["Pod IPs"])
	// TODO: Check if the number of pod ips in map match the
	// number of coreDNS pods
	if e.path == "" {
		health.DigIPs(e.client, IPs)
	}
	if e.path != "" && e.podsAllowed == true {
		// createPod()
		// exit program
	}
	if e.path != "" && e.podsAllowed == false {
		udpPort := int32(e.port)
		var service, err = health.GetServiceByPort(udpPort, e.client)
		if err != nil {
			logrus.Error(err)
		} else {
			// Convert ExternalIPs (which is a slice of strings) to a map
			// this is so done that digIPs method can know that these are svc IPs
			logrus.Info("External service discovered: ", service.Name)
			IPs := make(map[string][]string)
			IPs["Service IPs"] = make([]string, 1)
			IPs["Service IPs"] = append(IPs["Service IPs"], service.Spec.ExternalIPs...)
			health.DigIPs(e.client, IPs)
		}

	}
	logrus.Info("using the client variable ", e.client.LegacyPrefix)
	for {
		//infinite loop
	}
}
