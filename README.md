# ChainEVM — AWS Lambda para Execução de Operações EVM

O **ChainEVM** é um AWS Lambda especializado responsável por **executar todas as operações relacionadas às blockchains compatíveis com EVM** (Ethereum, Polygon, BNB Chain, Arbitrum, Optimism, Avalanche).

Ele funciona como o executor especializado acionado pelo ecossistema de orquestração (**ChainOrchestrator**).

---

## �� Responsabilidades

- ✅ **Processar mensagens** recebidas da fila SQS: `chainorchestrator-evm-queue-production`
- ✅ **Interpretar tipos de operação** EVM solicitadas
- ✅ **Executar chamadas on-chain** (RPC) para qualquer chain EVM suportada
- ✅ **Assinar e enviar transações** quando solicitado
- ✅ **Padronizar e validar** todos os retornos das RPCs
- ✅ **Registrar logs estruturados** para auditoria e rastreabilidade
- ✅ **Persistir transações e dados EVM** em DynamoDB
- ✅ **Garantir execução idempotente** via idempotency key
- ✅ **Executar operações resilientes** com retry automático via SQS

---

## 🏗️ Arquitetura

```
SQS Queue (chainorchestrator-evm-queue-production)
    ↓
Lambda Handler (SQS Trigger)
    ↓
[Domain Layer] → Validação de tipos, Value Objects
    ↓
[Application Layer] → Use Cases, DTOs, Business Logic
    ↓
[Infrastructure Layer] → RPC Clients, DynamoDB, SQS, Logger
    ↓
DynamoDB Table (evm-transactions)
```

### Padrões Arquitetônicos

- **Clean Architecture** - Separação clara entre camadas
- **Domain-Driven Design (DDD)** - Lógica de negócio centralizada
- **Event-Driven** - Orientado a eventos via SQS
- **Dependency Injection** - Via construtores (sem framework para Lambda)
- **Logs Estruturados** - Zap Logger com contexto estruturado

---

## 📦 Estrutura do Projeto

```
ChainEVM/
├── cmd/
│   └── lambda/
│       └── main.go                    # Handler Lambda
├── internal/
│   ├── application/
│   │   ├── dtos/                     # Data Transfer Objects
│   │   └── usecases/                 # Casos de uso
│   ├── domain/
│   │   ├── entities/                 # Entidades de domínio
│   │   └── valueobjects/             # Value Objects
│   ├── infrastructure/
│   │   ├── eventbus/                 # SQS Consumer
│   │   ├── rpc/                      # EVM RPC Clients
│   │   ├── database/                 # DynamoDB Repository
│   │   └── logger/                   # Logger (Zap)
│   └── interfaces/
│       └── handlers/                 # (Reservado para expansão)
├── pkg/
│   ├── config/                       # Configurações
│   └── errors/                       # Erros customizados
├── terraform/                        # Infraestrutura como Código
├── docs/                             # Documentação
├── go.mod                            # Dependências Go
├── Makefile                          # Automação
└── README.md                         # Este arquivo
```

---

## 🔗 Fluxo Simplificado

1. **ChainOrchestrator** envia instrução → SQS `chainorchestrator-evm-queue-production`
2. **ChainEVM Lambda** é acionado automaticamente (SQS Trigger)
3. Lambda **processa mensagem**:
   - Valida entrada (chain type, operation type, endereços)
   - Verifica **idempotência** (já processado?)
   - Executa operação (read ou write)
   - Salva resultado em **DynamoDB**
4. **Resposta** é retornada ao pipeline de orquestração
5. **Ack** de mensagem SQS (delete) → conclusão

---

## 🛠️ Stack Tecnológica

- **Linguagem:** Go 1.24+
- **Runtime:** AWS Lambda (Go Runtime)
- **Queue:** AWS SQS
- **Persistência:** AWS DynamoDB
- **RPC Library:** go-ethereum
- **Logging:** Zap (estruturado)
- **IaC:** Terraform

---

## 🚀 Deployment

### Pré-requisitos

- Go 1.24+
- AWS CLI v2
- Terraform
- Credenciais AWS com permissões para Lambda, SQS, DynamoDB

### Build

```bash
make build
```

Isso gera `lambda-deployment.zip` pronto para deploy.

### Deploy via Terraform

```bash
cd terraform/
terraform init
terraform plan
terraform apply
```

### Deploy Manual

```bash
aws lambda update-function-code \
  --function-name ChainEVM \
  --zip-file fileb://lambda-deployment.zip \
  --region us-east-1
```

