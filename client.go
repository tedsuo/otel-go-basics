package main

import (
	"context"
	"net/http"

	"github.com/lightstep/otel-launcher-go/launcher"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/label"
)

var t = global.Tracer("hello-client")

func main() {
	otel := launcher.ConfigureOpentelemetry(
		launcher.WithServiceName("hello-client"),
		launcher.WithAccessToken("05137244f31c795a90417deceef6bc28"),
		launcher.WithPropagators([]string{"b3", "cc"}),
	)
	defer otel.Shutdown()
	ctx, span := t.Start(context.Background(), "all requests")
	defer span.End()

	for i := 0; i < 5; i++ {
		makeRequest(ctx)
	}
}

func makeRequest(ctx context.Context) {
	client := http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	ctx = otel.ContextWithBaggageValues(ctx, label.String("ProjectID", "456"))
	bags := otel.Baggage(ctx)

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
