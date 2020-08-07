package netns

import (
	"os/exec"

	"github.com/containernetworking/plugins/pkg/ns"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// pods is a list of all pods on the cluster.
func getPIDs(pods *[]v1.Pod) []string {
	list := make([]string, 1)
	// find the conatiner id for podname
	for _, pod := range *pods {
		containerID := pod.Status.ContainerStatuses[0].ContainerID
		// taking only 12 digits and excluding "docker://"
		containerID = containerID[9:21]

		//[debug]
		logrus.Info("container ID: ", containerID)

		// command to obtain the process id
		cmd := exec.Command("docker", "inspect", "-f", "'{{.State.Pid}}'", containerID)
		out, err := cmd.CombinedOutput()

		//[debug]
		logrus.Info("Output after docker inspect: ", string(out))

		if err != nil {
			logrus.Error("Error getting the pid ", err)
			out = make([]byte, 0)
		}

		pid := string(out)
		list = append(list, pid)
	}
	logrus.Info("The list of pids: ", list)
	return list
}

// GetNetNS returns a list of ns.NetNS objects
func GetNetNS(client *kubernetes.Clientset) *[]ns.NetNS {
	list := make([]ns.NetNS, 0)

	pods := ListPods(client)
	pids := getPIDs(pods)

	for _, pid := range pids {
		// if the container is not on this node, the pid would be empty
		if pid == "" {
			continue
		}

		// removing quotes from pid
		pid = pid[1 : len(pid)-2]

		// the location for the netns on the node
		// hostProc is the mount of the proc dir of the host
		path := "/hostProc/" + pid + "/ns/net"

		// obtaining the NetNS object
		netns, err := ns.GetNS(path)
		if err != nil {
			logrus.Error("Error getting NS object: ", err)
		} else {
			list = append(list, netns)
		}
	}
	return &list
}
