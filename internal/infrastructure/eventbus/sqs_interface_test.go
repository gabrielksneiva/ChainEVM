package eventbus

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/stretchr/testify/assert"
)

func TestNewSQSAdapter(t *testing.T) {
	t.Parallel()

	client := &sqs.Client{}
	adapter := NewSQSAdapter(client)

	assert.NotNil(t, adapter)

	// Verificar que o adapter implementa a interface
	_, ok := interface{}(adapter).(SQSClient)
	assert.True(t, ok, "adapter should implement SQSClient interface")
}
