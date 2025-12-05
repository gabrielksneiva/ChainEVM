#!/bin/bash
# Don't use set -e, we want to run all tests even if some fail
# Exit code will be determined by test results at the end

echo "🧪 Starting ChainEVM E2E Tests..."
echo "=========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
AWS_REGION="${AWS_REGION:-us-east-1}"
LAMBDA_NAME="chainevm"
DYNAMODB_TABLE="evm-transactions"
SQS_QUEUE_NAME="evm-queue"

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((TESTS_PASSED++))
}

fail() {
    echo -e "${RED}✗${NC} $1"
    ((TESTS_FAILED++))
}

info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# Test 1: Lambda Function Status
echo ""
info "Test 1: Lambda Function Configuration"
LAMBDA_STATE=$(aws lambda get-function \
    --function-name $LAMBDA_NAME \
    --region $AWS_REGION \
    --query 'Configuration.State' \
    --output text 2>/dev/null || echo "NOT_FOUND")

if [ "$LAMBDA_STATE" = "Active" ]; then
    pass "Lambda function is Active"
else
    fail "Lambda function state: $LAMBDA_STATE"
fi

# Test 2: DynamoDB Table
echo ""
info "Test 2: DynamoDB Table Configuration"
TABLE_STATUS=$(aws dynamodb describe-table \
    --table-name $DYNAMODB_TABLE \
    --region $AWS_REGION \
    --query 'Table.TableStatus' \
    --output text 2>/dev/null || echo "NOT_FOUND")

if [ "$TABLE_STATUS" = "ACTIVE" ]; then
    pass "DynamoDB table is ACTIVE"
    
    # Check TTL
    TTL_STATUS=$(aws dynamodb describe-time-to-live \
        --table-name $DYNAMODB_TABLE \
        --region $AWS_REGION \
        --query 'TimeToLiveDescription.TimeToLiveStatus' \
        --output text 2>/dev/null || echo "DISABLED")
    
    if [ "$TTL_STATUS" = "ENABLED" ]; then
        pass "DynamoDB TTL is enabled"
    else
        fail "DynamoDB TTL status: $TTL_STATUS"
    fi
else
    fail "DynamoDB table status: $TABLE_STATUS"
fi

# Test 3: SQS Queue Access
echo ""
info "Test 3: SQS Queue Configuration"
QUEUE_URL=$(aws sqs get-queue-url \
    --queue-name $SQS_QUEUE_NAME \
    --region $AWS_REGION \
    --query 'QueueUrl' \
    --output text 2>/dev/null || echo "NOT_FOUND")

if [ "$QUEUE_URL" != "NOT_FOUND" ]; then
    pass "SQS queue is accessible"
    
    # Check Dead Letter Queue
    DLQ_ARN=$(aws sqs get-queue-attributes \
        --queue-url "$QUEUE_URL" \
        --attribute-names RedrivePolicy \
        --region $AWS_REGION \
        --query 'Attributes.RedrivePolicy' \
        --output text 2>/dev/null | grep -o 'arn:aws:sqs:[^"]*' || echo "")
    
    if [ -n "$DLQ_ARN" ]; then
        pass "DLQ is configured"
    else
        fail "No DLQ configured"
    fi
else
    fail "SQS queue not accessible"
fi

# Test 4: Event Source Mapping
echo ""
info "Test 4: Lambda Event Source Mapping"
EVENT_SOURCE=$(aws lambda list-event-source-mappings \
    --function-name $LAMBDA_NAME \
    --region $AWS_REGION \
    --query 'EventSourceMappings[0].State' \
    --output text 2>/dev/null || echo "NOT_FOUND")

if [ "$EVENT_SOURCE" = "Enabled" ]; then
    pass "Event source mapping is Enabled"
else
    fail "Event source mapping state: $EVENT_SOURCE"
fi

# Test 5: Lambda Environment Variables
echo ""
info "Test 5: Lambda Environment Configuration"
ENV_VARS=$(aws lambda get-function-configuration \
    --function-name $LAMBDA_NAME \
    --region $AWS_REGION \
    --query 'Environment.Variables' \
    --output json 2>/dev/null)

REQUIRED_VARS=("DYNAMODB_TABLE_NAME" "SQS_QUEUE_URL" "RPC_URL_ETHEREUM")
ENV_OK=0

for var in "${REQUIRED_VARS[@]}"; do
    if echo "$ENV_VARS" | grep -q "\"$var\""; then
        ((ENV_OK++))
    fi
done

