output "lambda_function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.evm_executor.arn
}

output "lambda_function_name" {
  description = "Name of the Lambda function"
  value       = aws_lambda_function.evm_executor.function_name
}

output "lambda_function_version" {
  description = "Version of the Lambda function"
  value       = aws_lambda_function.evm_executor.version
}

output "lambda_role_arn" {
  description = "ARN of the Lambda IAM role"
  value       = aws_iam_role.lambda_role.arn
}

output "sqs_queue_url" {
  description = "URL of the SQS queue"
  value       = local.evm_queue_url
}

output "sqs_queue_arn" {
  description = "ARN of the SQS queue"
  value       = local.evm_queue_arn
}

output "sqs_dlq_url" {
  description = "URL of the SQS Dead Letter Queue"
  value       = local.evm_dlq_url
}

output "sqs_dlq_arn" {
  description = "ARN of the SQS Dead Letter Queue"
  value       = local.evm_dlq_arn
}

output "dynamodb_table_name" {
  description = "Name of the DynamoDB table"
  value       = aws_dynamodb_table.transactions.name
}

output "dynamodb_table_arn" {
  description = "ARN of the DynamoDB table"
  value       = aws_dynamodb_table.transactions.arn
}

output "dynamodb_table_stream_arn" {
  description = "ARN of the DynamoDB stream"
  value       = aws_dynamodb_table.transactions.stream_arn
}

output "cloudwatch_log_group" {
  description = "CloudWatch log group for Lambda"
  value       = aws_cloudwatch_log_group.lambda_logs.name
}

output "event_source_mapping_uuid" {
  description = "UUID of the Lambda event source mapping"
  value       = aws_lambda_event_source_mapping.sqs_trigger.uuid
}

output "deployment_summary" {
  description = "Summary of deployed resources"
  value = {
    lambda_function = aws_lambda_function.evm_executor.function_name
    sqs_queue       = var.sqs_queue_name
    dynamodb_table  = aws_dynamodb_table.transactions.name
    region          = var.aws_region
  }
}
