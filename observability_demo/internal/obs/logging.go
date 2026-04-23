package obs

import (
	"context"
	"log"
	"log/slog"
	"os"

	"encroach-demo/internal/config"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func InitLogging(level config.ObsLevel, endpoint string, headers map[string]string) (*slog.Logger, func()) {
	if !level.Has(config.ObsLogs) {
		handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
		logger := slog.New(handler)
		slog.SetDefault(logger)
		return logger, func() {}
	}

	opts := []otlploghttp.Option{
		otlploghttp.WithEndpoint(endpoint),
		otlploghttp.WithURLPath("/api/default/v1/logs"),
		otlploghttp.WithInsecure(),
	}
	if len(headers) > 0 {
		opts = append(opts, otlploghttp.WithHeaders(headers))
	}

	exporter, err := otlploghttp.New(context.Background(), opts...)
	if err != nil {
		log.Fatalf("log exporter: %v", err)
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName("encroach"),
			semconv.ServiceVersion("demo"),
		),
	)
	if err != nil {
		log.Fatalf("log resource: %v", err)
	}

	provider := sdklog.NewLoggerProvider(
		sdklog.WithProcessor(sdklog.NewBatchProcessor(exporter)),
		sdklog.WithResource(res),
	)

	otelHandler := otelslog.NewHandler("encroach", otelslog.WithLoggerProvider(provider))
	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})

	logger := slog.New(fanoutHandler{jsonHandler, otelHandler})
	slog.SetDefault(logger)

	return logger, func() {
		if err := provider.Shutdown(context.Background()); err != nil {
			log.Printf("log provider shutdown: %v", err)
		}
	}
}

type fanoutHandler struct {
	a, b slog.Handler
}

func (f fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return f.a.Enabled(ctx, level) || f.b.Enabled(ctx, level)
}

func (f fanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	if f.a.Enabled(ctx, r.Level) {
		_ = f.a.Handle(ctx, r.Clone())
	}
	if f.b.Enabled(ctx, r.Level) {
		_ = f.b.Handle(ctx, r.Clone())
	}
	return nil
}

func (f fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return fanoutHandler{f.a.WithAttrs(attrs), f.b.WithAttrs(attrs)}
}

func (f fanoutHandler) WithGroup(name string) slog.Handler {
	return fanoutHandler{f.a.WithGroup(name), f.b.WithGroup(name)}
}
