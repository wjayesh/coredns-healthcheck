package exporter

import (
	"github.com/WJayesh/coredns-healthcheck/pkg/health"
	"github.com/prometheus/client_golang/prometheus"
)

//DNSCollector is a struct containing pointers to
//prometheus descriptors for each metric
type DNSCollector struct {
	dnsQueryCount *prometheus.Desc
}

//NewDNSCollector is a constructor that initializes every descriptor and
//returns a pointer to the collector
func NewDNSCollector() *DNSCollector {
	return &DNSCollector{
		dnsQueryCount: prometheus.NewDesc(
			"dns_query_count",
			"Counts the number of DNS queries made",
			nil, nil,
		),
	}
}

//Describe writes all descriptors to the prometheus desc channel.
func (collector *DNSCollector) Describe(ch chan<- *prometheus.Desc) {

	//Using a helper to return the Desc from the struct.
	prometheus.DescribeByCollect(collector, ch)
}

//Collect implements required collect function for all promehteus collectors
func (collector *DNSCollector) Collect(ch chan<- prometheus.Metric) {

	//calling func to get metrics from app
	queryCount := health.GetDNSMetrics()

	//Write values to the channel
	ch <- prometheus.MustNewConstMetric(
		collector.dnsQueryCount,
		prometheus.CounterValue,
		queryCount,
	)
}