---

## 📝 Tipos de Operação Suportadas

### Write Operations (modificam estado)
- `TRANSFER` - Transferência de ETH/tokens
- `DEPLOY` - Deployment de contrato
- `CALL` - Chamada de função em contrato
- `APPROVE` - Aprovação de gastos (ERC-20)
- `SWAP` - Troca em DEX
- `STAKE` - Staking
- `UNSTAKE` - Unstaking
- `WITHDRAW` - Saque/withdraw
- `MINT` - Mint de tokens
- `BURN` - Burn de tokens

### Read Operations (apenas leitura)
- `QUERY` - Query customizada
- `GET_BALANCE` - Saldo de endereço
- `GET_NONCE` - Nonce de endereço
- `ESTIMATE_GAS` - Estimativa de gas

---

## 🌐 Blockchains Suportadas

- ✅ Ethereum (Mainnet)
- ✅ Polygon
- ✅ BNB Smart Chain
- ✅ Arbitrum
- ✅ Optimism
- ✅ Avalanche

---

## 📡 Mensagem SQS (Input)

```json
{
  "operation_id": "123e4567-e89b-12d3-a456-426614174000",
  "chain_type": "POLYGON",
  "operation_type": "TRANSFER",
  "from_address": "0x1234567890123456789012345678901234567890",
  "to_address": "0x0987654321098765432109876543210987654321",
  "payload": {
    "amount": "1000000000000000000",
    "data": "0x"
  },
  "idempotency_key": "550e8400-e29b-41d4-a716-446655440001"
}
```

---

## 📤 Resposta da Operação (Output)

```json
{
  "operation_id": "123e4567-e89b-12d3-a456-426614174000",
  "chain_type": "POLYGON",
  "transaction_hash": "0xabc123def456...",
  "status": "SUCCESS",
  "block_number": 45678901,
  "gas_used": 21000,
  "gas_price": "50000000000",
  "error_message": "",
  "created_at": "2024-12-04T10:30:00Z",
  "executed_at": "2024-12-04T10:31:15Z"
}
```

---

## 🔐 Segurança & Boas Práticas

- ✅ **Idempotência garantida** via idempotency key
- ✅ **Validação rigorosa** de entrada em todos os níveis
- ✅ **Logs estruturados** para auditoria
- ✅ **Timeouts** configuráveis para RPC calls
- ✅ **Retry automático** via SQS visibility timeout
- ✅ **Encriptação** de dados em repouso (DynamoDB)
- ✅ **IAM roles** com princípio de menor privilégio

---

## 📊 Variáveis de Ambiente

```bash
# AWS
AWS_REGION=us-east-1

# SQS
SQS_QUEUE_URL=https://sqs.us-east-1.amazonaws.com/123456789/chainorchestrator-evm-queue-production

# DynamoDB
DYNAMODB_TABLE_NAME=evm-transactions

# RPC URLs (por chain)
RPC_URL_ETHEREUM=https://eth-mainnet.g.alchemy.com/v2/YOUR_KEY
RPC_URL_POLYGON=https://polygon-mainnet.g.alchemy.com/v2/YOUR_KEY
RPC_URL_BSC=https://bsc-mainnet.infura.io/v3/YOUR_KEY
RPC_URL_ARBITRUM=https://arb-mainnet.g.alchemy.com/v2/YOUR_KEY
RPC_URL_OPTIMISM=https://opt-mainnet.g.alchemy.com/v2/YOUR_KEY
RPC_URL_AVALANCHE=https://avax-mainnet.g.alchemy.com/v2/YOUR_KEY

# Timeouts
REQUEST_TIMEOUT_SECONDS=30
RPC_TIMEOUT_SECONDS=10

# Confirmações
REQUIRED_CONFIRMATIONS=12

# Ambiente
ENVIRONMENT=production
```

---

## 🧪 Testes

```bash
make test
```

---

## 📚 Documentação

- [SETUP.md](docs/SETUP.md) - Guia de configuração local
- [DEPLOY.md](docs/DEPLOY.md) - Guia de deployment

---

## 🤝 Integração com ChainOrchestrator

O ChainEVM é acionado automaticamente quando o ChainOrchestrator publica uma mensagem na fila SQS com o seguinte padrão:

```
1. Orchestrator valida operação
2. Orchestrator publica em SQS
3. ChainEVM Lambda é acionado
4. ChainEVM executa operação on-chain
5. ChainEVM persiste resultado em DynamoDB
```

---

## 📝 License

Privado - Gabriel K. Sneiva

