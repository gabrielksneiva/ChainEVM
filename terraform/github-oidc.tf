# GitHub OIDC Provider and IAM Roles for CI/CD
# This allows GitHub Actions to assume AWS IAM roles without storing credentials
#
# ⚠️  IMPORTANT CIRCULAR DEPENDENCY NOTE:
# GitHub Actions uses these roles to run Terraform. If Terraform tries to CREATE
# these roles, it creates a circular dependency: GitHub Actions can't assume the 
# role to run Terraform, because the role doesn't exist yet.
#
# SOLUTION: These roles must be created MANUALLY first (see github-oidc-bootstrap.tf.example)
# After creation, Terraform will manage their policies and permissions.

terraform {
  required_version = ">= 1.0"
}

# Data sources para obter informações da conta e região
data "aws_caller_identity" "current" {}
data "aws_region" "current" {}

locals {
  github_repo = "gabrielksneiva/ChainEVM"
  github_oidc_provider_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:oidc-provider/token.actions.githubusercontent.com"
  terraform_state_bucket  = "chainevm-terraform-state"
  terraform_state_key     = "evm/terraform.tfstate"
}

# Reference existing GitHub Actions roles (created manually during bootstrap)
data "aws_iam_role" "github_actions_terraform" {
  name = "github-actions-terraform-role"
}

data "aws_iam_role" "github_actions_deploy" {
  name = "github-actions-deploy-role"
}

data "aws_iam_role" "github_actions_e2e" {
  name = "github-actions-e2e-role"
}

# Policy for Terraform - S3 backend access and infrastructure management
resource "aws_iam_role_policy" "github_actions_terraform_policy" {
  name = "github-actions-terraform-policy"
  role = data.aws_iam_role.github_actions_terraform.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "S3BackendAccess"
        Effect = "Allow"
        Action = [
          "s3:ListBucket",
          "s3:GetBucketVersioning",
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject"
        ]
        Resource = [
          "arn:aws:s3:::${local.terraform_state_bucket}",
          "arn:aws:s3:::${local.terraform_state_bucket}/*"
        ]
      },
      {
        Sid    = "DynamoDBStatelock"
        Effect = "Allow"
        Action = [
          "dynamodb:DescribeTable",
          "dynamodb:GetItem",
          "dynamodb:PutItem",
          "dynamodb:DeleteItem"
        ]
        Resource = "arn:aws:dynamodb:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:table/terraform-state-lock"
      },
      {
        Sid    = "DynamoDBManagement"
        Effect = "Allow"
        Action = [
          "dynamodb:CreateTable",
          "dynamodb:DeleteTable",
          "dynamodb:DescribeTable",
          "dynamodb:ListTables",
          "dynamodb:UpdateTable",
          "dynamodb:TagResource",
          "dynamodb:UntagResource"
        ]
        Resource = "arn:aws:dynamodb:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:table/*"
      },
      {
        Sid    = "LambdaManagement"
        Effect = "Allow"
        Action = [
          "lambda:CreateFunction",
          "lambda:DeleteFunction",
          "lambda:GetFunction",
          "lambda:ListFunctions",
          "lambda:UpdateFunctionCode",
          "lambda:UpdateFunctionConfiguration",
          "lambda:TagResource",
          "lambda:UntagResource",
          "lambda:CreateAlias",
          "lambda:DeleteAlias",
          "lambda:UpdateAlias",
          "lambda:GetAlias"
        ]
        Resource = "arn:aws:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:*"
      },
      {
        Sid    = "SQSManagement"
        Effect = "Allow"
        Action = [
          "sqs:CreateQueue",
          "sqs:DeleteQueue",
          "sqs:GetQueueAttributes",
          "sqs:ListQueues",
          "sqs:SetQueueAttributes",
          "sqs:TagQueue",
          "sqs:UntagQueue"
        ]
        Resource = "arn:aws:sqs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
      },
      {
        Sid    = "IAMManagement"
        Effect = "Allow"
        Action = [
          "iam:CreateRole",
          "iam:DeleteRole",
          "iam:GetRole",
          "iam:GetRolePolicy",
          "iam:PassRole",
          "iam:AttachRolePolicy",
          "iam:DetachRolePolicy",
          "iam:PutRolePolicy",
          "iam:DeleteRolePolicy",
          "iam:ListRolePolicies",
          "iam:TagRole",
          "iam:UntagRole"
        ]
        Resource = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/*"
      },
      {
        Sid    = "CloudWatchLogsManagement"
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:DeleteLogGroup",
          "logs:DescribeLogGroups",
          "logs:TagLogGroup",
          "logs:UntagLogGroup"
        ]
        Resource = "arn:aws:logs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:log-group:*"
      },
      {
        Sid    = "CloudWatchMetrics"
        Effect = "Allow"
        Action = [
          "cloudwatch:PutMetricAlarm",
          "cloudwatch:DeleteAlarms",
          "cloudwatch:DescribeAlarms"
        ]
        Resource = "arn:aws:cloudwatch:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:alarm:*"
      }
    ]
  })
}

# Policy for Lambda Deployment
resource "aws_iam_role_policy" "github_actions_deploy_policy" {
  name = "github-actions-deploy-policy"
  role = data.aws_iam_role.github_actions_deploy.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "lambda:UpdateFunctionCode",
          "lambda:GetFunction"
        ]
        Resource = [
          "arn:aws:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:chainevm",
          "arn:aws:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:chainevm-prod"
        ]
      }
    ]
  })
}

# Policy for E2E Tests
resource "aws_iam_role_policy" "github_actions_e2e_policy" {
  name = "github-actions-e2e-policy"
  role = data.aws_iam_role.github_actions_e2e.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "DynamoDBRead"
        Effect = "Allow"
        Action = [
          "dynamodb:GetItem",
          "dynamodb:Query",
          "dynamodb:Scan",
          "dynamodb:DescribeTable"
        ]
        Resource = "arn:aws:dynamodb:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:table/*"
      },
      {
        Sid    = "SQSRead"
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:GetQueueAttributes",
          "sqs:GetQueueUrl"
        ]
        Resource = "arn:aws:sqs:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:*"
      },
      {
        Sid    = "LambdaInvoke"
        Effect = "Allow"
        Action = [
          "lambda:InvokeFunction"
        ]
        Resource = [
          "arn:aws:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:chainevm",
          "arn:aws:lambda:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:function:chainevm-prod"
        ]
      }
    ]
  })
}

# Outputs for reference
output "github_actions_terraform_role_arn" {
  description = "ARN of GitHub Actions Terraform role"
  value       = data.aws_iam_role.github_actions_terraform.arn
}

output "github_actions_deploy_role_arn" {
  description = "ARN of GitHub Actions Deploy role"
  value       = data.aws_iam_role.github_actions_deploy.arn
}

output "github_actions_e2e_role_arn" {
  description = "ARN of GitHub Actions E2E role"
  value       = data.aws_iam_role.github_actions_e2e.arn
}
