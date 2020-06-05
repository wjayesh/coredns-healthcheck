# HealthCheck-CoreDNS
![Docker Image CI](https://github.com/WJayesh/healthCheck/workflows/Docker%20Image%20CI/badge.svg)

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
Information pertaining to deploying this binary on Docker and Kubernetes is provided in the [Deployment.md](https://github.com/WJayesh/healthCheck/blob/master/DEPLOYMENT.md) file. 
