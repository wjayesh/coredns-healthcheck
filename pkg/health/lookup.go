package health

import (
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

var (
	namespace string = "kube-system"
	svcName   string = "kube-dns"
)

// FindIPs will return a map of IP addresses grouped by Service and Pods
/** These IP addresses will be used by the application when it's running inside
the cluster
We take both Service IPs and Pod IPs to be pinged because
there it is possible that there are multiple point of failures.
On top of that, individual pods can be remedied.
*/
func FindIPs(client *kubernetes.Clientset) map[string][]string {

	// We'll first add the Service IP to the map.

	var svc, err = GetService(svcName, namespace, client)
	var groupedIPs map[string][]string
	groupedIPs = make(map[string][]string)
	if err == nil {
		a := make([]string, 1)
		groupedIPs["Service IPs"] = append(a, svc.Spec.ClusterIP)
	} else {
		logrus.Error(err)
	}

	// Now, we will add the IP addresses of the pods that are served by svc

	var pods, e = GetPods(svc, namespace, client)
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
