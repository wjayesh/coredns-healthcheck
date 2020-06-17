package health

import (
	"os/exec"
	"strings"

	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

// Dig calls the q executable with arg ip
// this functionality was not implemented in-line in main because
// we might change the working later depending on best practices.
func Dig(ip string) (string, error) {
	// using the k8s service to check DNS availability
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
		return false
	} else if !strings.Contains(out, "NOERROR.") {
		logrus.Info("Status code not equal to NOERROR.")
		return false
	}
	logrus.Info("DNS response is valid. Restarting of pods not needed.")
	return true
}

// DigIPs performs queries on kubernetes service running on the default namespace
// using coreDNS pods and svcs
func DigIPs(client *kubernetes.Clientset, IPs map[string][]string) {
	// TODO: Instead of if statements, implement labels
	if IPs["Pod IPs"] != nil {
		podIPs := IPs["Pod IPs"]
		for _, ip := range podIPs {
			out, err := Dig(ip)
			if err != nil {
				logrus.Error(err)
			} else {
				if !IsValidOutput(out) {
					logrus.Info("No DNS response from IP Addr: ", ip)
					logrus.Info("Restarting Pod...")
					RestartPod(client, ip)
				} else {
					logrus.Info("DNS response from IP Addr: ", ip, out)
				}
			}
		}
	}

	// Now check Service IPs
	if IPs["Service IPs"] != nil {
		serviceIPs := IPs["Service IPs"]
		for _, ip := range serviceIPs {
			out, err := Dig(ip)
			if err != nil {
				logrus.Error(err)
			} else {
				if !IsValidOutput(out) {
					logrus.Info("No DNS response from Service IP Addr: ", ip)
					logrus.Info("Restarting all service pods...")
					RestartPod(client, namespace, IPs["Pod IPs"]...)
				} else {
					logrus.Info("DNS response from Service IP Addr: ", ip, out)
				}
			}
		}
	}
}
