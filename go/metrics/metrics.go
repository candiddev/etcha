// Package metrics contains metrics for Etcha.
package metrics

import (
	"context"

	"github.com/candiddev/shared/go/logger"
	"github.com/candiddev/shared/go/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

//nolint:revive
type (
	CommandMode   string
	SourceTrigger string
)

// Various enums.
const (
	CommandModeChange    CommandMode   = "change"
	CommandModeCheck     CommandMode   = "check"
	CommandModeRemove    CommandMode   = "remove"
	SourceTriggerEvent   SourceTrigger = "event"
	SourceTriggerPull    SourceTrigger = "pull"
	SourceTriggerPush    SourceTrigger = "push"
	SourceTriggerWebhook SourceTrigger = "webhook"
)

func GetCommandID(ctx context.Context) string {
	return logger.GetAttribute(ctx, "commandID")
}

func SetCommandID(ctx context.Context, command string) context.Context {
	return logger.SetAttribute(ctx, "commandID", command)
}

func GetCommandMode(ctx context.Context) string {
	return logger.GetAttribute(ctx, "commandMode")
}

func SetCommandMode(ctx context.Context, mode CommandMode) context.Context {
	return logger.SetAttribute(ctx, "commandMode", string(mode))
}

func GetSourceName(ctx context.Context) string {
	return logger.GetAttribute(ctx, "sourceName")
}

func SetSourceName(ctx context.Context, source string) context.Context {
	return logger.SetAttribute(ctx, "sourceName", source)
}

func GetSourceTrigger(ctx context.Context) string {
	return logger.GetAttribute(ctx, "sourceTrigger")
}

func SetSourceTrigger(ctx context.Context, trigger SourceTrigger) context.Context {
	return logger.SetAttribute(ctx, "sourceTrigger", trigger)
}

//nolint:gochecknoglobals
var (
	commands        *prometheus.CounterVec
	sources         *prometheus.CounterVec
	sourcesCommands *prometheus.GaugeVec
)

func init() { //nolint:gochecknoinits
	commands = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Help: "Counter of Commands that have ran",
			Name: "etcha_commands_total",
		},
		[]string{"error", "id", "mode", "source"},
	)
	prometheus.MustRegister(commands)

	sources = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Help: "Counter of Sources that have ran",
			Name: "etcha_sources_total",
		},
		[]string{"error", "name", "trigger"},
	)
	prometheus.MustRegister(sources)

	sourcesCommands = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Help: "Gauage of Source commands that have ran",
			Name: "etcha_sources_commands",
		},
		[]string{"mode", "name", "trigger"},
	)
	prometheus.MustRegister(sourcesCommands)
}

func CollectCommands(ctx context.Context, err bool) {
	commands.WithLabelValues(metrics.ParseBool(err), GetCommandID(ctx), GetCommandMode(ctx), GetSourceName(ctx)).Inc()
}

func CollectSources(ctx context.Context, err bool) {
	sources.WithLabelValues(metrics.ParseBool(err), GetSourceName(ctx), GetSourceTrigger(ctx)).Inc()
}

func CollectSourcesCommands(ctx context.Context, value int) {
	sourcesCommands.WithLabelValues(GetCommandMode(ctx), GetSourceName(ctx), GetSourceTrigger(ctx)).Set(float64(value))
}
