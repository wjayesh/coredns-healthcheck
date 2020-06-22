module github.com/WJayesh/coredns-healthcheck

go 1.14

require (
	github.com/miekg/dns v1.1.29
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	k8s.io/api v0.18.4
	k8s.io/apimachinery v0.18.4
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/metrics v0.18.4
)
