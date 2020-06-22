package health

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	mv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

// GetPods will return a PodList of the pods served by the service svc
// Possible package name: pods. so that func call becomes pods.GetByService
// and pods.GetByIP (which can be used by pods.Restart())
func GetPods(svc *v1.Service, namespace string,
	client *kubernetes.Clientset) (*v1.PodList, error) {

	set := labels.Set(svc.Spec.Selector)
	// preparing a listOptions with the selector from the service
	listOptions := mv1.ListOptions{LabelSelector: set.AsSelector().String()}

	//using the API to get a PodList that satisfies the selector value
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), listOptions)
	if err == nil {
		return pods, err
	}
	return nil, errors.New("No Pods found for service" + svc.Name)
}

//RestartPod restarts the coredns pods with matching ips
// for this purpose, we only need to delete the pods;
// The deployment controller will create new pods automatically
func RestartPod(client *kubernetes.Clientset, namespace string, ips ...string) {
	// Get pods from IPs
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), mv1.ListOptions{})
	if err != nil {
		logrus.Error("Error listing all pods", err)
	} else {
		for _, pod := range pods.Items {
			for _, ip := range ips {
				if pod.Status.PodIP == ip {
					logrus.Info("Pod to be deleted: ", pod.Name)
					err := client.CoreV1().Pods(namespace).Delete(context.TODO(), pod.Name, mv1.DeleteOptions{})
					logrus.Error("Error deleting pod: ", err)
				}
			}
		}
	}
}
