package netperf

import (
	"fmt"
	"strings"
	"time"

	"gihub.com/jtaleric/k8s-netperf/metrics"
	stats "github.com/montanaflynn/stats"
)

// Data describes the result data
type Data struct {
	Config
	Metric            string
	SameNode          bool
	HostNetwork       bool
	ClientNodeInfo    metrics.NodeInfo
	ServerNodeInfo    metrics.NodeInfo
	Sample            Sample
	StartTime         time.Time
	EndTime           time.Time
	Service           bool
	ThroughputSummary []float64
	LatencySummary    []float64
	ClientMetrics     metrics.NodeCPU
	ServerMetrics     metrics.NodeCPU
	ClientPodCPU      metrics.PodValues
	ServerPodCPU      metrics.PodValues
}

// ScenarioResults each scenario could have multiple results
type ScenarioResults struct {
	Results []Data
}

// average accepts array of floats to calculate average
func average(vals []float64) (float64, error) {
	return stats.Median(vals)
}

// percentile accepts array of floats and the desired %tile to calculate
func percentile(vals []float64, ptile float64) (float64, error) {
	return stats.Percentile(vals, ptile)
}

// CheckResults will check to see if there are results with a specific Profile like TCP_STREAM
// returns true if there are results with provided string
func checkResults(s ScenarioResults, check string) bool {
	for t := range s.Results {
		if strings.Contains(s.Results[t].Profile, check) {
			return true
		}
	}
	return false
}

// CheckHostResults will check to see if there are hostNet results
// returns true if there are results with hostNetwork
func CheckHostResults(s ScenarioResults) bool {
	for t := range s.Results {
		if s.Results[t].HostNetwork {
			return true
		}
	}
	return false
}

// TCPThroughputDiff accepts the Scenario Results and calculates the %diff.
// returns
func TCPThroughputDiff(s ScenarioResults) (float64, error) {
	// We will focus on TCP STREAM
	hostPerf := 0.0
	podPerf := 0.0
	for t := range s.Results {
		if !s.Results[t].Service {
			if s.Results[t].HostNetwork {
				if s.Results[t].Profile == "TCP_STREAM" {
					hostPerf, _ = average(s.Results[t].ThroughputSummary)
				}
			} else {
				if s.Results[t].Profile == "TCP_STREAM" {
					podPerf, _ = average(s.Results[t].ThroughputSummary)
				}
			}
		}
	}
	return calDiff(hostPerf, podPerf), nil
}

// calDiff will determine the %diff between two values.
// returns a float64 which is the %diff
func calDiff(a float64, b float64) float64 {
	return (a - b) / ((a + b) / 2) * 100
}

// ShowPodCPU accepts ScenarioResults and presents to the user via stdout the PodCPU info
func ShowPodCPU(s ScenarioResults) {
	fmt.Printf("%s Pod CPU Utilization %s\r\n", strings.Repeat("-", 73), strings.Repeat("-", 73))
	fmt.Printf("%-18s | %-15s | %-15s | %-15s | %-15s | %-15s | %-25s | %-15s \r\n", "Role", "Scenario", "Host Network", "Service", "Message Size", "Same node", "Pod", "Utilization")
	fmt.Printf("%s\r\n", strings.Repeat("-", 166))
	for _, r := range s.Results {
		for _, pod := range r.ClientPodCPU.Results {
			fmt.Printf("📊 %-15s | %-15s | %-15t | %-15t | %-15d | %-15t | %-25s | %-15f \r\n", "Client", r.Profile, r.HostNetwork, r.Service, r.MessageSize, r.SameNode, fmt.Sprintf("%.20s", pod.Name), pod.Value)
		}
		fmt.Printf("%s\r\n", strings.Repeat("-", 166))
		for _, pod := range r.ServerPodCPU.Results {
			fmt.Printf("📊 %-15s | %-15s | %-15t | %-15t | %-15d | %-15t | %-25s | %-15f \r\n", "Server", r.Profile, r.HostNetwork, r.Service, r.MessageSize, r.SameNode, fmt.Sprintf("%.20s", pod.Name), pod.Value)
		}
		fmt.Printf("%s\r\n", strings.Repeat("-", 166))
	}
}

// ShowNodeCPU accepts ScenarioResults and presents to the user via stdout the NodeCPU info
func ShowNodeCPU(s ScenarioResults) {
	fmt.Printf("%s Node CPU Utilization %s\r\n", strings.Repeat("-", 111), strings.Repeat("-", 111))
	fmt.Printf("%-18s | %-15s | %-15s | %-15s | %-15s | %-15s | %-15s | %-15s | %-15s | %-15s | %-15s | %-15s | %-15s | %-15s\r\n", "Role", "Scenario", "Host Network", "Service", "Message Size", "Same node", "Idle CPU", "User CPU", "System CPU", "Steal CPU", "IOWait CPU", "Nice CPU", "SoftIRQ CPU", "IRQ CPU")
	fmt.Printf("%s\r\n", strings.Repeat("-", 245))
	for _, r := range s.Results {
		ccpu := r.ClientMetrics
		scpu := r.ServerMetrics
		fmt.Printf("📊 %-15s | %-15s | %-15t | %-15t | %-15d | %-15t | %-15f | %-15f | %-15f | %-15f | %-15f | %-15f | %-15f | %-15f\r\n", "Client", r.Profile, r.HostNetwork, r.Service, r.MessageSize, r.SameNode, ccpu.Idle, ccpu.User, ccpu.System, ccpu.Steal, ccpu.Iowait, ccpu.Nice, ccpu.Softirq, ccpu.Irq)
		fmt.Printf("📊 %-15s | %-15s | %-15t | %-15t | %-15d | %-15t | %-15f | %-15f | %-15f | %-15f | %-15f | %-15f | %-15f | %-15f\r\n", "Server", r.Profile, r.HostNetwork, r.Service, r.MessageSize, r.SameNode, scpu.Idle, scpu.User, scpu.System, scpu.Steal, ccpu.Iowait, ccpu.Nice, ccpu.Softirq, ccpu.Irq)
	}
	fmt.Printf("%s\r\n", strings.Repeat("-", 245))
}

