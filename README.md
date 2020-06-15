# HealthCheck-CoreDNS
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://pkg.go.dev/github.com/WJayesh/health-check) &nbsp;
![Docker Image CI](https://github.com/WJayesh/healthCheck/workflows/Docker%20Image%20CI/badge.svg) &nbsp; ![lint-test](https://github.com/WJayesh/health-check/workflows/lint-test/badge.svg) 

Repository to host work done as part of the Community Bridge program under CoreDNS. 

The [Wiki](https://github.com/WJayesh/healthCheck/wiki) section holds a list of milestones achieved to help track the current development status. 

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

## Deployment 

The application can be deployed either inside a Kubernetes cluster or outside it. When the deployment is done as a pod in a cluster, 
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
  containers:
  - name: health-check-container
    image: wjayesh/health:latest
    args: ["-path=PATH", "-allowPods=BOOL", "-udpPort=PORT"]
  restartPolicy: OnFailure
  ```
  
  #### Note 
  * Keep in mind that you cannot use environment variables like `"$(PORT)"` as identifiers inside the args field. 
  This is because there is no shell being run in the container and your variables won't resolve to their values. 
  
  * Make sure your service account has a role that can access the services and pods resources of your cluster. An example ClusterRole with basic privileges is shown below. 
  ```yml
  kind: ClusterRole
  apiVersion: rbac.authorization.k8s.io/v1
  metadata:
    namespace: default
    name: ip-finder
  rules:
  - apiGroups: [""] # "" indicates the core API group
    resources: ["services", "pods"]
    verbs: ["get", "watch", "list", "create", "update", "delete"]
  ```
  This cluster role can be bound to your default service account in the default namespace as follows:
  ```
  kubectl create clusterrolebinding ip-finder-pod \
  --clusterrole=ip-finder  \
  --serviceaccount=default:default
  ```
  

