package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/lightstep/otel-launcher-go/launcher"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/semconv"
	"go.opentelemetry.io/otel/trace"
)

// Create one tracer per package
// NOTE: You only need a tracer if you are creating your own spans
var tracer trace.Tracer

func init() {
	// Name the tracer after the package, or the service if you are in main
	tracer = otel.Tracer("hello-server")
}

func main() {
	// Create an OpenTelemetry SDK using the launcher
	otel := launcher.ConfigureOpentelemetry(
		launcher.WithServiceName("hello-server-4"),
		launcher.WithAccessToken("ACCESS TOKEN"),
		launcher.WithPropagators([]string{"b3", "baggage"}),
		launcher.WithResourceAttributes(map[string]string{
			"something":                      "else",
			string(semconv.ContainerNameKey): "my-container-name",
		}),
	)
	// Shut down the SDK to flush any remaining data before the program exits
	defer otel.Shutdown()

	wrappedHandler := otelhttp.NewHandler(http.HandlerFunc(helloHandler), "/hello")
	http.Handle("/hello", wrappedHandler)
	http.ListenAndServe(":9000", nil)
}

// Example HTTP Handler
func helloHandler(w http.ResponseWriter, req *http.Request) {
	cxt := req.Context()
	span := trace.SpanFromContext(cxt)
	span.SetAttributes(semconv.HTTPRouteKey.String("hello"))

	cxt, span = tracer.Start(cxt, "my-server-span")
	defer span.End()

	projectID := baggage.Value(cxt, "ProjectID")
	span.SetAttributes(label.KeyValue{Key: "ProjectID", Value: projectID})
	span.RecordError(errors.New("ooops"))
	span.AddEvent("writing response", trace.WithAttributes(
		label.String("hello", "world"),
		label.Int("answer", 42),
	))

	time.Sleep(time.Second)

	w.Write([]byte("Hello World"))
}
