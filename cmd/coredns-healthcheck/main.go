// This is a tool to monitor the health of the coredns deployment and perform automated
// remedies in case of a failure.
package main

import (
	"flag"

	"github.com/WJayesh/coredns-healthcheck/pkg/engine"
)

var (
	pathToCfg   = flag.String("path", "", "the path to the kubeconfig file")
	podsAllowed = flag.String("allowPods", "false", "allow creation of lightweight pods in cluster")
	udpPort     = flag.String("port", "53", "the udp port for the dns server")
	memFactor   = flag.String("memFactor", "2", "the factor with which to increase memory limit")
	replicas    = flag.String("replicas", "2", "the number of CoreDNS pods in deployment")
)

var (
	namespace  = "kube-system"
	svcName    = "kube-dns"
	deployment = "coredns"
)

func main() {
	flag.Parse()

	prefs := make(map[string]string)
	prefs["podsAllowed"] = *podsAllowed
	prefs["port"] = *udpPort
	prefs["namespace"] = namespace
	prefs["svcName"] = svcName
	prefs["deployment"] = deployment
	prefs["memFactor"] = *memFactor
	prefs["replicas"] = *replicas

	var e = engine.New(prefs)

	var client = e.Init(*pathToCfg)

	e.Start(client)
}
