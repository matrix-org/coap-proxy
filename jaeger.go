package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

// Env flags
var jaegerHost, useJaeger = os.LookupEnv("SYNAPSE_JAEGER_HOST")

// setupJaegerTracing is a function that sets up OpenTracing with a
// configuration specific to the CoAP proxy.
func setupJaegerTracing() io.Closer {
	jaegerHost := jaegerHost
	serverName := os.Getenv("SYNAPSE_SERVER_NAME")

	serviceName := fmt.Sprintf("proxy-%s", serverName)

	var cfg jaegercfg.Configuration

	if useJaeger {
		cfg = jaegercfg.Configuration{
			Sampler: &jaegercfg.SamplerConfig{
				Type:  jaeger.SamplerTypeConst,
				Param: 1,
			},
			Reporter: &jaegercfg.ReporterConfig{
				LogSpans:           true,
				LocalAgentHostPort: fmt.Sprintf("%s:6831", jaegerHost),
			},
			ServiceName: serviceName,
		}
	} else {
		cfg = jaegercfg.Configuration{}
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Initialize tracer with a logger and a metrics factory
	closer, err := cfg.InitGlobalTracer(
		serviceName,
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		log.Printf("Could not initialize jaeger tracer: %s", err.Error())
		return nil
	}

	return closer
}
