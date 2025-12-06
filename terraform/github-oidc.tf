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
}

# TODO: Import existing GitHub Actions roles after manual bootstrap
# terraform import aws_iam_role.github_actions_terraform github-actions-terraform-role
# terraform import aws_iam_role.github_actions_deploy github-actions-deploy-role
# terraform import aws_iam_role.github_actions_e2e github-actions-e2e-role
