# Locals para construir ARNs e URLs das filas SQS (criadas pelo chainorchestrator)
locals {
  aws_account_id = data.aws_caller_identity.current.account_id
  aws_region     = data.aws_region.current.name

  evm_queue_arn = "arn:aws:sqs:${local.aws_region}:${local.aws_account_id}:${var.sqs_queue_name}"
  evm_queue_url = "https://sqs.${local.aws_region}.amazonaws.com/${local.aws_account_id}/${var.sqs_queue_name}"
  evm_dlq_arn   = "arn:aws:sqs:${local.aws_region}:${local.aws_account_id}:${var.sqs_dlq_name}"
  evm_dlq_url   = "https://sqs.${local.aws_region}.amazonaws.com/${local.aws_account_id}/${var.sqs_dlq_name}"
}

# Data sources para obter informações da conta e região
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

# CloudWatch Alarm for queue depth
resource "aws_cloudwatch_metric_alarm" "sqs_queue_depth" {
  alarm_name          = "${var.sqs_queue_name}-depth-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = 300
  statistic           = "Average"
  threshold           = 100
  alarm_description   = "Alert when SQS queue has too many messages"

  dimensions = {
    QueueName = var.sqs_queue_name
  }
}

# CloudWatch Alarm for DLQ
resource "aws_cloudwatch_metric_alarm" "sqs_dlq_messages" {
  alarm_name          = "${var.sqs_queue_name}-dlq-messages"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 1
  metric_name         = "ApproximateNumberOfMessagesVisible"
  namespace           = "AWS/SQS"
  period              = 300
  statistic           = "Average"
  threshold           = 1
  alarm_description   = "Alert when messages appear in DLQ"

  dimensions = {
    QueueName = var.sqs_dlq_name
  }
}