// ShowStreamResult accepts NetPerfResults to display to the user via stdout
func ShowStreamResult(s ScenarioResults) {
	if checkResults(s, "STREAM") {
		fmt.Printf("%s Stream Results %s\r\n", strings.Repeat("-", 69), strings.Repeat("-", 69))
		fmt.Printf("%-18s | %-15s |%-15s | %-15s | %-15s | %-15s | %-15s | %-15s\r\n", "Scenario", "Host Network", "Service", "Message Size", "Same node", "Duration", "Samples", "Avg value")
		fmt.Printf("%s\r\n", strings.Repeat("-", 155))
		for _, r := range s.Results {
			if strings.Contains(r.Profile, "STREAM") {
				avg, _ := average(r.ThroughputSummary)
				fmt.Printf("📊 %-15s | %-15t |%-15t | %-15d | %-15t | %-15d | %-15d | %-15f (%s) \r\n", r.Profile, r.HostNetwork, r.Service, r.MessageSize, r.SameNode, r.Duration, r.Samples, avg, r.Metric)
			}
		}
		fmt.Printf("%s\r\n", strings.Repeat("-", 155))

	}
}

// ShowLatencyResult accepts NetPerfResults to display to the user via stdout
func ShowLatencyResult(s ScenarioResults) {
	if checkResults(s, "STREAM") {
		fmt.Printf("%s Stream Latency Results %s\r\n", strings.Repeat("-", 65), strings.Repeat("-", 65))
		fmt.Printf("%-18s | %-15s |%-15s | %-15s | %-15s | %-15s | %-15s | %-15s\r\n", "Scenario", "Host Network", "Service", "Message Size", "Same node", "Duration", "Samples", "99%tile value")
		fmt.Printf("%s\r\n", strings.Repeat("-", 155))
		for _, r := range s.Results {
			if strings.Contains(r.Profile, "STREAM") {
				avg, _ := average(r.LatencySummary)
				fmt.Printf("📊 %-15s | %-15t |%-15t | %-15d | %-15t | %-15d | %-15d | %-15f (%s) \r\n", r.Profile, r.HostNetwork, r.Service, r.MessageSize, r.SameNode, r.Duration, r.Samples, avg, "usec")
			}
		}
		fmt.Printf("%s\r\n", strings.Repeat("-", 155))
	}
	if checkResults(s, "RR") {
		fmt.Printf("%s RR Latency Results %s\r\n", strings.Repeat("-", 66), strings.Repeat("-", 66))
		fmt.Printf("%-18s | %-15s |%-15s | %-15s | %-15s | %-15s | %-15s | %-15s\r\n", "Scenario", "Host Network", "Service", "Message Size", "Same node", "Duration", "Samples", "99%tile value")
		fmt.Printf("%s\r\n", strings.Repeat("-", 155))
		for _, r := range s.Results {
			if strings.Contains(r.Profile, "RR") {
				avg, _ := average(r.LatencySummary)
				fmt.Printf("📊 %-15s | %-15t |%-15t | %-15d | %-15t | %-15d | %-15d | %-15f (%s) \r\n", r.Profile, r.HostNetwork, r.Service, r.MessageSize, r.SameNode, r.Duration, r.Samples, avg, "usec")
			}
		}
		fmt.Printf("%s\r\n", strings.Repeat("-", 155))
	}
}

// ShowRRResult will display the RR performance results
// Currently showing the Avg Value.
// TODO: Capture latency values
func ShowRRResult(s ScenarioResults) {
	if checkResults(s, "RR") {
		fmt.Printf("%s RR Results %s\r\n", strings.Repeat("-", 72), strings.Repeat("-", 72))
		fmt.Printf("%-18s | %-15s |%-15s | %-15s | %-15s | %-15s | %-15s | %-15s\r\n", "Scenario", "Host Network", "Service", "Message Size", "Same node", "Duration", "Samples", "Avg value")
		fmt.Printf("%s\r\n", strings.Repeat("-", 155))
		for _, r := range s.Results {
			if strings.Contains(r.Profile, "RR") {
				avg, _ := average(r.ThroughputSummary)
				fmt.Printf("📊 %-15s | %-15t |%-15t | %-15d | %-15t | %-15d | %-15d | %-15f (%s) \r\n", r.Profile, r.HostNetwork, r.Service, r.MessageSize, r.SameNode, r.Duration, r.Samples, avg, r.Metric)
			}
		}
		fmt.Printf("%s\r\n", strings.Repeat("-", 155))
	}
}
