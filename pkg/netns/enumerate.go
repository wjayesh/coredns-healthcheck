package netns

import (
	"os/exec"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

// pods is a list of all pods on the cluster.
func getPIDs(pods []*v1.Pod) []string {
	list := make([]string, 1)
	// find the conatiner id for podname
	for _, pod := range pods {
		containerID := pod.Status.ContainerStatuses[0].ContainerID
		cmd := exec.Command("docker inspect", "-f", "'{{.State.Pid}}'", containerID)
		out, err := cmd.CombinedOutput()
		if err != nil {
			logrus.Error("Error getting the pid", err)
			out = make([]byte, 1)
		}
		pid := string(out)
		list = append(list, pid)
	}
	return list
}

func GetNetNS() {
	//pods := ListPods()
	//getPIDs()

}
