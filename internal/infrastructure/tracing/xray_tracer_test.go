package tracing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

const (
	testKeyString contextKey = "key"
)

func TestNewXRayTracerDisabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	require.NotNil(t, tracer)
	assert.False(t, tracer.IsEnabled())
}

func TestNewXRayTracerEnabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(true, logger)

	require.NotNil(t, tracer)
	assert.True(t, tracer.IsEnabled())
}

func TestStartSegment(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()
	newCtx, end := tracer.StartSegment(ctx, "test-segment")

	require.NotNil(t, end)
	assert.Equal(t, ctx, newCtx)
	end()
}

func TestStartSegmentMultipleTimes(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()

	ctx1, end1 := tracer.StartSegment(ctx, "segment-1")
	assert.NotNil(t, end1)

	_, end2 := tracer.StartSegment(ctx1, "segment-2")
	assert.NotNil(t, end2)

	end2()
	end1()
}

func TestCaptureCall(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()
	called := false

	err := tracer.CaptureCall(ctx, "test-call", func(c context.Context) error {
		called = true
		return nil
	})

	require.NoError(t, err)
	assert.True(t, called)
}

func TestCaptureCallWithError(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()
	testErr := assert.AnError

	err := tracer.CaptureCall(ctx, "test-call-error", func(c context.Context) error {
		return testErr
	})

	require.Error(t, err)
	assert.Equal(t, testErr, err)
}

