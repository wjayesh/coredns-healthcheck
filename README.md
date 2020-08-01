# HealthCheck-CoreDNS
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/WJayesh/health-check) &nbsp;
[![Go Report Card](https://goreportcard.com/badge/github.com/wjayesh/coredns-healthcheck)](https://goreportcard.com/report/github.com/wjayesh/coredns-healthcheck) &nbsp;
![Docker Image CI](https://github.com/WJayesh/healthCheck/workflows/Docker%20Image%20CI/badge.svg) &nbsp; ![lint-test](https://github.com/WJayesh/health-check/workflows/lint-test/badge.svg) 

A binary and packages to perform health checks on pods and services running on Kubernetes and to remedy any failures.

## Contents

* [**Objective**](https://github.com/wjayesh/coredns-healthcheck/tree/main/#objective)

* [**Motivation and Scope**](https://github.com/wjayesh/coredns-healthcheck/tree/main/#motivation-and-scope)

* [**Architecture**](https://github.com/wjayesh/coredns-healthcheck/tree/main/#architecture)

* [**Workflow**](https://github.com/wjayesh/coredns-healthcheck/tree/main/#workflow)

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
- Restart CoreDNS pods by interacting with Kubernetes API through the Golang application, if the response from the cluster and pod IPs is unsatisfactory.  

Thus, making the state of CoreDNS available externally and reliably is important to ensure important services run as they are expected to.

## Architecture 

The binary is designed to access all network namespaces and run commands in them. As such, it needs to be deployed on every node, just like a CNI plugin. 
This can be achieved through a `DaemonSet`.

![Architecture](https://github.com/wjayesh/coredns-healthcheck/blob/docs/assets/docs/images/Architecture%201.png)

Inside a node, there exist different pods bound to their respective namespaces. The binary is deployed on the host network and is thus on the root network namespace. 

![Inside Node](https://github.com/wjayesh/coredns-healthcheck/blob/docs/assets/docs/images/Inside%20Node.png)

### Workflow

Firstly, the binary queries the CoreDNS pods from the host namespace and checks the repsonse. 

* If the response received is unsatisfactory, then the pods are restarted, or the memory limit is increased if restarting doesn't help.

* If the response is correct, the binary proceeds to list all namespaces on the host and starts a session in each, one by one.

  ![Arch. Wf 1](https://github.com/wjayesh/coredns-healthcheck/blob/main/assets/docs/images/Arch.%20Wf%201.png)

  It then queries the CoreDNS pods from every namespace to check the DNS availability.

  ![Arch. Wf 2](https://github.com/wjayesh/coredns-healthcheck/blob/main/assets/docs/images/Arch.%20Wf%202.png)

  If the service is unavailable from any namespace, the `etc/resolv.conf` file is then inspected to look for possible causes of failure. 
  

## Deployment 


The application can be deployed either inside a Kubernetes cluster or outside it. Even inside a cluster, it can be deployed as a `DaemonSet` so that ir runs on all nodes or it can be deployed on a single node too. 
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
  - name: health-check-container
    image: wjayesh/health:latest
    args: ["-path=PATH", "-allowPods=BOOL", "-udpPort=PORT"]
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
      - name: health-check-container
        image: wjayesh/health:latest
        args: ["-path=PATH", "-allowPods=BOOL", "-udpPort=PORT"]
        restartPolicy: OnFailure
  ```
  
  #### Note 
  * The DNS ploicy needs to be set to `ClusterFirstWithHostNet` explicitly. Reference: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service/#pod-s-dns-policy
  
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
  

