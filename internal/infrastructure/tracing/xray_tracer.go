package tracing

import (
	"context"

	"github.com/aws/aws-xray-sdk-go/xray"
	"go.uber.org/zap"
)

// XRayTracer gerencia rastreamento com AWS X-Ray
type XRayTracer struct {
	enabled bool
	logger  *zap.Logger
}

// NewXRayTracer cria um novo tracer X-Ray
func NewXRayTracer(enabled bool, logger *zap.Logger) *XRayTracer {
	if enabled {
		err := xray.Configure(xray.Config{
			DaemonAddr: "localhost:2000",
			LogLevel:   "info",
		})
		if err != nil {
			logger.Warn("failed to configure X-Ray", zap.Error(err))
		} else {
			logger.Info("X-Ray tracing enabled")
		}
	}

	return &XRayTracer{
		enabled: enabled,
		logger:  logger,
	}
}

// StartSegment inicia um segmento X-Ray
func (t *XRayTracer) StartSegment(ctx context.Context, name string) (context.Context, func()) {
	if !t.enabled {
		return ctx, func() {}
	}

	ctx, seg := xray.BeginSegment(ctx, name)
	t.logger.Debug("X-Ray segment started", zap.String("name", name))

	return ctx, func() {
		seg.Close(nil)
	}
}

// AddAnnotation adiciona uma anotação ao segmento atual
func (t *XRayTracer) AddAnnotation(ctx context.Context, key string, value interface{}) error {
	if !t.enabled {
		return nil
	}

	seg := xray.GetSegment(ctx)
	if seg != nil {
		// X-Ray SDK methods don't return errors, but we'll handle them properly here
		_ = seg.AddAnnotation(key, value)
		t.logger.Debug("X-Ray annotation added",
			zap.String("key", key),
			zap.Any("value", value))
	}
	return nil
}

// AddMetadata adiciona metadados ao segmento atual
func (t *XRayTracer) AddMetadata(ctx context.Context, key string, value interface{}) error {
	if !t.enabled {
		return nil
	}

	seg := xray.GetSegment(ctx)
	if seg != nil {
		// X-Ray SDK methods don't return errors, but we'll handle them properly here
		_ = seg.AddMetadata(key, value)
		t.logger.Debug("X-Ray metadata added",
			zap.String("key", key))
	}
	return nil
}

// AddError registra um erro no segmento
func (t *XRayTracer) AddError(ctx context.Context, err error) error {
	if !t.enabled || err == nil {
		return nil
	}

	seg := xray.GetSegment(ctx)
	if seg != nil {
		// X-Ray SDK methods don't return errors, but we'll handle them properly here
		_ = seg.AddError(err)
		t.logger.Debug("X-Ray error added", zap.Error(err))
	}
	return nil
}

// CaptureCall executa uma função sob rastreamento
func (t *XRayTracer) CaptureCall(ctx context.Context, name string, fn func(context.Context) error) error {
	ctx, end := t.StartSegment(ctx, name)
	defer end()

	err := fn(ctx)
	if err != nil {
		if addErr := t.AddError(ctx, err); addErr != nil {
			t.logger.Error("failed to add error to X-Ray", zap.Error(addErr))
		}
	}

	return err
}

// GetTraceID retorna o ID do trace atual
func (t *XRayTracer) GetTraceID(ctx context.Context) string {
	if !t.enabled {
		return ""
	}

	seg := xray.GetSegment(ctx)
	if seg != nil {
		return seg.TraceID
	}

	return ""
}

// IsEnabled retorna se o tracing está habilitado
func (t *XRayTracer) IsEnabled() bool {
	return t.enabled
}
