// Package engine uses pkg/health to provide a quick way to start a health check
package engine

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/WJayesh/coredns-healthcheck/pkg/exporter"
	"github.com/WJayesh/coredns-healthcheck/pkg/health"
	"github.com/WJayesh/coredns-healthcheck/pkg/netns"
	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

// Engine is a structure to hold the details about the deployment mode
// and the pods to be tested.
type Engine struct {
	path        string                // path to the kubeconfig file
	podsAllowed bool                  // true, if creating pods allowed inside cluster
	port        int                   // the port of the service to be tested. default: 53
	namespace   string                // the namespace of the resource
	svcName     string                // the name of the service
	deployment  string                // the name of the deployment
	memFactor   int                   // the factor by which to increase memory
	replicas    int                   // no of CoreDNS pods in deployment
	client      *kubernetes.Clientset // the clientset
}

// New returns an Engine instance initialized with the
// supplied preferences
func New(prefs map[string]string) *Engine {
	var e Engine
	podsAllowed, err := strconv.ParseBool(prefs["podsAllowed"])
	if err == nil {
		e.podsAllowed = podsAllowed
	}
	port, err := strconv.Atoi(prefs["port"])
	if err == nil {
		e.port = port
	}
	mf, err := strconv.Atoi(prefs["memFactor"])
	if err == nil {
		e.memFactor = mf
	}
	replicas, err := strconv.Atoi(prefs["replicas"])
	if err == nil {
		e.replicas = replicas
	}
	e.svcName = prefs["svcName"]
	e.namespace = prefs["namespace"]
	e.deployment = prefs["deployment"]
	return &e
}

// Init connects the application to the cluster's api-server
func (e *Engine) Init(path string) *kubernetes.Clientset {

	// creating instances of metrics collectors
	remInstance := exporter.NewRemedyCollector()
	dnsInstance := exporter.NewDNSCollector()

	// registering metrics with prometheus client
	prometheus.MustRegister(remInstance)
	prometheus.MustRegister(dnsInstance)

	// helper function to begin listening on a separate goroutine
	startHandler()

	var err error

	// obataining the k8s clientset
	e.client, err = health.GetClient(e.path)
	if e.client == nil {
		logrus.Error("Client not found: ", err)
	}
	logrus.Info("Client received: ", e.client.LegacyPrefix)

	// initialzing the deployment client
	health.InitDClient(e.client, e.namespace)
	return e.client
}

// startHandler begins listening and serving asynchronously
func startHandler() {
	go func() {
		// start the HTTP server and expose
		// any metrics on the /metrics endpoint.

		http.Handle("/metrics", promhttp.Handler())
		logrus.Info("Beginning to serve on port :9890")
		logrus.Fatal(http.ListenAndServe(":9890", nil))
	}()
}

// Start runs the health check and checks for failures.
// It also attempts to fix any terminated pods.
func (e *Engine) Start(client *kubernetes.Clientset) {
Start:
	var IPs = health.FindIPs(e.namespace, e.svcName, e.replicas, client)
	// initiate first phase
	err := e.firstPhase(client, true, IPs)

	// if first phase terminated without errors, then perfrom the second phase of checks.
	if err == nil {
		e.secondPhase(client, IPs)
	}

	logrus.Info("using the client variable ", client.LegacyPrefix)
	for {
		time.Sleep(1 * time.Second)
		goto Start
		//infinite loop
	}
}

func (e *Engine) firstPhase(client *kubernetes.Clientset, remedy bool, IPs map[string][]string) error {
	var success bool

	logrus.Info("Service IPs: ", IPs["Service IPs"])
	logrus.Info("Pod IPs: ", IPs["Pod IPs"])
	// Check if the number of pod ips in map match the
	// number of coreDNS pods
	// Answer: There's no need to do that here. The case where
	// there's less number of IPs is highly improbable to happen.
	// This is because in GetPods func, we do not specify that we
	// are searching for just "Running" pods.
	if e.path == "" {
		success = health.DigIPs(client, e.deployment, e.memFactor, remedy, IPs)
	}
	if e.path != "" && e.podsAllowed == true {
		// createPod()
		// exit program
	}
	if e.path != "" && e.podsAllowed == false {
		udpPort := int32(e.port)
		var service, err = health.GetServiceByPort(udpPort, client)
		if err != nil {
			logrus.Error(err)
		} else {
			// Convert ExternalIPs (which is a slice of strings) to a map
			// this is so done that digIPs method can know that these are svc IPs
			logrus.Info("External service discovered: ", service.Name)
			IPs := make(map[string][]string)
			IPs["Service IPs"] = make([]string, 1)
			IPs["Service IPs"] = append(IPs["Service IPs"], service.Spec.ExternalIPs...)
			success = health.DigIPs(client, e.deployment, e.memFactor, remedy, IPs)
		}

	}
	if success == true {
		return nil
	}
	return errors.New("First Phase error detected")
}

func (e *Engine) secondPhase(client *kubernetes.Clientset, IPs map[string][]string) {
	list := netns.GetNetNS(client)
	for _, targetNS := range *list {
		if err := targetNS.Do(func(_ ns.NetNS) error {
			// inside a net ns
			logrus.Info("Inside Namespace with path: ", targetNS.Path())

			er := e.firstPhase(client, false, IPs)
			if er != nil {
				logrus.Info("Error querying the CoreDNS from namespace with path: ",
					targetNS.Path())
				logrus.Error("Error: ", er)

				netns.RemedyNS(&targetNS, e.svcName, e.namespace, e.client)
				// check if etc/namespace exists or not
				// this is because we can't access it directly as setns doesn't
				// relocate etc/namespace/.../ to etc/resolv.conf
			}

			logrus.Info("Successfully queried CoreDNS pods from namespace with path: ",
				targetNS.Path())
			return nil

		}); err != nil {
			logrus.Error("Error performing function inside namespace: ", err)
		} else {
			logrus.Info("Function executed successfully inside namespace")
		}
	}
}