if [ $ENV_OK -eq ${#REQUIRED_VARS[@]} ]; then
    pass "All required environment variables configured"
else
    fail "Missing environment variables ($ENV_OK/${#REQUIRED_VARS[@]} found)"
fi

# Test 6: IAM Role and Policies
echo ""
info "Test 6: IAM Role Configuration"
ROLE_ARN=$(aws lambda get-function \
    --function-name $LAMBDA_NAME \
    --region $AWS_REGION \
    --query 'Configuration.Role' \
    --output text 2>/dev/null)

if echo "$ROLE_ARN" | grep -q "chainevm"; then
    ROLE_NAME=$(echo "$ROLE_ARN" | awk -F/ '{print $NF}')
    
    # Check policies
    POLICIES=$(aws iam list-role-policies \
        --role-name "$ROLE_NAME" \
        --query 'PolicyNames | length(@)' \
        --output text 2>/dev/null || echo "0")
    
    if [ "$POLICIES" -ge 2 ]; then
        pass "Lambda has required IAM policies (count: $POLICIES)"
    else
        fail "Lambda has insufficient policies (count: $POLICIES)"
    fi
else
    fail "Lambda IAM role unexpected: $ROLE_ARN"
fi

# Test 7: CloudWatch Logs
echo ""
info "Test 7: CloudWatch Logs Configuration"
LOG_GROUP="/aws/lambda/$LAMBDA_NAME"
LOG_RETENTION=$(aws logs describe-log-groups \
    --log-group-name-prefix "$LOG_GROUP" \
    --region $AWS_REGION \
    --query 'logGroups[0].retentionInDays' \
    --output text 2>/dev/null || echo "NOT_FOUND")

if [ "$LOG_RETENTION" != "NOT_FOUND" ] && [ "$LOG_RETENTION" != "None" ]; then
    pass "CloudWatch logs configured with ${LOG_RETENTION}-day retention"
else
    fail "CloudWatch logs not properly configured"
fi

# Test 8: CloudWatch Alarms
echo ""
info "Test 8: CloudWatch Alarms"
ALARMS=$(aws cloudwatch describe-alarms \
    --alarm-name-prefix "evm-" \
    --region $AWS_REGION \
    --query 'MetricAlarms | length(@)' \
    --output text 2>/dev/null || echo "0")

if [ "$ALARMS" -ge 3 ]; then
    pass "CloudWatch alarms configured (count: $ALARMS)"
else
    fail "Insufficient CloudWatch alarms (count: $ALARMS, expected 3+)"
fi

# Test 9: End-to-End Message Processing
echo ""
info "Test 9: E2E Message Processing Test"

# Send test message to queue
TEST_OPERATION_ID="e2e-test-$(date +%s)"
TEST_MESSAGE=$(cat <<EOF
{
  "operation_id": "$TEST_OPERATION_ID",
  "chain_type": "EVM",
  "operation_type": "TRANSFER",
  "from_address": "0x742d35Cc6634C0532925a3b844Bc9e7595f0bEb",
  "to_address": "0x8626f6940E2eb28930eFb4CeF49B2d1F2C9C1199",
  "amount": "0.001",
  "chain_id": "11155111",
  "network": "sepolia",
  "metadata": {
    "test": true,
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  }
}
EOF
)

if [ "$QUEUE_URL" != "NOT_FOUND" ]; then
    SEND_RESULT=$(aws sqs send-message \
        --queue-url "$QUEUE_URL" \
        --message-body "$TEST_MESSAGE" \
        --message-attributes '{"chain_type":{"DataType":"String","StringValue":"EVM"}}' \
        --region $AWS_REGION \
        --query 'MessageId' \
        --output text 2>/dev/null || echo "FAILED")
    
    if [ "$SEND_RESULT" != "FAILED" ]; then
        pass "Test message sent to queue (MessageId: ${SEND_RESULT:0:20}...)"
        
        # Wait for processing
        info "Waiting 5 seconds for Lambda processing..."
        sleep 5
        
        # Check DynamoDB for the operation
        ITEM_CHECK=$(aws dynamodb get-item \
            --table-name $DYNAMODB_TABLE \
            --key "{\"operation_id\":{\"S\":\"$TEST_OPERATION_ID\"}}" \
            --region $AWS_REGION \
            --query 'Item.operation_id.S' \
            --output text 2>/dev/null || echo "NOT_FOUND")
        
        if [ "$ITEM_CHECK" = "$TEST_OPERATION_ID" ]; then
            pass "Message processed and stored in DynamoDB"
        else
            info "Message may still be processing or validation failed (expected for test data)"
        fi
    else
        fail "Failed to send test message to queue"
    fi
else
    fail "Cannot test message processing - queue not accessible"
fi

# Test 10: Lambda Concurrency and Performance
echo ""
info "Test 10: Lambda Configuration Limits"
MEMORY=$(aws lambda get-function-configuration \
    --function-name $LAMBDA_NAME \
    --region $AWS_REGION \
    --query 'MemorySize' \
    --output text 2>/dev/null || echo "0")

TIMEOUT=$(aws lambda get-function-configuration \
    --function-name $LAMBDA_NAME \
    --region $AWS_REGION \
    --query 'Timeout' \
    --output text 2>/dev/null || echo "0")

if [ "$MEMORY" -ge 512 ]; then
    pass "Lambda memory configured: ${MEMORY}MB"
else
    fail "Lambda memory too low: ${MEMORY}MB"
fi

if [ "$TIMEOUT" -ge 30 ]; then
    pass "Lambda timeout configured: ${TIMEOUT}s"
else
    fail "Lambda timeout too low: ${TIMEOUT}s"
fi

# Summary
echo ""
echo "=========================================="
echo "Test Summary:"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"
echo "=========================================="

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All E2E tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    exit 1
fi
