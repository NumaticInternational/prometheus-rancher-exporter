package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

func joinLabels(labels map[string]string) string {
	var result string
	var labelsArray []string
	for name, val := range labels {
		labelsArray = append(labelsArray, fmt.Sprintf("%s=%s", name, val))
	}
	if len(labelsArray) > 0 {
		// Sort for same order
		sort.Strings(labelsArray)
		result = fmt.Sprintf(",%s,", strings.Join(labelsArray, ","))
	}
	return result
}

// addMetrics - Add's all of the GuageVecs to the `guageVecs` map, returns the map.
func addMetrics() map[string]*prometheus.GaugeVec {
	gaugeVecs := make(map[string]*prometheus.GaugeVec)

	// Stack Metrics
	gaugeVecs["stacksHealth"] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "stack_health_status",
			Help:      "HealthState of defined stack as reported by Rancher",
		}, []string{"name", "health_state", "system"})
	gaugeVecs["stacksState"] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "stack_state",
			Help:      "State of defined stack as reported by Rancher",
		}, []string{"name", "state", "system"})

	// Service Metrics
	gaugeVecs["servicesScale"] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_scale",
			Help:      "scale of defined service as reported by Rancher",
		}, []string{"name", "stack_name", "labels"})
	gaugeVecs["servicesHealth"] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_health_status",
			Help:      "HealthState of the service, as reported by the Rancher API. Either (1) or (0)",
		}, []string{"name", "stack_name", "health_state", "labels"})
	gaugeVecs["servicesState"] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "service_state",
			Help:      "State of the service, as reported by the Rancher API",
		}, []string{"name", "stack_name", "state", "labels"})

	// Host Metrics
	gaugeVecs["hostsState"] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      ("host_state"),
			Help:      "State of defined host as reported by the Rancher API",
		}, []string{"name", "state", "labels"})
	gaugeVecs["hostAgentsState"] = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      ("host_agent_state"),
			Help:      "State of defined host agent as reported by the Rancher API",
		}, []string{"name", "state", "labels"})

	return gaugeVecs
}

// checkMetric - Checks the base type stored in the API is correct, this ensures we are setting the right metric for the right endpoint.
func checkMetric(endpoint string, baseType string) bool {
	e := strings.TrimSuffix(endpoint, "s")

	// Backwards compatibility fix, the API in V1 wrong, this is to cover v1 usage.
	if baseType == "environment" && e == "stack" {
		return true
	} else if e == "service" && (baseType == "externalService" || baseType == "loadBalancerService") {
		return true
	} else if e != baseType {
		log.Errorf("API MisMatch, expected %s metric, got %s metric", e, baseType)
		return false
	}

	return true
}

// setServiceMetrics - Logic to set the state of a system as a gauge metric
func (e *Exporter) setServiceMetrics(name string, stack string, state string, health string, scale int, labels map[string]string) error {
	labelsStr := joinLabels(labels)
	e.gaugeVecs["servicesScale"].With(prometheus.Labels{
		"name":       name,
		"stack_name": stack,
		"labels":     labelsStr,
	}).Set(float64(scale))
	for _, y := range healthStates {
		gauge := e.gaugeVecs["servicesHealth"].With(prometheus.Labels{
			"name":         name,
			"stack_name":   stack,
			"health_state": y,
			"labels":       labelsStr,
		})
		if health == y {
			gauge.Set(1)
		} else {
			gauge.Set(0)
		}
	}
	for _, y := range serviceStates {
		gauge := e.gaugeVecs["servicesState"].With(prometheus.Labels{
			"name":       name,
			"stack_name": stack,
			"state":      y,
			"labels":     labelsStr,
		})
		if state == y {
			gauge.Set(1)
		} else {
			gauge.Set(0)
		}
	}
	return nil
}

// setStackMetrics - Logic to set the state of a system as a gauge metric
func (e *Exporter) setStackMetrics(name string, state string, health string, system string) error {
	for _, y := range healthStates {
		gauge := e.gaugeVecs["stacksHealth"].With(prometheus.Labels{
			"name":         name,
			"health_state": y,
			"system":       system,
		})
		if health == y {
			gauge.Set(1)
		} else {
			gauge.Set(0)
		}
	}
	for _, y := range stackStates {
		gauge := e.gaugeVecs["stacksState"].With(prometheus.Labels{
			"name":   name,
			"state":  y,
			"system": system,
		})
		if state == y {
			gauge.Set(1)
		} else {
			gauge.Set(0)
		}
	}
	return nil
}

// setHostMetrics - Logic to set the state of a system as a gauge metric
func (e *Exporter) setHostMetrics(name string, state, agentState string, labels map[string]string) error {
	labelsStr := joinLabels(labels)
	for _, y := range hostStates {
		gauge := e.gaugeVecs["hostsState"].With(prometheus.Labels{
			"name":   name,
			"state":  y,
			"labels": labelsStr,
		})
		if state == y {
			gauge.Set(1)
		} else {
			gauge.Set(0)
		}
	}
	for _, y := range agentStates {
		gauge := e.gaugeVecs["hostAgentsState"].With(prometheus.Labels{
			"name":   name,
			"state":  y,
			"labels": labelsStr,
		})
		if agentState == y {
			gauge.Set(1)
		} else {
			gauge.Set(0)
		}
	}
	return nil
}
