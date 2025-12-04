# ChainEVM - Terraform Infrastructure

Código Terraform para provisionar toda a infraestrutura do ChainEVM na AWS.

## Recursos criados

- ✅ AWS Lambda (ChainEVM executor)
- ✅ AWS SQS (queue + DLQ)
- ✅ AWS DynamoDB (transactions table)
- ✅ IAM Roles e Policies
- ✅ CloudWatch Log Groups
- ✅ CloudWatch Alarms
- ✅ KMS Keys (encryption)

## Pré-requisitos

- Terraform >= 1.0
- AWS CLI v2 configurado
- Credenciais AWS com permissões necessárias

## Deploy

```bash
# 1. Inicialize
terraform init

# 2. Verifique
terraform plan

# 3. Aplique
terraform apply
```

## Configuração

Edite `terraform.tfvars` (ou use variáveis de ambiente):

```hcl
aws_region              = "us-east-1"
environment             = "production"
lambda_memory_size      = 512
lambda_timeout          = 60
rpc_url_ethereum        = "https://eth-mainnet.g.alchemy.com/v2/YOUR_KEY"
rpc_url_polygon         = "https://polygon-mainnet.g.alchemy.com/v2/YOUR_KEY"
# ... outras chains
```

## Arquivos

- `main.tf` - Provider e configuração básica
- `variables.tf` - Variáveis de entrada
- `lambda.tf` - Lambda function e IAM
- `sqs.tf` - SQS queue e DLQ
- `dynamodb.tf` - DynamoDB table
- `outputs.tf` - Saídas do deploy

## Backend State

Para usar S3 backend (recomendado):

```hcl
# terraform/backends.tf
terraform {
  backend "s3" {
    bucket         = "seu-bucket"
    key            = "chainevm/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "terraform-locks"
  }
}
```

## Limpeza

```bash
terraform destroy
```

