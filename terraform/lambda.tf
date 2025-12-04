# IAM Role for Lambda
resource "aws_iam_role" "lambda_role" {
  name = "${var.lambda_function_name}-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

# IAM Policy for Lambda - basic execution
resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  role       = aws_iam_role.lambda_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# IAM Policy for Lambda - SQS access
resource "aws_iam_role_policy" "lambda_sqs_policy" {
  name = "${var.lambda_function_name}-sqs-policy"
  role = aws_iam_role.lambda_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:ChangeMessageVisibility"
        ]
        Resource = data.aws_sqs_queue.evm_queue.arn
      }
    ]
  })
}

# IAM Policy for Lambda - DynamoDB access
resource "aws_iam_role_policy" "lambda_dynamodb_policy" {
  name = "${var.lambda_function_name}-dynamodb-policy"
  role = aws_iam_role.lambda_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:PutItem",
          "dynamodb:GetItem",
          "dynamodb:UpdateItem",
          "dynamodb:Query"
        ]
        Resource = aws_dynamodb_table.transactions.arn
      }
    ]
  })
}

# Lambda Function
resource "aws_lambda_function" "evm_executor" {
  filename         = var.lambda_file_path
  function_name    = var.lambda_function_name
  role             = aws_iam_role.lambda_role.arn
  handler          = "bootstrap"
  runtime          = var.lambda_runtime
  timeout          = var.lambda_timeout
  memory_size      = var.lambda_memory_size
  source_code_hash = fileexists(var.lambda_file_path) ? filebase64sha256(var.lambda_file_path) : ""

  environment {
    variables = {
      ENVIRONMENT             = var.environment
      DYNAMODB_TABLE_NAME     = aws_dynamodb_table.transactions.name
      SQS_QUEUE_URL           = data.aws_sqs_queue.evm_queue.url
      RPC_URL_ETHEREUM        = var.rpc_url_ethereum
      RPC_URL_POLYGON         = var.rpc_url_polygon
      RPC_URL_BSC             = var.rpc_url_bsc
      RPC_URL_ARBITRUM        = var.rpc_url_arbitrum
      RPC_URL_OPTIMISM        = var.rpc_url_optimism
      RPC_URL_AVALANCHE       = var.rpc_url_avalanche
      RPC_TIMEOUT_SECONDS     = var.rpc_timeout_seconds
      REQUEST_TIMEOUT_SECONDS = 30
      REQUIRED_CONFIRMATIONS  = var.required_confirmations
    }
  }

  depends_on = [
    aws_iam_role_policy.lambda_sqs_policy,
    aws_iam_role_policy.lambda_dynamodb_policy,
    aws_iam_role_policy_attachment.lambda_basic_execution
  ]
}

# Lambda Event Source Mapping (SQS trigger)
resource "aws_lambda_event_source_mapping" "sqs_trigger" {
  event_source_arn                   = data.aws_sqs_queue.evm_queue.arn
  function_name                      = aws_lambda_function.evm_executor.arn
  batch_size                         = 1
  maximum_batching_window_in_seconds = 5

  # On error, send to DLQ
  function_response_types = ["ReportBatchItemFailures"]
}

# CloudWatch Log Group for Lambda
resource "aws_cloudwatch_log_group" "lambda_logs" {
  name              = "/aws/lambda/${var.lambda_function_name}"
  retention_in_days = 14

  depends_on = [aws_lambda_function.evm_executor]
}

# Lambda Alias for easy versioning
resource "aws_lambda_alias" "prod" {
  name             = "prod"
  description      = "Production alias"
  function_name    = aws_lambda_function.evm_executor.function_name
  function_version = aws_lambda_function.evm_executor.version
}
