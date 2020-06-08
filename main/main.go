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
	logrus.Info("Service IPs: ", IPs["Service IPs"])
	logrus.Info("Pod IPs: ", IPs["Pod IPs"])
	// TODO: Check if the number of pod ips in map match the
	// number of coreDNS pods
	if *pathToCfg == "" {
		digIPs(client, IPs)
	}
	if *pathToCfg != "" && *podsAllowed == true {
		// createPod()
		// exit program
	}
	if *pathToCfg != "" && *podsAllowed == false {
		port := int32(*udpPort)
		var service, err = utils.GetService("", "", port, client)
		if err != nil {
			logrus.Error(err)
		} else {
			// Convert ExternalIPs (which is a slice of strings) to a map
			// this is so done that digIPs method can know that these are svc IPs
			logrus.Info("External service discovered: ", service.Name)
			IPs := make(map[string][]string)
			IPs["Service IPs"] = make([]string, 1)
			IPs["Service IPs"] = append(IPs["Service IPs"], service.Spec.ExternalIPs...)
			digIPs(client, IPs)
		}

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
	groupedIPs = make(map[string][]string)
	if err == nil {
		a := make([]string, 1)
		groupedIPs["Service IPs"] = append(a, svc.Spec.ClusterIP)
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
	} else {
		logrus.Error(err)
	}
	return groupedIPs
}

func digIPs(client *kubernetes.Clientset, IPs map[string][]string) {
	// TODO: Instead of if statements, implement labels
	if IPs["Pod IPs"] != nil {
		podIPs := IPs["Pod IPs"]
		for _, ip := range podIPs {
			out, err := utils.Dig(ip)
			if err != nil {
				logrus.Error(err)
			} else {
				if !utils.IsValidOutput(out) {
					logrus.Info("No DNS response from IP Addr: ", ip)
					logrus.Info("Restarting Pod...")
					utils.RestartPod(client, ip)
				} else {
					logrus.Info("DNS response from IP Addr: ", ip, out)
				}
			}
		}
	}

	// Now check Service IPs
	if IPs["Service IPs"] != nil {
		serviceIPs := IPs["Service IPs"]
		for _, ip := range serviceIPs {
			out, err := utils.Dig(ip)
			if err != nil {
				logrus.Error(err)
			} else {
				if !utils.IsValidOutput(out) {
					logrus.Info("No DNS response from Service IP Addr: ", ip)
					logrus.Info("Restarting all service pods...")
					utils.RestartPod(client, namespace, IPs["Pod IPs"]...)
				} else {
					logrus.Info("DNS response from Service IP Addr: ", ip, out)
				}
			}
		}
	}
}
