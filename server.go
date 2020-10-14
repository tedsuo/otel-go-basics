package main

import (
	"errors"
	"net/http"
	"time"

	"github.com/lightstep/otel-launcher-go/launcher"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/label"
	"go.opentelemetry.io/otel/semconv"
)

var tracer = global.Tracer("hello-server")

func helloHandler(w http.ResponseWriter, req *http.Request) {
	cxt := req.Context()
	span := trace.SpanFromContext(cxt)
	span.SetAttributes(semconv.HTTPRouteKey.String("hello"))

	cxt, span = tracer.Start(cxt, "my-server-span")
	defer span.End()

	projectID := otel.BaggageValue(cxt, "ProjectID")
	span.SetAttributes(label.KeyValue{Key: "ProjectID", Value: projectID})
	span.RecordError(cxt, errors.New("ooops"), trace.WithErrorStatus(codes.Error))
	span.AddEvent(cxt, "writing response", label.String("content", "Hello World"))

	time.Sleep(time.Second)

	w.Write([]byte("Hello World"))
}

func main() {
	otel := launcher.ConfigureOpentelemetry(
		launcher.WithServiceName("hello-server-4"),
		launcher.WithAccessToken("05137244f31c795a90417deceef6bc28"),
		launcher.WithPropagators([]string{"b3", "cc"}),
		launcher.WithResourceAttributes(map[string]string{
			"something":                      "else",
			string(semconv.ContainerNameKey): "my-container-name",
		}),
	)
	defer otel.Shutdown()

	wrappedHandler := otelhttp.NewHandler(http.HandlerFunc(helloHandler), "/hello")
	http.Handle("/hello", wrappedHandler)
	http.ListenAndServe(":9000", nil)
}
