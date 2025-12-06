# AWS Region Configuration  
# Updated: 2025-12-04 23:15 - Testing E2E with CloudWatch DescribeAlarms fix
variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "lambda_function_name" {
  description = "Lambda function name"
  type        = string
  default     = "chainevm"
}

variable "lambda_runtime" {
  description = "Lambda runtime"
  type        = string
  default     = "provided.al2"
}

variable "lambda_timeout" {
  description = "Lambda timeout in seconds"
  type        = number
  default     = 60
}

variable "lambda_memory_size" {
  description = "Lambda memory size in MB"
  type        = number
  default     = 512
}

variable "lambda_file_path" {
  description = "Path to Lambda deployment zip"
  type        = string
  default     = "../lambda-deployment.zip"
}

variable "sqs_queue_name" {
  description = "SQS queue name"
  type        = string
  default     = "evm-queue"
}

variable "sqs_dlq_name" {
  description = "SQS dead letter queue name"
  type        = string
  default     = "evm-dlq"
}

variable "sqs_visibility_timeout" {
  description = "SQS visibility timeout in seconds"
  type        = number
  default     = 300
}

variable "sqs_retention_period" {
  description = "SQS retention period in seconds (14 days)"
  type        = number
  default     = 1209600
}

variable "dynamodb_table_name" {
  description = "DynamoDB table name"
  type        = string
  default     = "evm-transactions"
}

variable "dynamodb_ttl_attribute" {
  description = "DynamoDB TTL attribute name"
  type        = string
  default     = "ttl"
}

variable "rpc_url_ethereum" {
  description = "Ethereum RPC URL"
  type        = string
  sensitive   = true
  default     = ""
}

variable "rpc_url_polygon" {
  description = "Polygon RPC URL"
  type        = string
  sensitive   = true
  default     = ""
}

variable "rpc_url_bsc" {
  description = "BSC RPC URL"
  type        = string
  sensitive   = true
  default     = ""
}

variable "rpc_url_arbitrum" {
  description = "Arbitrum RPC URL"
  type        = string
  sensitive   = true
  default     = ""
}

variable "rpc_url_optimism" {
  description = "Optimism RPC URL"
  type        = string
  sensitive   = true
  default     = ""
}

variable "rpc_url_avalanche" {
  description = "Avalanche RPC URL"
  type        = string
  sensitive   = true
  default     = ""
}

variable "rpc_timeout_seconds" {
  description = "RPC timeout in seconds"
  type        = number
  default     = 10
}

variable "required_confirmations" {
  description = "Required confirmations for transactions"
  type        = number
  default     = 12
}