func TestAddAnnotationAndMetadata(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()

	if err := tracer.AddAnnotation(ctx, "key", "value"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tracer.AddMetadata(ctx, "key", "value"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tracer.AddError(ctx, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetTraceID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()
	traceID := tracer.GetTraceID(ctx)

	assert.Equal(t, "", traceID)
}

func TestGetTraceIDEnabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(true, logger)

	ctx := context.Background()
	traceID := tracer.GetTraceID(ctx)

	assert.Equal(t, "", traceID)
}

func TestAddAnnotationWithDifferentTypes(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()

	if err := tracer.AddAnnotation(ctx, "string_key", "string_value"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tracer.AddAnnotation(ctx, "int_key", 42); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tracer.AddAnnotation(ctx, "bool_key", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tracer.AddAnnotation(ctx, "float_key", 3.14); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddMetadataWithDifferentTypes(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()

	if err := tracer.AddMetadata(ctx, "metadata_string", "value"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tracer.AddMetadata(ctx, "metadata_map", map[string]interface{}{
		"nested": "value",
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddErrorWithNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()

	if err := tracer.AddError(ctx, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCaptureCallContextPropagation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()
	var capturedCtx context.Context

	err := tracer.CaptureCall(ctx, "test-call", func(c context.Context) error {
		capturedCtx = c
		return nil
	})

	require.NoError(t, err)
	assert.NotNil(t, capturedCtx)
}

func TestAddErrorWithValue(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()
	err := assert.AnError

	if addErr := tracer.AddError(ctx, err); addErr != nil {
		t.Fatalf("unexpected error: %v", addErr)
	}
}

func TestStartSegmentDisabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()

	_, end := tracer.StartSegment(ctx, "test")
	assert.NotNil(t, end)
}

func TestCaptureCallMultiple(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()
	count := 0

	for i := 0; i < 3; i++ {
		if err := tracer.CaptureCall(ctx, "call-"+string(rune(48+i)), func(c context.Context) error {
			count++
			return nil
		}); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	assert.Equal(t, 3, count)
}

func TestAddAnnotationMultipleTypes(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()
	_, end := tracer.StartSegment(ctx, "test-annot")

	// Test annotations don't panic
	if err := tracer.AddAnnotation(ctx, "string_key", "value"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tracer.AddAnnotation(ctx, "int_key", 42); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tracer.AddAnnotation(ctx, "bool_key", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tracer.AddAnnotation(ctx, "float_key", 3.14); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	end()
}

func TestAddMetadataMultipleKeys(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()
	_, end := tracer.StartSegment(ctx, "test-meta")

	metadata := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
		"nested": map[string]interface{}{
			"deep": "value",
		},
	}

	for key, value := range metadata {
		if err := tracer.AddMetadata(ctx, key, value); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	end()
}

func TestCaptureCallWithDifferentErrors(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()

	t.Run("capture call with error", func(t *testing.T) {
		err := tracer.CaptureCall(ctx, "error-call", func(c context.Context) error {
			return fmt.Errorf("test error")
		})
		// When disabled, error is returned
		_ = err
	})

	t.Run("capture call with nil error", func(t *testing.T) {
		err := tracer.CaptureCall(ctx, "nil-call", func(c context.Context) error {
			return nil
		})
		assert.NoError(t, err)
	})
}

func TestTracerStateConsistency(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	t.Run("enabled tracer state", func(t *testing.T) {
		tracer := NewXRayTracer(true, logger)
		assert.True(t, tracer.IsEnabled())
	})

	t.Run("disabled tracer state", func(t *testing.T) {
		tracer := NewXRayTracer(false, logger)
		assert.False(t, tracer.IsEnabled())
	})
}

func TestCaptureCallContextPreservation(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	originalCtx := context.WithValue(context.Background(), testKeyString, "value")

	err := tracer.CaptureCall(originalCtx, "preserve-call", func(c context.Context) error {
		assert.Equal(t, "value", c.Value(testKeyString))
		return nil
	})

	assert.NoError(t, err)
}

func TestSegmentEndMultipleCalls(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()
	_, end := tracer.StartSegment(ctx, "segment")

	// Multiple calls to end should not panic
	end()
	end()
	end()
}

func TestTracerWithMultipleSegments(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()

	ctx1, end1 := tracer.StartSegment(ctx, "seg1")
	ctx2, end2 := tracer.StartSegment(ctx1, "seg2")
	ctx3, end3 := tracer.StartSegment(ctx2, "seg3")

	if err := tracer.AddAnnotation(ctx3, "depth", 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	end3()
	end2()
	end1()
}

func TestTracerEnabledStateManagement(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	disabledTracer := NewXRayTracer(false, logger)
	assert.False(t, disabledTracer.IsEnabled())

	enabledTracer := NewXRayTracer(true, logger)
	assert.True(t, enabledTracer.IsEnabled())
}

func TestSegmentOperationsSequence(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()

	// Sequence of operations
	ctx1, end1 := tracer.StartSegment(ctx, "operation1")
	if err := tracer.AddAnnotation(ctx1, "step", 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tracer.AddMetadata(ctx1, "metadata", map[string]interface{}{"key": "value"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	end1()

	ctx2, end2 := tracer.StartSegment(ctx, "operation2")
	if err := tracer.AddAnnotation(ctx2, "step", 2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	end2()

	ctx3, end3 := tracer.StartSegment(ctx, "operation3")
	if err := tracer.AddAnnotation(ctx3, "step", 3); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	end3()
}

func TestCaptureCallReturnValue(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()

	t.Run("capture call returns nil error for nil return", func(t *testing.T) {
		err := tracer.CaptureCall(ctx, "test-func", func(c context.Context) error {
			return nil
		})
		assert.NoError(t, err)
	})

	t.Run("capture call returns error for error return", func(t *testing.T) {
		err := tracer.CaptureCall(ctx, "test-func", func(c context.Context) error {
			return fmt.Errorf("operation failed")
		})
		assert.Error(t, err)
	})
}

func TestTracerAnnotationEdgeCases(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	ctx := context.Background()
	_, end := tracer.StartSegment(ctx, "test")

	// Add various annotation types
	err := tracer.AddAnnotation(ctx, "nil_value", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = tracer.AddAnnotation(ctx, "empty_string", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = tracer.AddAnnotation(ctx, "zero_int", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = tracer.AddAnnotation(ctx, "false_bool", false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	end()
}

func TestTracerWithDifferentContexts(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(false, logger)

	t.Run("background context", func(t *testing.T) {
		ctx := context.Background()
		_, end := tracer.StartSegment(ctx, "bg-segment")
		end()
	})

	t.Run("with value context", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), testKeyString, "value")
		_, end := tracer.StartSegment(ctx, "value-segment")
		end()
	})

	t.Run("with timeout context", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, end := tracer.StartSegment(ctx, "timeout-segment")
		end()
	})
}

// Additional tests for enabled tracer to improve coverage
func TestAddAnnotationWhenEnabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(true, logger)

	ctx := context.Background()
	_, end := tracer.StartSegment(ctx, "test-segment")

	// These should work without panicking when enabled
	err := tracer.AddAnnotation(ctx, "test_key", "test_value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = tracer.AddAnnotation(ctx, "number", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	end()
}

func TestAddMetadataWhenEnabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(true, logger)

	ctx := context.Background()
	_, end := tracer.StartSegment(ctx, "test-segment")

	err := tracer.AddMetadata(ctx, "test_key", "test_value")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = tracer.AddMetadata(ctx, "complex", map[string]interface{}{"nested": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	end()
}

func TestAddErrorWhenEnabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(true, logger)

	ctx := context.Background()
	_, end := tracer.StartSegment(ctx, "test-segment")

	if err := tracer.AddError(ctx, fmt.Errorf("test error")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tracer.AddError(ctx, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	end()
}

func TestStartSegmentWhenEnabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(true, logger)

	ctx := context.Background()
	newCtx, end := tracer.StartSegment(ctx, "enabled-segment")

	// Context should be modified when enabled (X-Ray adds segment to context)
	assert.NotNil(t, newCtx)
	assert.NotNil(t, end)
	end()
}

func TestCaptureCallWhenEnabled(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(true, logger)

	ctx := context.Background()
	called := false

	err := tracer.CaptureCall(ctx, "enabled-call", func(c context.Context) error {
		called = true
		return nil
	})

	assert.NoError(t, err)
	assert.True(t, called)
}

func TestGetTraceIDWhenSegmentIsNil(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	tracer := NewXRayTracer(true, logger)

	ctx := context.Background()
	// When segment is nil in the context, GetTraceID should return empty string
	traceID := tracer.GetTraceID(ctx)
	assert.Equal(t, "", traceID)
}
