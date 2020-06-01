# Deployment 

The application can be deployed either inside a Kubernetes cluster or outside it. When the deployment is done as a pod in a cluster, 
no flags need to be used. 

When deploying externally, the `kubeconfig` file path has to be provided so that authentication with the api-server can be done. 

Additionally, two flags can be set as the need be:

1) `allowPods` : `boolean` value that states whether pod creation inside the cluster is allowed.
2) `udpPort` : If CoreDNS pods are using some port other than port 53, specify that here. 

## Docker

The `Dockerfile` is present at the root directory and an image is also pushed to the DockerHub at `wjayesh/health`. 

To deploy the application on Docker, use the following command:
```python
docker run wjayesh/health:v1 -path=PATH -allowPods=BOOL -udpPort=PORT
```

## Kubernetes

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
  
  ### Note 
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
    verbs: ["get", "watch", "list"]
  ```
  This cluster role can be bound to your default service account in the default namespace as follows:
  ```
  kubectl create clusterrolebinding ip-finder-pod \
  --clusterrole=ip-finder  \
  --serviceaccount=default:default
  ```
  
