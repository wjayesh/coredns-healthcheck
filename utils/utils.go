package utils

import (
	"errors"
	"os/exec"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	mv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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
// If the port is not -1, we look for any service exposing the port
// to the outside world.
func GetService(name string, namespace string, port int32,
	client *kubernetes.Clientset) (*v1.Service, error) {
	if port == -1 {
		var svc, err = client.CoreV1().Services(namespace).Get(name, mv1.GetOptions{})
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
	// find services that expose the given port
	svcs, err := client.CoreV1().Services(namespace).List(mv1.ListOptions{})
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

// GetPods will return a PodList of the pods served by the service svc
// Possible package name: pods. so that func call becomes pods.GetByService
// and pods.GetByIP (which can be used by pods.Restart())
func GetPods(svc *v1.Service, namespace string,
	client *kubernetes.Clientset) (*v1.PodList, error) {

	set := labels.Set(svc.Spec.Selector)
	// preparing a listOptions with the selector from the service
	listOptions := mv1.ListOptions{LabelSelector: set.AsSelector().String()}

	//using the API to get a PodList that satisfies the selector value
	pods, err := client.CoreV1().Pods(namespace).List(listOptions)
	if err == nil {
		return pods, err
	}
	return nil, errors.New("No Pods found for service" + svc.Name)
}

// Dig calls the q executable with arg ip
// this functionality was not implemented in-line in main because
// we might change the working later depending on best practices.
func Dig(ip string) (string, error) {
	out, err := exec.Command("./q", ip).Output()
	if err != nil {
		// the issue is likely to be non ip specific
		// thus we are not passing ip info with the error
		return "", err
	}
	output := string(out)
	return output, nil
}

// IsValidOutput checks the output string to determine if
// the output is a valid DNS response
func IsValidOutput(out string) bool {
	return false
}

//RestartPod restarts the coredns pods with matching ips
// for this purpose, we only need to delete the pods;
// The deployment controller will create new pods automatically
func RestartPod(client *kubernetes.Clientset, namespace string, ips ...string) {
	// Get pods from IPs
	pods, err := client.CoreV1().Pods(namespace).List(mv1.ListOptions{})
	if err != nil {
		logrus.Error("Error listing all pods", err)
	} else {
		for _, pod := range pods.Items {
			for _, ip := range ips {
				if pod.Status.PodIP == ip {
					logrus.Info("Pod to be deleted: ", pod.Name)
					err := client.CoreV1().Pods(namespace).Delete(pod.Name, &mv1.DeleteOptions{})
					logrus.Error("Error deleting pod: ", err)
				}
			}
		}
	}
}
