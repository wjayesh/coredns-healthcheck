package main

import (
	"flag"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"github.com/WJayesh/healthCheck/utils"
)

var (
	namespace string = "kube-system"
	svcName   string = "kube-dns"
)

var (
	pathToCfg   = flag.String("path", "", "the path to the kubeconfig file")
	podsAllowed = flag.Bool("allowPods", false, "allow creation of lightweight pods in cluster")
	udpPort     = flag.Int("port", 53, "the udp port for the dns server")
)

func main() {
	flag.Parse()
	var client = connectToAPIServer()
	var IPs = findIPs(client)
	logrus.Info("Service IP: ", IPs["Service IP"])
	logrus.Info("Pod IPs: ", IPs["Pod IPs"])
	//var IPs = findIPs(client)
	if *pathToCfg == "" {
		//digIPs(IPs, nil)
	}
	if *pathToCfg != "" && *podsAllowed == true {
		//createPod()
		//digIPs(IPs, pod)
	}
	if *pathToCfg != "" && *podsAllowed == false {
		//var service, err = utils.GetService(nil, nil, *udpPort, client)
		//digIPs(service.ExternalIPs, nil)
	}
	logrus.Info("using the client variable ", client.LegacyPrefix)
	for {
		//infinite loop
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

/** findIPs will return a map of IP addresses grouped by Service and Pods
These IP addresses will be used by the application when it's running inside
the cluster
We take both Service IPs and Pod IPs to be pinged because
there it is possible that there are multiple point of failures.
On top of that, individual pods can be remedied.
*/
func findIPs(client *kubernetes.Clientset) map[string][]string {

	// We'll first add the Service IP to the map.

	var svc, err = utils.GetService(svcName, namespace, -1, client)
	var groupedIPs map[string][]string
	if err == nil {
		a := make([]string, 1)
		groupedIPs["Service IP"] = append(a, svc.Spec.ClusterIP)
	} else {
		logrus.Error(err)
	}

	// Now, we will add the IP addresses of the pods that are served by svc

	var pods, e = utils.GetPods(svc, namespace, client)
	if e == nil {
		// There are two pods for CoreDNS.
		// but shouldn't be hardcoded (TODO)
		groupedIPs["Pod IPs"] = make([]string, 2)
		for _, pod := range pods.Items {
			groupedIPs["Pod IPs"] = append(groupedIPs["Pod IPs"], pod.Status.PodIP)
		}
	}
	return groupedIPs
}
