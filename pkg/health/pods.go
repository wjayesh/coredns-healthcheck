package health

import (
	"context"
	"errors"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	mv1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

var (
	oomCount     float64
	restartCount float64
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

// RemedyPod chooses the right way to bring failed pods back to life
func RemedyPod(client *kubernetes.Clientset, namespace string, ts []time.Time, ips ...string) {
	// Get pods from IPs
	logrus.Info("Inside RemedyPod with ips: ", ips)
	pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), mv1.ListOptions{})
	if err != nil {
		logrus.Error("Error listing all pods", err)
	} else {
		for _, ip := range ips {
			for _, pod := range pods.Items {
				logrus.Info("ip being matched: ", ip)
				logrus.Info("pod.Status.PodIP = ", pod.Status.PodIP)
				if pod.Status.PodIP == ip {
					if IsOutOfMemory(ts) {
						oomCount = oomCount + 1
						AddMemory(memFactor, pod.Name)
						return
					}
					logrus.Info("Restarting Pod")
					restartCount = restartCount + 1
					RestartPod(pod)
				}
			}
		}
	}
}

// RestartPod restarts the coredns pods.
// For this purpose, we only need to delete the pods;
// The deployment controller will create new pods automatically
func RestartPod(pod v1.Pod) {
	logrus.Info("Pod to be deleted: ", pod.Name)
	err := client.CoreV1().Pods(namespace).Delete(context.TODO(), pod.Name, mv1.DeleteOptions{})
	logrus.Error("Error deleting pod: ", err)

	// No need to sleep here till all pods are running again.
	// this is taken care of in lookup.go (checking if all pods are running).
}

// GetRemedyMetrics returns
// 1) the number of oom errors
// 2) number of restarts performed
// 3) total number of errors
func GetRemedyMetrics() (oom float64, restart float64, total float64) {
	return oomCount, restartCount, oomCount + restartCount
}
