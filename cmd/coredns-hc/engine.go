package main

import (
	"strconv"

	"github.com/WJayesh/health-check/pkg/health"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type engine struct {
	path        string
	podsAllowed bool
	port        int
	client      *kubernetes.Clientset
}

func (e engine) Init(path string) {
	var err error
	e.client, err = health.GetClient(e.path)
	if e.client == nil {
		logrus.Error("Client not found: ", err)
	}
	logrus.Info("Client received: ", e.client.LegacyPrefix)
}

func (e engine) SetOptions(prefs map[string]string) {
	podsAllowed, err := strconv.ParseBool(prefs["podsAllowed"])
	if err == nil {
		e.podsAllowed = podsAllowed
	}
	port, err := strconv.Atoi(prefs["port"])
	if err == nil {
		e.port = port
	}

}

func (e engine) Start() {
	var IPs = health.FindIPs(e.client)
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
