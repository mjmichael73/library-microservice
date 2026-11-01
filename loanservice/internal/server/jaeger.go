package server

import (
	"io"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

const (
	JAEGER_SERVICE_NAME = "loan-service"
	JAEGER_HOST_PORT    = "jaeger:6831"
	JAEGER_SAMPLE_RATE  = 1.0
)

// jaegerServiceName returns the JAEGER_SERVICE_NAME env override, falling back
// to the built-in default when unset.
func jaegerServiceName() string {
	if v := os.Getenv("JAEGER_SERVICE_NAME"); v != "" {
		return v
	}
	return JAEGER_SERVICE_NAME
}

// jaegerHostPort returns the JAEGER_HOST_PORT env override, falling back to the
// built-in default when unset.
func jaegerHostPort() string {
	if v := os.Getenv("JAEGER_HOST_PORT"); v != "" {
		return v
	}
	return JAEGER_HOST_PORT
}

// jaegerSampleRate returns the JAEGER_SAMPLE_RATE env override parsed as a
// float64, falling back to the default of 1.0 when unset or invalid.
func jaegerSampleRate() float64 {
	v := os.Getenv("JAEGER_SAMPLE_RATE")
	if v == "" {
		return JAEGER_SAMPLE_RATE
	}
	rate, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return JAEGER_SAMPLE_RATE
	}
	return rate
}

func (s *EchoServer) initJaeger() (io.Closer, error) {
	cfg := jaegercfg.Configuration{
		ServiceName: jaegerServiceName(),
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: jaegerSampleRate(),
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: jaegerHostPort(),
		},
	}
	jLogger := jaegerlog.StdLogger
	jMetricsFactory := metrics.NullFactory

	tracer, closer, err := cfg.NewTracer(
		jaegercfg.Logger(jLogger),
		jaegercfg.Metrics(jMetricsFactory),
	)
	if err != nil {
		return nil, err
	}
	opentracing.SetGlobalTracer(tracer)
	return closer, nil
}

func JaegerTracingMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			operationName := c.Request().Method + " " + c.Path()
			tracer := opentracing.GlobalTracer()

			wireContext, err := tracer.Extract(
				opentracing.HTTPHeaders,
				opentracing.HTTPHeadersCarrier(c.Request().Header),
			)

			var span opentracing.Span
			if err != nil {
				span = tracer.StartSpan(operationName)
			} else {
				span = tracer.StartSpan(operationName, opentracing.ChildOf(wireContext))
			}
			defer span.Finish()

			span.SetTag("http.method", c.Request().Method)
			span.SetTag("http.url", c.Request().RequestURI)
			span.SetTag("component", jaegerServiceName())

			ctx := opentracing.ContextWithSpan(c.Request().Context(), span)
			c.SetRequest(c.Request().WithContext(ctx))

			err = next(c)

			status := c.Response().Status
			span.SetTag("http.status_code", status)
			if status >= 500 {
				span.SetTag("error", true)
			}

			return err
		}
	}
}
