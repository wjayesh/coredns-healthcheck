// This is a tool to monitor the health of the coredns deployment and perform automated
// remedies in case of a failure.

package main

import (
	"flag"

	"github.com/WJayesh/health-check/pkg/engine"
)

var (
	pathToCfg   = flag.String("path", "", "the path to the kubeconfig file")
	podsAllowed = flag.String("allowPods", "false", "allow creation of lightweight pods in cluster")
	udpPort     = flag.String("port", "53", "the udp port for the dns server")
)

var (
	namespace string = "kube-system"
	svcName   string = "kube-dns"
)

func main() {
	flag.Parse()

	prefs := make(map[string]string)
	prefs["podsAllowed"] = *podsAllowed
	prefs["port"] = *udpPort
	prefs["namespace"] = namespace
	prefs["svcName"] = svcName

	var e = engine.New(prefs)

	e.Init(*pathToCfg)

	e.Start()
}
