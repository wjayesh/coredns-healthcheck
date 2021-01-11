# HealthCheck-CoreDNS
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/WJayesh/health-check) &nbsp;
[![Go Report Card](https://goreportcard.com/badge/github.com/wjayesh/coredns-healthcheck)](https://goreportcard.com/report/github.com/wjayesh/coredns-healthcheck) &nbsp;
![Docker Image CI](https://github.com/WJayesh/healthCheck/workflows/Docker%20Image%20CI/badge.svg) &nbsp; ![lint-test](https://github.com/WJayesh/health-check/workflows/lint-test/badge.svg) 

A binary and packages to perform health checks on pods and services running on Kubernetes and to remedy any failures.

**push event**

## Contents

* [**Objective**](https://github.com/wjayesh/coredns-healthcheck/tree/main/#objective)

* [**Motivation and Scope**](https://github.com/wjayesh/coredns-healthcheck/tree/main/#motivation-and-scope)

* [**Architecture**](https://github.com/wjayesh/coredns-healthcheck/tree/main/#architecture)

* [**Workflow**](https://github.com/wjayesh/coredns-healthcheck/tree/main/#workflow)

* [**Prometheus Monitoring**](https://github.com/wjayesh/coredns-healthcheck#prometheus-exporter)

* [**Deployment**](https://github.com/wjayesh/coredns-healthcheck/tree/main/#deployment)

* [**Milestones**](https://github.com/WJayesh/coredns-healthcheck/tree/main#milestones-)

## Objective

CoreDNS is the cluster DNS server for Kubernetes and is very critical for the overall health of the Kubernetes cluster. It is important to monitor the health of CoreDNS itself and restarting or repairing any CoreDNS pods that are not behaving correctly.  
 
While CoreDNS exposes a health check itself in the form of Kubernetes’ `livenessProbe`: 


- The health check is not UDP (DNS) based. There have been cases where the health port is accessible (for TCP) but CoreDNS itself isn't (UDP). This protocol difference means that CoreDNS is unhealthy from a cluster standpoint, but the control plane can't see this.  
- The existing health check is also launched locally (the `kubelet` uses the `livenessProbe`) and the situation could be different for pods accessing it remotely. 
 
## Motivation and Scope


This project idea aims to get around limitations on Kubernetes’ health check and build an application that: 


- Checks CoreDNS health externally through UDP (DNS), from a remote Golang application. 
- Remedy CoreDNS pods which involves restarting, increasing memory limits, changing network configuration and more, if the response from the cluster and pod IPs is unsatisfactory. This is done by interacting with Kubernetes API through the Golang application.  

Thus, making the state of CoreDNS available externally and reliably is important to ensure important services run as they are expected to.

## Architecture 

The binary is designed to access all network namespaces and run commands in them. As such, it needs to be deployed on every node, just like a CNI plugin. 
This can be achieved through a `DaemonSet`.

![Architecture](https://github.com/wjayesh/coredns-healthcheck/blob/docs/assets/docs/images/Architecture%201.png)

Inside a node, there exist different pods bound to their respective namespaces. The binary is deployed on the host network and is thus on the root network namespace. 

![Inside Node](https://github.com/wjayesh/coredns-healthcheck/blob/docs/assets/docs/images/Inside%20Node.png)

### Workflow

Firstly, the binary queries the CoreDNS pods from the host network namespace and checks the response. 

* If the response received is unsatisfactory, then the pods are restarted, or the memory limit is increased if restarting doesn't help.

* If the response is correct, the binary proceeds to list all network namespaces on the host and starts a session in each, one by one.

  ![Arch. Wf 1](https://github.com/wjayesh/coredns-healthcheck/blob/main/assets/docs/images/Arch.%20Wf%201.png)
  
  #### Entering a namespace
  
  To  accomplish this, we need the PIDs of the container so that the network namespaces can be located on the host's `/proc` directory.

  One way to get the PIDs is to use the following command.

  ```
  pid = "$(docker inspect -f '{{.State.Pid}}' "container_id")"
  ```

  This however, requires that the container be able to run docker commands on the host. The way this application approaches this problem is:

  1) Mounting the docker daemon from the host to the container (`/var/run/docker.sock`)

  2) Installing the Docker CLI as part of the Dockerfile. This CLI is necessary to communicate with the daemon. 

  3) Mounting the host's `/proc` directory on the container. This way, we have to the network namespaces corresponding to the different pids. 


  We can use the path to the network namespace to obtain an [`NetNS`](https://pkg.go.dev/github.com/containernetworking/plugins/pkg/ns?tab=doc#NetNS) object from the package ns. This object has a function `Do()` that can be used to execute functions inside other namespaces. 
  
  Thus, the binary now queries the CoreDNS pods from every namespace to check the DNS availability.

  ![Arch. Wf 2](https://github.com/wjayesh/coredns-healthcheck/blob/main/assets/docs/images/Arch.%20Wf%202.png)
  
  The following is a screenshot from a trial run to test this functionality. We can see that the two pod IPs and one service IPs are queried.
  The logs serve to show the namespace in which the operations are performed.
  
  ![NetNS check](https://user-images.githubusercontent.com/37150991/89679103-0aab7600-d90e-11ea-811f-edd2af28f130.png)
  

  If the service is unavailable from any namespace, then a series of checks are performed:
  
  *  `etc/resolv.conf` file is inspected to check if the nameserver is the correct one (corresponding to the DNS service). 
  
  &nbsp;
  
  Other measures are proposed but not yet implemented. These include:
  
  * Checking if the route table has an entry for a `DefaultGateway` and if its the one corresponding to the bridge.
  * Do virtual interfaces (veth) exist and are they connected to the bridge?
  * Are `iptables` rules restricting connections?
  
  
## Prometheus Exporter

A exporter library is implemented at [`pkg/exporter`](https://github.com/wjayesh/coredns-healthcheck/tree/main/pkg/exporter) that takes values from the application and registers them with Prometheus using the golang client. 

The exporter will help determine the number of times the remedies were required, how often the pods failed, the primary reasons for the failures (ascertained by the type of remedy that fixed it) among other things.

### Available Groups Of Data

Name | Description
-----| ------
remedy | This group has metrics related to the remedial measures taken when the pods fail, such as restarting pods or increasing memory allocation.
dns | This group has metrics that deal with dns queries made by the application and their response. 

&nbsp;

Remedy group has the following available metrics:

Name | Metric Type | Exposed Information
----  | ---- | ---
`oom_count` | Counter | Counts the number of OOM errors encountered
`restart_count` | Counter | Counts the number of restarts performed on the pods
`total_failures` | Counter | Counts the total number of failures of the pods under check


DNS group has the following available metrics:

Name | Metric Type | Exposed Information
----  | ---- | ---
`dns_query_count` | Counter | Counts the number of DNS queries made
`dns_query_response_time` | Histogram | The time it takes to finish dns queries

### Port

The metrics are exposed on the port `9890` of the pod's IP address. The endpoint is `/metrics`. 

A sample call could be done in the following manner:
```bash
curl 10.244.0.13:9890/metrics 
```

### Visuals

A sample run with an orchestrated OOM error:

![OOM](https://user-images.githubusercontent.com/37150991/89416485-f10e0100-d74a-11ea-8eeb-07ae7bda3a22.png)

This shows the number of restarts that fixed the pods and the total number of failures. 

![Restarts and Failures](https://user-images.githubusercontent.com/37150991/89416967-afca2100-d74b-11ea-8873-99905b42b57a.png)


More visuals will be added as tests proceed.


## Deployment 


The application can be deployed either inside a Kubernetes cluster or outside it. Even inside a cluster, it can be deployed as a `DaemonSet` so that it runs on all nodes or it can be deployed on a single node too. 
The driving principle behind this binary is that it should function gracefully under all conditions. 

When the deployment is done as a pod in a cluster, 
no flags need to be used. 

When deploying externally, the `kubeconfig` file path has to be provided so that authentication with the api-server can be done. 

Additionally, two flags can be set as the need be:

1) `allowPods` : `boolean` value that states whether pod creation inside the cluster is allowed.
2) `udpPort` : If CoreDNS pods are using some port other than port 53, specify that here. 

### Docker

The `Dockerfile` is present at the root directory and an image is also pushed to the DockerHub at `wjayesh/health`. 

To deploy the application on Docker, use the following command:
```python
docker run wjayesh/health:latest -path=PATH -allowPods=BOOL -udpPort=PORT
```

### Kubernetes

The image from `wjayesh/health` can be used to create Pods. A basic `YAML` description is provided below. 
```yml
apiVersion: v1
kind: Pod
metadata:
  name: health-check
  labels:
    target: coredns-deployment
spec:
  hostNetwork: true    # to have access to all netns
  dnsPolicy: ClusterFirstWithHostNet  # needs to be set explicitly 
   containers:
   - image: wjayesh/health:latest
     name: health-check-container
     args: ["-path=PATH", "-allowPods=BOOL", "-port=PORT"]  # fill these values or skip line if deployed 
     volumeMounts:                                          # on the same cluster as the pods under check
     - mountPath: /hostProc  
       name: proc
     - mountPath: /var/run/docker.sock
       name: docker-sock
       readOnly: false
     - mountPath: /var/lib/docker
       name: docker-directory
       readOnly: false
     securityContext:
       privileged: true   # to run docker command for finding PIDs of containers whose ns to enter
   volumes:
   - name: proc   # the net ns will be located at /proc/pid/ns/net
     hostPath:
       path: /proc     
   - name: docker-sock   # using the docker daemon of the host 
     hostPath:
       path: "/var/run/docker.sock"
       type: File
   - name: docker-directory  # contains all images and other info
     hostPath:
       path: "/var/lib/docker"
  restartPolicy: OnFailure
  ```
  
  This definition can be used in a `DaemonSet`. An example:
  
```yml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: coredns-hc
  namespace: kube-system
  labels:
    k8s-app: health-check
spec:
  selector:
    matchLabels:
      target: coredns-deployment
  template:
    metadata:
      labels:
        target: coredns-deployment
    spec:
      tolerations:
      # this toleration is to have the daemonset runnable on master nodes
      # remove it if your masters can't run pods
      - key: node-role.kubernetes.io/master
        effect: NoSchedule
      hostNetwork: true
      dnsPolicy: ClusterFirstWithHostNet  # needs to be set explicitly 
      containers:
      - image: wjayesh/health:latest
        name: health-check-container
        args: ["-path=PATH", "-allowPods=BOOL", "-udpPort=PORT"]
        volumeMounts:
        - mountPath: /hostProc  
          name: proc
        - mountPath: /var/run/docker.sock
          name: docker-sock
          readOnly: false
        - mountPath: /var/lib/docker
          name: docker-directory
          readOnly: false
        securityContext:
          privileged: true   # to run docker command for finding PIDs of containers whose ns to enter
      volumes:
      - name: proc   # the net ns will be located at /proc/pid/ns/net
        hostPath:
          path: /proc     
      - name: docker-sock   # using the docker daemon of the host 
        hostPath:
          path: "/var/run/docker.sock"
          type: File
      - name: docker-directory  # contains all images and other info
        hostPath:
          path: "/var/lib/docker"
  ```
  
  #### Note 
  * The DNS ploicy needs to be set to `ClusterFirstWithHostNet` explicitly. Reference: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy
  
  * To allow the application to access the namespaces (specifically to first get the PIDs used to locate the network ns), it has to be run as a privileged pod.
  If creation of the pod fails in your cluster, check whether you have a   [`PodSecurityPolicy`](https://kubernetes.io/docs/concepts/policy/pod-security-policy/#privileged) that allows privileged pods. 
  
  * Keep in mind that you cannot use environment variables like `"$(PORT)"` as identifiers inside the args field. 
  This is because there is no shell being run in the container and your variables won't resolve to their values. 
  
  * Make sure your service account has a role that can access the services, pods and deployment resources of your cluster. An example ClusterRole with some privileges is shown below. 
  ```yml
  kind: ClusterRole
  apiVersion: rbac.authorization.k8s.io/v1
  metadata:
    namespace: kube-system
    name: health-manager
  rules:
  - apiGroups: [""] # "" indicates the core API group
    resources: ["services", "pods"]
    verbs: ["get", "watch", "list", "create", "update", "patch",  "delete"]
  - apiGroups: ["extensions", "apps"]
    resources: ["deployments"]
    verbs: ["get", "watch", "list", "create", "update", "patch",  "delete"]
  ```
  This cluster role can be bound to your default service account in the default namespace as follows:
  ```
  kubectl create clusterrolebinding health-role-pod \
  --clusterrole=health-manager  \
  --serviceaccount=default:default
  ```
  ## Milestones ✨
  
Here I will list the milestones achieved in sync with the tasks done on the project board.  

* Connection to the api-server established on an AKS cluster. 

  ![Cloud Shell logs from the health-check pod](https://user-images.githubusercontent.com/37150991/83383036-ee68f580-a401-11ea-9340-970411d09652.png)

* Service and Pod IPs successfully retrieved.

  ![Logs show the different IPs](https://user-images.githubusercontent.com/37150991/83383009-e6a95100-a401-11ea-810d-a3d3b4c83b98.png)

* Restarting CoreDNS pods through the binary. The logs shows the pods to be deleted. 

  ![The logs show the pods to be deleted](https://user-images.githubusercontent.com/37150991/84131657-7a15fe00-aa62-11ea-8ebd-5410f28ce786.png)

  The condition of invalid output has been harcoded in order to force a restart, for testing purposes. 

  ![We can see that new pods have been created](https://user-images.githubusercontent.com/37150991/84131724-8dc16480-aa62-11ea-85d9-d9fe5117cbad.png)
  We can see that new pods have been created. 

* Functionality of `dig` replicated using `miekg/exdns/q` inside the health-check app.
  The first two IPs belong to the CoreDNS pods. The third is the `ClusterIP` for the `kube-dns` service.

  ![](https://user-images.githubusercontent.com/37150991/84260469-b9624e80-ab37-11ea-8a53-a4e3f8d95875.png)
  I have selected the `kubernetes.default` service to test the DNS response. 
  
* Doubling the memory allocation for the CoreDNS pods.

  ![](https://user-images.githubusercontent.com/37150991/89102633-dd039000-d428-11ea-9292-41d54b61a3bf.png)
  At first, the memory limit is 170Mi
  

  ![](https://user-images.githubusercontent.com/37150991/89102657-0e7c5b80-d429-11ea-8760-c5ec87e7ea97.png)
  Later, it is doubled to 340Mi. This was triggered when multiple subsequent restarts failed to make the pods healthy.
  

