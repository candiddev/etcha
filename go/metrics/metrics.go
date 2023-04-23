// Package metrics contains metrics for Etcha.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

//nolint:gochecknoglobals
var (
	CommandsChecked *prometheus.CounterVec
	CommandsChanged *prometheus.CounterVec
	CommandsRemoved *prometheus.CounterVec
)

func init() { //nolint:gochecknoinits
	CommandsChanged = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Help: "Counter of Commands that have ran ExecChange",
			Name: "etcha_commands_changed_total",
		},
		[]string{"id", "status"},
	)
	prometheus.MustRegister(CommandsChanged)

	CommandsChecked = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Help: "Counter of Commands that have ran ExecCheck",
			Name: "etcha_commands_checked_total",
		},
		[]string{"id", "status"},
	)
	prometheus.MustRegister(CommandsChecked)

	CommandsRemoved = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Help: "Counter of Commands that have ran ExecRemove",
			Name: "etcha_commands_removed_total",
		},
		[]string{"id", "status"},
	)
	prometheus.MustRegister(CommandsRemoved)
}
