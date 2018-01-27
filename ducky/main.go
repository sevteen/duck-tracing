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
	"time"
)

const authHeaderToken = "X-Auth-Token"

var authHostPort = getEnv("AUTH_HOST_PORT", "localhost:9090")

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
	Owner                string
	Value                string
	CreatedAt            string
	Valid                bool
	BasicAuthHeaderValue string
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

	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	return cfg.InitGlobalTracer(
		getEnv("SERVICE_NAME", "Ducky"),
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
}

func handleDucks(w http.ResponseWriter, r *http.Request) {
	sp := opentracing.StartSpan("ducks")
	defer sp.Finish()
	if !isValidRequest(r, w, sp) {
		return
	}

	switch r.Method {
	case "GET":
		handleDuckGet(r, w, sp)
	case "POST":
		handleDuckPost(r, sp)
	}
}

func isValidRequest(r *http.Request, w http.ResponseWriter, parentSpan opentracing.Span) bool {
	validateSpan := opentracing.StartSpan(
		"validateRequest", opentracing.ChildOf(parentSpan.Context())).
		SetTag("URL", r.URL).SetTag("METHOD", r.Method)
	defer validateSpan.Finish()
	validRequest := true
	if !hasTokenHeader(r) {
		w.WriteHeader(401)
		writeString(w, "http://"+authHostPort+"/tokens?authHeaderName="+authHeaderToken)
		validRequest = false
	} else {
		token := r.Header.Get(authHeaderToken)
		fetchTokenSpan := opentracing.StartSpan(
			"fetchToken", opentracing.ChildOf(validateSpan.Context())).
			SetTag("authServerAddress", authHostPort)
		resp, err := fetchToken(token, fetchTokenSpan)
		defer fetchTokenSpan.Finish()

		if err != nil {
			log.Fatalf("Failed to fetch token %s", err.Error())
			validRequest = false
		} else {
			var t Token
			deserialize(resp.Body, &t)
			if !isTokenValid(t, fetchTokenSpan) {
				log.Printf("Token %s is not valid", token)
				w.WriteHeader(403)
				validRequest = false
			}
		}
	}
	return validRequest
}

func isTokenValid(t Token, parentSpan opentracing.Span) bool {
	sp := opentracing.StartSpan(
		"validateToken", opentracing.ChildOf(parentSpan.Context())).
			SetTag("token", t.Value).SetTag("owner", t.Owner)

	defer sp.Finish()
	time.Sleep(1500 * time.Millisecond)
	return t.Valid
}

func fetchToken(token string, fetchTokenSpan opentracing.Span) (*http.Response, error) {
	request, _ := http.NewRequest("GET", "http://"+authHostPort+"/tokens/"+token, nil)

	carrier := opentracing.HTTPHeadersCarrier(request.Header)
	opentracing.GlobalTracer().Inject(
		fetchTokenSpan.Context(),
		opentracing.HTTPHeaders,
		carrier)

	return (&http.Client{}).Do(request)
}

func handleDuckGet(r *http.Request, w http.ResponseWriter, parentSpan opentracing.Span) {
	sp := opentracing.StartSpan(
		"getDucks", opentracing.ChildOf(parentSpan.Context()))
	name := r.URL.Query().Get("name")
	if len(name) > 0 {
		sp.SetTag("name", name)
	}
	defer sp.Finish()
	ducks := getDucks(r)
	writeJson(w, ducks, sp)
}

func getDucks(r *http.Request) []Duck {
	var ducks []Duck
	name := r.URL.Query().Get("name")
	if len(name) > 0 {
		duck, _ := duckRepository.GetByName(name)
		if duck != nil {
			ducks = make([]Duck, 1)
			ducks[0] = *duck
		} else {
			ducks = make([]Duck, 0)
		}
	} else {
		ducks, _ = duckRepository.GetAll()
	}
	return ducks
}

func handleDuckPost(r *http.Request, parentSpan opentracing.Span) {
	sp := opentracing.StartSpan(
		"addDuck", opentracing.ChildOf(parentSpan.Context()))
	defer sp.Finish()
	duckRepository.Add(parseDuck(r.Body, sp))
}

func hasTokenHeader(r *http.Request) bool {
	return len(r.Header.Get(authHeaderToken)) > 0
}

func writeJson(w http.ResponseWriter, o interface{}, parentSpan opentracing.Span) {
	if parentSpan != nil {
		serializationSpan := opentracing.StartSpan(
			"jsonSerialization", opentracing.ChildOf(parentSpan.Context()))
		defer serializationSpan.Finish()
	}
	time.Sleep(200 * time.Millisecond)
	m, _ := json.Marshal(o)
	w.Header().Set("Content-Type", "application/json; utf-8")
	fmt.Fprint(w, string(m))
}

func writeString(w http.ResponseWriter, msg string) {
	w.Write([]byte(msg))
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
