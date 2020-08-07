package netns

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

var (
	svcName string
)

// RemedyNS runs a series of checks and fixes DNS unavailibilty from currNS
func RemedyNS(currNS *ns.NetNS, name string, ns string,
	client *kubernetes.Clientset) error {

	svcName = name
	ip := GetServiceIP(svcName, ns, client)

	//first, we check the resolv.conf file to check if resolvers are correct
	//
	cmd := exec.Command("cat", "/etc/resolv.conf")
	output, err := cmd.CombinedOutput()
	out := string(output)

	if err != nil {
		logrus.Error("Error opening resolv.conf: ", err)
		return errors.New("Cannot open resolv.conf: " + err.Error())
	}

	if !strings.Contains(out, ip) {
		logrus.Info("Incorrect resolver: ", out)
		logrus.Info("Replacing it with the svc ip ", ip)

		// replace existing ip with svc ip
		out = out[0:12] + ip + out[25:]
	}
	return nil
}
