package main

import (
	"flag"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"github.com/WJayesh/healthCheck/utils"
)

var namespace string = "kube-dns"
var svcName string = "kube-dns"

var pathToCfg = flag.String("path", "", "the path to the kubeconfig file")
var podsAllowed = flag.Bool("allowPods", false, "allow creation of lightweight pods in cluster")

func main() {
	flag.Parse()
	var client = connectToAPIServer()
	var IPs = findIPs(client)
	if *pathToCfg == "" {
		//digIPs(IPs, nil)
	}
	if *pathToCfg != "" && *podsAllowed == true {
		//createPod()
		//digIPs(IPs, pod)
	}
	if *pathToCfg != "" && *podsAllowed == false {
		//var service = GetService(nil, nil, udpPort, client)
		//digIPs(service.ExternalIPs, nil)
	}
}

// The user needs to have a service account
func connectToAPIServer() *kubernetes.Clientset {
	var client, err = utils.GetClient(*pathToCfg)
	if client == nil {
		logrus.Error("Client not found: ", err)
		return nil
	}
	logrus.Info("Client received: ", client.LegacyPrefix)
	return client
}

// findIPs will return a map of IP addresses grouped by Service and Pods
// These IP addresses will be used by the application when it's running inside
// the cluster
func findIPs(client *kubernetes.Clientset) map[string][]string {
	var svc, err = utils.GetService(svcName, namespace, -1, client)
	var groupedIPs map[string][]string
	if err == nil {
		a := make([]string, 1)
		groupedIPs["Service IP"] = append(a, svc.Spec.ClusterIP)
	}
	//var pods, e = GetPods(svc, namespace, client)
	return groupedIPs
}
