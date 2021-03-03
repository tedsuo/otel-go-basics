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
		launcher.WithServiceVersion("1.3"),
		launcher.WithAccessToken("PS4AExeR1V9kLHqtSIk6to3okwIEQ+4DZPEFmJAWQl+wMisnXWuYGAXjiERYT5pCIDV2nAY4eOYMau/r2iQ="),
		launcher.WithPropagators([]string{"b3", "baggage"}),
		launcher.WithResourceAttributes(map[string]string{
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

	// You can create child spans from the span in the current context.
	// The returned context contains the new child span
	cxt, span = tracer.Start(cxt, "my-child-span")
	defer span.End()

	// Use baggage to get the projectID set by the client
	projectID := baggage.Value(cxt, "ProjectID")

	// Set attributes to add indexes to you spans. This helps you group and
	// compare traces.
	span.SetAttributes(label.KeyValue{Key: "ProjectID", Value: projectID})

	// Adding events is the tracing equivalent to logging.
	// These events will show up as logs on the span.
	span.AddEvent("writing response", trace.WithAttributes(
		label.String("hello", "world"),
		label.Int("answer", 42),
	))

	// Errors can be recorded as events
	span.RecordError(errors.New("ooops"))

	time.Sleep(time.Second)

	w.Write([]byte("Hello World"))
}
