package main

import (
	"log"
	"fmt"
	"net/http"
	"github.com/uber/jaeger-lib/metrics"

	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/opentracing/opentracing-go"
	"io"
)

func main() {
	closer, err := configureTracer()
	if err != nil {
		log.Printf("Could not initialize jaeger tracer: %s", err.Error())
		return
	}
	defer closer.Close()

	http.HandleFunc("/hello", helloHandler)
	http.ListenAndServe(":8080", nil)
}

func configureTracer() (io.Closer, error) {
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans: true,
		},
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Initialize tracer with configureTracer logger and configureTracer metrics factory
	return cfg.InitGlobalTracer(
		"Ducky",
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	sp := opentracing.StartSpan("hello")
	defer sp.Finish()
	fmt.Fprintf(w, buildHelloMessage(r, sp))
}
func buildHelloMessage(r *http.Request, parentSpan opentracing.Span) string {
	defer opentracing.StartSpan("buildHelloMsg",
		opentracing.ChildOf(parentSpan.Context())).Finish()
	return fmt.Sprintf("Hello from %s", r.Host)
}
