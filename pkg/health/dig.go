package health

import (
	"os/exec"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

var (
	ts         []time.Time // An int slice to hold a fixed number of timestaps when the pods had failed
	deployment string      // Name of deployment
	memFactor  int         // The factor by which to increase memory limit
)

// Dig calls the q executable with arg ip
// this functionality was not implemented in-line in main because
// we might change the working later depending on best practices.
func Dig(ip string) (string, error) {
	// using the k8s service to check DNS availability
	logrus.Info("IP address being queried: ", ip)
	cmd := exec.Command("./q", "@"+ip, "kubernetes.default.svc.cluster.local")
	out, err := cmd.CombinedOutput()
	logrus.Info("Output after executing q: ", string(out))
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
	if strings.Contains(out, "i/o timeout") {
		logrus.Info("I/O Timeout detected.")
		ts = append(ts, time.Now())
		if len(ts) > 5 {
			ts = ts[len(ts)-6:]
		}
		return false
	} else if !strings.Contains(out, "NOERROR") {
		logrus.Info("Status code not equal to NOERROR.")
		ts = append(ts, time.Now())
		if len(ts) > 5 {
			ts = ts[len(ts)-6:]
		}
		return false
	}
	logrus.Info("DNS response is valid. Restarting of pods not needed.")
	logrus.Info("Timestamp array: ", ts)
	return true
}

// DigIPs performs queries on kubernetes service running on the default namespace
// using coreDNS pods and svcs
func DigIPs(client *kubernetes.Clientset, dn string, mf int, IPs map[string][]string) {
	deployment = dn
	memFactor = mf
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
					logrus.Info("Restarting Pod...")
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
					logrus.Info("Restarting all service pods...")
					RemedyPod(client, namespace, ts, IPs["Pod IPs"]...)
				} else {
					logrus.Info("DNS response from Service IP Addr: ", ip, out)
				}
			}
		}
	}
}
