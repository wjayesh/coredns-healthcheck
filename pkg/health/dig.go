package health

import (
	"math"
	"os/exec"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

var (
	ts         []time.Time // An int slice to hold a fixed number of timestaps when the pods had failed
	deployment string      // Name of deployment
	memFactor  int         // The factor by which to increase memory limit
)

// variables for instrumentation
var (
	dnsQueryCount float64 // Metric counting the number of time a DNS query was performed.
	respTime      = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "dns_query_response_time",
		Help:    "The time it takes to finish dns queries",
		Buckets: prometheus.DefBuckets,
	}) // prometheus histogram to measure dns res times
)

// Dig calls the q executable with arg ip
// this functionality was not implemented in-line in main because
// we might change the working later depending on best practices.
func Dig(ip string) (string, error) {
	// using the k8s service to check DNS availability
	logrus.Info("IP address being queried: ", ip)
	before := time.Now()
	cmd := exec.Command("./q", "@"+ip, "kubernetes.default.svc.cluster.local")
	out, err := cmd.CombinedOutput()
	after := time.Now()

	// calculating the time taken to get response
	duration := float64(after.Nanosecond()-before.Nanosecond()) / math.Pow(10, 6)
	// adding value to histogram
	respTime.Observe(duration)

	logrus.Info("Output after executing q: ", string(out))

	if err != nil {
		// the issue is likely to be non ip specific
		// thus we are not passing ip info with the error
		return "", err
	}

	// incrementing metric
	dnsQueryCount = dnsQueryCount + 1

	output := string(out)
	return output, nil
}

// IsValidOutput checks the output string to determine if
// the output is a valid DNS response
func IsValidOutput(out string) bool {
	if strings.Contains(out, "i/o timeout") {
		logrus.Info("I/O Timeout detected.")
		ts = append(ts, time.Now())
		if len(ts) > 12 {
			ts = ts[len(ts)-13:]
		}
		return false
	} else if !strings.Contains(out, "NOERROR") {
		logrus.Info("Status code not equal to NOERROR.")
		ts = append(ts, time.Now())
		if len(ts) > 12 {
			ts = ts[len(ts)-13:]
		}
		return false
	}
	logrus.Info("DNS response is valid. Remedying of pods not needed.")
	logrus.Info("Timestamp array: ", ts)
	return true
}

// DigIPs performs queries on kubernetes service running on the default namespace
// using coreDNS pods and svcs
func DigIPs(client *kubernetes.Clientset, dn string, mf int, remedy bool, IPs map[string][]string) bool {
	deployment = dn
	memFactor = mf
	var success = true
	// TODO: Instead of if statements, implement labels
	if len(IPs["Pod IPs"]) != 0 {
		podIPs := IPs["Pod IPs"]
		for _, ip := range podIPs {
			if ip == "" {
				continue
			}
			out, err := Dig(ip)
			if err != nil {
				logrus.Error(err)
			} else {
				if !IsValidOutput(out) {
					logrus.Info("No DNS response from IP Addr: ", ip)
					success = false

					// we remedy pods only if this bool is true
					if remedy != true {
						continue
					}

					logrus.Info("Remedying Pod...")
					RemedyPod(client, namespace, ts, ip)
				} else {
					logrus.Info("DNS response from IP Addr: ", ip, out)
				}
			}
		}
	}

	// Now check Service IPs
	if len(IPs["Service IPs"]) != 0 {
		serviceIPs := IPs["Service IPs"]
		for _, ip := range serviceIPs {
			if ip == "" {
				continue
			}
			out, err := Dig(ip)
			if err != nil {
				logrus.Error(err)
			} else {
				if !IsValidOutput(out) {
					logrus.Info("No DNS response from Service IP Addr: ", ip)
					success = false

					// we remedy pods only if this bool is true
					if remedy != true {
						continue
					}

					logrus.Info("Remedying all service pods...")
					RemedyPod(client, namespace, ts, IPs["Pod IPs"]...)
				} else {
					logrus.Info("DNS response from Service IP Addr: ", ip, out)
				}
			}
		}
	}
	return success
}

// GetDNSMetrics exports metric variables to the collector function
func GetDNSMetrics() (queries float64, respTime *prometheus.Histogram) {
	return dnsQueryCount, respTime
}
