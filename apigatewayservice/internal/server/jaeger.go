package server

import (
	"io"

	"github.com/labstack/echo/v4"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

const (
	JAEGER_SERVICE_NAME = "apigateway-service"
	JAEGER_HOST_PORT = "jaeger:6831"
)

func (s *EchoServer) initJaeger() (io.Closer, error) {
	cfg := jaegercfg.Configuration{
		ServiceName: JAEGER_SERVICE_NAME,
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:           true,
			LocalAgentHostPort: JAEGER_HOST_PORT,
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
			span.SetTag("component", JAEGER_SERVICE_NAME)

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
