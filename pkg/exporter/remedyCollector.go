package exporter

import (
	"github.com/WJayesh/coredns-healthcheck/pkg/health"
	"github.com/prometheus/client_golang/prometheus"
)

//RemedyCollector is a struct containing pointers to
//prometheus descriptors for each metric
type RemedyCollector struct {
	oomCount      *prometheus.Desc
	restartCount  *prometheus.Desc
	totalFailures *prometheus.Desc
}

//NewRemedyCollector is a constructor that initializes every descriptor and
//returns a pointer to the collector
func NewRemedyCollector() *RemedyCollector {
	return &RemedyCollector{
		oomCount: prometheus.NewDesc("oom_count",
			"Counts the number of OOM errors encountered",
			nil, nil,
		),
		restartCount: prometheus.NewDesc("restart_count",
			"Counts the number of restarts performed on the pods",
			nil, nil,
		),
		totalFailures: prometheus.NewDesc("total_failures",
			"Counts the total number of failures of the pods under check",
			nil, nil,
		),
	}
}

//Describe writes all descriptors to the prometheus desc channel.
func (collector *RemedyCollector) Describe(ch chan<- *prometheus.Desc) {

	//Using a helper to return the Desc from the struct.
	prometheus.DescribeByCollect(collector, ch)
}

//Collect implements required collect function for all promehteus collectors
func (collector *RemedyCollector) Collect(ch chan<- prometheus.Metric) {

	//logic to determine proper metric value to return to prometheus
	//for each descriptor.
	oom, restart, total := health.GetRemedyMetrics()

	//Write latest value for each metric in the prometheus metric channel.
	ch <- prometheus.MustNewConstMetric(
		collector.oomCount,
		prometheus.CounterValue,
		oom)

	ch <- prometheus.MustNewConstMetric(
		collector.restartCount,
		prometheus.CounterValue,
		restart)

	ch <- prometheus.MustNewConstMetric(
		collector.totalFailures,
		prometheus.CounterValue,
		total)

}
