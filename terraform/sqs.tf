# Data source para a fila SQS existente (criada pelo chainorchestrator)
data "aws_sqs_queue" "evm_queue" {
  name = var.sqs_queue_name
}

# Data source para a DLQ existente (criada pelo chainorchestrator)
data "aws_sqs_queue" "evm_dlq" {
  name = var.sqs_dlq_name
}

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
    QueueName = data.aws_sqs_queue.evm_queue.name
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
    QueueName = data.aws_sqs_queue.evm_dlq.name
  }
}
