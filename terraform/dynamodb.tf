# DynamoDB Table for storing transactions
resource "aws_dynamodb_table" "transactions" {
  name         = var.dynamodb_table_name
  billing_mode = "PAY_PER_REQUEST" # On-demand billing
  hash_key     = "operation_id"

  attribute {
    name = "operation_id"
    type = "S"
  }

  # Global Secondary Index for querying by idempotency_key
  global_secondary_index {
    name            = "idempotency_key-index"
    hash_key        = "idempotency_key"
    projection_type = "ALL"
  }

  # Global Secondary Index for querying by status
  global_secondary_index {
    name            = "status-created_at-index"
    hash_key        = "status"
    range_key       = "created_at"
    projection_type = "KEYS_ONLY"
  }

  attribute {
    name = "idempotency_key"
    type = "S"
  }

  attribute {
    name = "status"
    type = "S"
  }

  attribute {
    name = "created_at"
    type = "S"
  }

  # TTL for automatic cleanup (90 days)
  ttl {
    attribute_name = var.dynamodb_ttl_attribute
    enabled        = true
  }

  # Point-in-time recovery for disaster recovery
  point_in_time_recovery {
    enabled = true
  }

  tags = {
    Description = "EVM transactions storage"
  }
}

# CloudWatch Alarm for item count
resource "aws_cloudwatch_metric_alarm" "dynamodb_item_count" {
  alarm_name          = "${var.dynamodb_table_name}-item-count-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ItemCount"
  namespace           = "AWS/DynamoDB"
  period              = 300
  statistic           = "Average"
  threshold           = 1000000 # 1 million items
  alarm_description   = "Alert when table has too many items"

  dimensions = {
    TableName = aws_dynamodb_table.transactions.name
  }
}
