package main

import (
	"flag"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"

	"github.com/WJayesh/healthCheck/utils"
)

var pathToCfg = flag.String("path", "", "the path to the kubeconfig file")
var podsAllowed = flag.Bool("allowPods", false, "allow creation of lightweight pods in cluster")

func main() {
	flag.Parse()
	connectToAPIServer()
}

func connectToAPIServer() *kubernetes.Clientset {
	var client, err = utils.GetClient(*pathToCfg)
	if client == nil {
		logrus.Error("Client not found: ", err)
		return nil
	} else {
		return client
	}
}
