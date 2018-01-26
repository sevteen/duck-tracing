package main

import (
	"os"
	"log"
	"net/http"
	"github.com/uber/jaeger-lib/metrics"

	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/opentracing/opentracing-go"
	"io"
	"fmt"
	"encoding/json"
)

const outHostPort string  = "";

const urlToRedirect string  = "";

type Duck struct {
	Name string
}

type DuckRepository interface {
	GetAll() ([]Duck, error)

	GetByName(name string) (*Duck, error)

	Add(duck Duck) error
}

type InMemoryDuckRepository struct {
	Ducks map[string]Duck
}

type Token struct {
	Owner string
	Value string
	CreatedAt string
	Valid bool
	basicAuthHeaderValue string
}
func (r InMemoryDuckRepository) GetAll() ([]Duck, error) {
	ducks := make([]Duck, len(r.Ducks))
	i := 0
	for _, d := range r.Ducks {
		ducks[i] = d
		i++
	}
	return ducks, nil
}

func (r InMemoryDuckRepository) GetByName(name string) (*Duck, error) {
	duck, ok := r.Ducks[name]
	if ok {
		return &duck, nil
	}
	return nil, nil
}

func (r InMemoryDuckRepository) Add(d Duck) error {
	r.Ducks[d.Name] = d
	return nil
}

var duckRepository DuckRepository = InMemoryDuckRepository{make(map[string]Duck)}

func main() {
	closer, err := configureTracer()
	if err != nil {
		log.Printf("Could not initialize jaeger tracer: %s", err.Error())
		return
	}
	defer closer.Close()

	setupServer()
}

func setupServer() {
	http.HandleFunc("/ducks", handleDucks)
	http.ListenAndServe(":8080", nil)
}

func configureTracer() (io.Closer, error) {
	cfg := jaegercfg.Configuration{
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: getEnv("JAEGER_AGENT_HOST_PORT", "localhost:6831"),
		},
	}

	// Example logger and metrics factory. Use github.com/uber/jaeger-client-go/log
	// and github.com/uber/jaeger-lib/metrics respectively to bind to real logging and metrics
	// frameworks.
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	// Initialize tracer with configureTracer logger and configureTracer metrics factory
	return cfg.InitGlobalTracer(
		getEnv("SERVICE_NAME", "Ducky"),
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
}

func handleDucks(w http.ResponseWriter, r *http.Request) {
	if !hasTokenHeader(r) {
		http.Redirect(w, r, urlToRedirect, 302)
	} else {
		resp, err := http.Get("http://" + outHostPort + "/tokens/" + r.Header.Get("X-Auth-Token"))
		var t Token
		if err != nil {
			deserialize(resp.Body, &t)
			if !t.Valid {http.RedirectHandler(outHostPort, 403)}
		} else  {
			log.Fatal(err.Error())
		}
	}

	switch r.Method {
	case "GET":
		var ducks []Duck
		var err error
		var sp opentracing.Span

		name := r.URL.Query().Get("name")
		if len(name) > 0 {
			sp = opentracing.StartSpan("getDuckByName").SetTag("name", name)
			defer sp.Finish()
			var duck *Duck
			duck, _ = duckRepository.GetByName(name)
			if duck != nil {
				ducks = make([]Duck, 1)
				ducks[0] = *duck
			}
		} else {
			sp = opentracing.StartSpan("getDucks")
			defer sp.Finish()
			ducks, err = duckRepository.GetAll()
		}
		if err != nil {
			writeError(w, err)
		}
		if ducks == nil {
			ducks = make([]Duck, 0)
		}
		writeJson(w, ducks, sp)
	case "POST":
		sp := opentracing.StartSpan("addDuck")
		defer sp.Finish()
		duckRepository.Add(parseDuck(r.Body, sp))
	}
}
func hasTokenHeader(r *http.Request) bool {
	return len(r.Header.Get("X-Auth-Token")) > 0
}

func writeError(w http.ResponseWriter, err error) {
	writeJson(w, err.Error(), nil)
}

func writeJson(w http.ResponseWriter, o interface{}, parentSpan opentracing.Span) {
	if parentSpan != nil {
		serializationSpan := opentracing.StartSpan(
			"jsonSerialization", opentracing.ChildOf(parentSpan.Context()))
		defer serializationSpan.Finish()
	}
	m, _ := json.Marshal(o)
	w.Header().Set("Content-Type", "application/json; utf-8")
	fmt.Fprint(w, string(m))
}

func parseDuck(rawDuck io.ReadCloser, parentSpan opentracing.Span) Duck {
	if parentSpan != nil {
		deserializationSpan := opentracing.StartSpan(
			"jsonDeserialization", opentracing.ChildOf(parentSpan.Context()))
		defer deserializationSpan.Finish()
	}
	var d Duck
	deserialize(rawDuck, &d)
	return d
}

func deserialize(r io.ReadCloser, d interface{}) {
	decoder := json.NewDecoder(r)
	decoder.Decode(&d)
	defer r.Close()
}

func getEnv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		value = defaultValue
	}
	return value
}
