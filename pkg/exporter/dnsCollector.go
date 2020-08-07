package exporter

import (
	"github.com/WJayesh/coredns-healthcheck/pkg/health"
	"github.com/prometheus/client_golang/prometheus"
)

//DNSCollector is a struct containing pointers to
//prometheus descriptors for each metric
type DNSCollector struct {
	dnsQueryCount *prometheus.Desc
	respTime      *prometheus.Histogram
}

//NewDNSCollector is a constructor that initializes every descriptor and
//returns a pointer to the collector
func NewDNSCollector() *DNSCollector {

	//getting histogram reference
	_, hist := health.GetDNSMetrics()

	return &DNSCollector{
		dnsQueryCount: prometheus.NewDesc(
			"dns_query_count",
			"Counts the number of DNS queries made",
			nil, nil,
		),
		respTime: hist,
	}
}

//Describe writes all descriptors to the prometheus desc channel.
func (collector *DNSCollector) Describe(ch chan<- *prometheus.Desc) {

	//Using a helper to return the Desc from the struct.
	//checking if histogram's describe is implicitly called
	prometheus.DescribeByCollect(collector, ch)
}

//Collect implements required collect function for all promehteus collectors
func (collector *DNSCollector) Collect(ch chan<- prometheus.Metric) {

	//calling func to get metrics from app
	queryCount, _ := health.GetDNSMetrics()

	//Write values to the channel
	ch <- prometheus.MustNewConstMetric(
		collector.dnsQueryCount,
		prometheus.CounterValue,
		queryCount,
	)

	// //adding metric to channel ch
	// hist := *collector.respTime
	// //checking for nil pointer
	// if hist == nil {
	// 	return
	// }
	// hist.Collect(ch)
}
