package main

import (
	"context"
	"net/http"

	"github.com/lightstep/otel-launcher-go/launcher"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Create one tracer per package
// NOTE: You only need a tracer if you are creating your own spans
var tracer trace.Tracer

func init() {
	// Name the tracer after the package, or the service if you are in main
	tracer = otel.Tracer("hello-client")
}

func main() {
	// Create an OpenTelemetry SDK using the launcher
	launcher := launcher.ConfigureOpentelemetry(
		launcher.WithServiceName("hello-client"),
		launcher.WithAccessToken("ACCESS TOKEN"),
		launcher.WithPropagators([]string{"b3", "baggage"}),
	)
	// Shut down the SDK to flush any remaining data before the program exits
	defer launcher.Shutdown()

	// Create a root span. This will be the parent of every span created
	// using this context
	ctx, span := tracer.Start(context.Background(), "all requests")
	defer span.End()

	for i := 0; i < 5; i++ {
		// These requests will join the trace as child spans
		makeRequest(ctx)
	}
}

func makeRequest(ctx context.Context) {
	// Trace an HTTP client by wrapping the transport
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	// All requests will now create spans.Make sure you pass the context correctly.
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:9000/hello", nil)
	if err != nil {
		panic(err)
	}

	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	res.Body.Close()
}
