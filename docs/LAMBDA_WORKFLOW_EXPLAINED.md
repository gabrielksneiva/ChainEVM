# 🚀 ChainEVM Lambda Workflow - Guia Completo

## Introdução: Entender o Fluxo de Uma Transação Blockchain

Imagine que você está em um **banco**. Alguém vem até você e diz: "Eu quero transferir dinheiro para outra conta." O que o banco faz?

1. 📋 **Abre um formulário** (recebe os dados da transação)
2. ✅ **Valida os dados** (verifica se você tem saldo)
3. 🔐 **Processa a transação** (comunica com o blockchain)
4. 💾 **Registra tudo** (guarda em um banco de dados)
5. 📧 **Confirma ao cliente** (retorna o status)

O ChainEVM Lambda funciona **exatamente assim**, mas para transações blockchain em múltiplas redes (Ethereum, Polygon, BSC, etc).

---

## 📊 Arquitetura Visual do Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│                      API Gateway / SNS                          │
│                    (Sistema externo)                            │
└────────────────────────────┬──────────────────────────────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │   ChainOrch.    │
                    │    Lambda       │ ◄── Processa requisições da API
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │   SNS Topic     │ ◄── Publica mensagens de eventos
                    └────────┬────────┘
                             │
                             ▼
                    ┌─────────────────┐
                    │   SQS Queue     │ ◄── Fila de espera (produtor/consumidor)
                    │  (Messages)     │
                    └────────┬────────┘
                             │
                    ┌────────▼──────────┐
                    │  ChainEVM Lambda  │ ◄── ⭐ VOCÊ ESTÁ AQUI
                    │  (Consumidor)     │
                    └────────┬──────────┘
                             │
        ┌────────────┬────────┼────────┬──────────┐
        ▼            ▼        ▼        ▼          ▼
    ┌─────┐   ┌─────────┐ ┌────────┐ ┌──────┐ ┌──────────┐
    │ ETH │   │ Polygon │ │  BSC   │ │Arbit.│ │Optimism  │
    │Sepo.│   │  Amoy   │ │Testnet │ │Sep.  │ │ Testnet  │
    └─────┘   └─────────┘ └────────┘ └──────┘ └──────────┘
    
        (Testnets EVM - Ethereum Sepolia, etc)
        
        ▼            ▼        ▼        ▼          ▼
    ┌──────────────────────────────────────────────────┐
    │        DynamoDB: chainevm-transactions-dev       │
    │     (Registro de todas as transações)            │
    └──────────────────────────────────────────────────┘
```

---

## 🔄 O Workflow Passo a Passo

### **FASE 1: INICIALIZAÇÃO (Quando o Lambda Sobe)**

```
┌─────────────────────────────────────────┐
│      AWS Lambda é Acionado               │
│    (Ambiente é preparado)                │
└─────────────────┬───────────────────────┘
                  │
     ┌────────────┼────────────┐
     │            │            │
     ▼            ▼            ▼
  ┌─────┐   ┌──────────┐  ┌────────────┐
  │Config│  │Logger    │  │AWS Clients │
  │Load  │  │Setup     │  │Initialize │
  └─────┘   └──────────┘  └────────────┘
     │            │            │
     └────────────┼────────────┘
                  │
                  ▼
     ┌──────────────────────────┐
     │ SQS + DynamoDB + RPC     │
     │ Clients conectados! ✓    │
     └──────────────────────────┘
```

**O que acontece na inicialização (`init()` function):**

1. **Carrega Configuração** (`pkgconfig.LoadConfig()`)
   - Lê variáveis de ambiente (RPC URLs, Table Names, Queue URLs)
   - Analogia: 📚 Buscar o manual de instruções antes de começar

2. **Inicializa Logger** (`logger.NewLogger()`)
   - Cria um sistema de logs
   - Analogia: 📝 Abrir um caderno para anotar tudo que vai acontecer

3. **Conecta aos Serviços AWS**
   ```go
   awsCfg := config.LoadDefaultConfig(ctx)  // Carrega credenciais
   sqsClient := sqs.NewFromConfig(awsCfg)   // Conecta ao SQS
   dynamoDB := dynamodb.NewFromConfig(awsCfg) // Conecta ao DynamoDB
   ```
   - Analogia: 🔌 Conectar cabos de rede, ligar impressora, etc

4. **Inicializa Clientes RPC** (Ethereum, Polygon, BSC, etc)
   - Para cada blockchain configurado, cria uma conexão
   - Analogia: 📞 Discar para múltiplas agências bancárias diferentes

---

### **FASE 2: RECEBIMENTO (Handler Principal)**

Quando uma **mensagem chega na SQS**, o AWS Lambda executa a função `handler`:

```go
func handler(ctx context.Context, event events.SQSEvent) error {
    // event.Records = Lista de mensagens da SQS
    for _, record := range event.Records {
        processMessage(ctx, record)
    }
}
```

**Analogia do Restaurante:**
- SQS é como o **ticket da cozinha** 
- Cada ticket tem um pedido (mensagem)
- O chef (Lambda) pega o ticket e processa

---

### **FASE 3: PROCESSAMENTO (ProcessMessage)**

Aqui é onde a **mágica acontece**:

```go
func processMessage(ctx context.Context, record events.SQSMessage) error {
    // PASSO 1: Parse JSON
    var msgBody eventbus.Message
    json.Unmarshal([]byte(record.Body), &msgBody)
    
    // PASSO 2: Converter para DTO (Data Transfer Object)
    req := &dtos.ExecuteTransactionRequest{
        OperationID:    msgBody.OperationID,
        ChainType:      msgBody.ChainType,      // Ethereum, Polygon, etc
        OperationType:  msgBody.OperationType,  // Transfer, Swap, etc
        FromAddress:    msgBody.FromAddress,    // Quem envia
        ToAddress:      msgBody.ToAddress,      // Quem recebe
        Payload:        msgBody.Payload,        // Dados da transação
        IdempotencyKey: msgBody.IdempotencyKey, // Previne duplicatas
    }
    
    // PASSO 3: Executar Use Case
    response, err := executeUseCase.Execute(ctx, req)
    
    // PASSO 4: Deletar mensagem (se sucesso)
    sqsConsumer.DeleteMessage(ctx, &receiptHandle)
}
```

**Fluxo detalhado:**

```
MENSAGEM SQS (JSON)
    │
    ▼
┌──────────────────────────────────────┐
│ PASSO 1: Parse & Validação          │
│ - Desserializar JSON                │
│ - Validar estrutura                 │
└──────────┬───────────────────────────┘
           │
           ▼
┌──────────────────────────────────────┐
│ PASSO 2: Execute Use Case            │
│ - Validar chain_type                │
│ - Validar endereços                 │
│ - Verificar idempotência            │
└──────────┬───────────────────────────┘
           │
           ▼
┌──────────────────────────────────────┐
│ PASSO 3: Chamar RPC Client           │
│ - Conectar ao blockchain correto    │
│ - Enviar transação                  │
│ - Aguardar confirmações             │
└──────────┬───────────────────────────┘
           │
           ▼
┌──────────────────────────────────────┐
│ PASSO 4: Salvar no DynamoDB         │
│ - Registrar status: PENDING         │
│ - Aguardar confirmações             │
│ - Atualizar para: CONFIRMED         │
└──────────┬───────────────────────────┘
           │
           ▼
┌──────────────────────────────────────┐
│ PASSO 5: Limpar Fila                │
│ - Deletar mensagem do SQS           │
│ - Confirmar sucesso                 │
└──────────────────────────────────────┘
```

---

## 🎯 O Use Case: ExecuteEVMTransactionUseCase

Este é o **coração** da aplicação. Implementa a lógica de negócio:

### **Estrutura:**

```go
type ExecuteEVMTransactionUseCase struct {
    rpcClients      map[string]rpc.RPCClient        // Conexões aos blockchains
    transactionRepo database.TransactionRepository   // Acesso ao DB
    logger          *zap.Logger                      // Logs
}
```

### **Método Execute() - O Fluxo Principal:**

```
1. VALIDAÇÃO
   ├─ Chain Type válida? (Ethereum, Polygon, etc)
   ├─ Endereços válidos? (Formato de endereço EVM)
   ├─ Operação válida? (Transfer, Swap, etc)
   └─ Payload valido?

2. VERIFICAR IDEMPOTÊNCIA
   ├─ Essa transação já foi processada?
   ├─ Se SIM → Retorna resultado anterior (evita duplicatas)
   └─ Se NÃO → Continua

3. SALVAR ESTADO INICIAL
   ├─ Cria Transaction entity no DynamoDB
   ├─ Status: PENDING
   └─ Timestamp: NOW

4. CHAMAR RPC CLIENT
   ├─ Seleciona cliente correto (Ethereum? Polygon?)
   ├─ Envia transação para o blockchain
   ├─ Obtém TX Hash
   └─ Status: BROADCAST

5. AGUARDAR CONFIRMAÇÕES
   ├─ Monitora confirmações na blockchain
   ├─ Verificação com Circuit Breaker (retry com backoff)
   └─ Status: CONFIRMED após N confirmações

6. RETORNAR RESULTADO
   ├─ Transaction Hash
   ├─ Status final
   ├─ Block number
   └─ Confirmations
```

### **Analogia do Banco Eletrônico:**

```
Você quer transferir $100

┌─────────────────────────────────────┐
│ 1. VALIDAÇÃO (5 segundos)           │
│ - Verifica se você é de verdade     │
│ - Valida se a senha está correta    │
│ - Confirma que você tem saldo       │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│ 2. PROTEÇÃO CONTRA DUPLICATAS       │
│ - Verifica se já pediu isso antes   │
│ - Se sim, retorna resultado anterior│
│ - Se não, continua                  │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│ 3. REGISTRAR (Arquivo criado)       │
│ - Abre um arquivo de transação      │
│ - Escreve: Data, Valor, Status      │
│ - Status = PROCESSANDO              │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│ 4. EXECUTAR (Comunica com Banco)    │
│ - Banco processa transação          │
│ - Gera número de confirmação (hash) │
│ - Status = ENVIADO                  │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│ 5. CONFIRMAR (Aguarda verificação)  │
│ - Banco verifica 3x (confirmações)  │
│ - Tudo OK? Status = CONFIRMADO      │
│ - Erro? Status = FALHOU             │
└─────────────────────────────────────┘
              │
              ▼
        ✅ SUCESSO!
```

---

## 🔗 Componentes Chave

### **1. SQS Consumer (Fila de Mensagens)**

**Localização:** `internal/infrastructure/eventbus/sqs_consumer.go`

```go
type SQSConsumer struct {
    adapter   SQSAdapter          // Adapter para SQS
    queueURL  string             // URL da fila
    logger    *zap.Logger        // Logs
}

// DeleteMessage - Remove mensagem da fila após processamento
func (c *SQSConsumer) DeleteMessage(ctx context.Context, receiptHandle *string) error
```

**Analogia:** O SQS Consumer é como um **funcionário que pega tickets da fila e marca como resolvido** quando o chef termina.

---

### **2. DynamoDB Repository (Banco de Dados)**

**Localização:** `internal/infrastructure/database/transaction_repository.go`

```go
type TransactionRepository interface {
    Create(ctx context.Context, tx *entities.EVMTransaction) error
    GetByOperationID(ctx context.Context, id string) (*entities.EVMTransaction, error)
    Update(ctx context.Context, tx *entities.EVMTransaction) error
}
```

**Estrutura de Dados:**

```
Tabela: chainevm-transactions-dev

PrimaryKey: operation_id (Identifica única transação)
Atributos:
├─ operation_id (String) - ID único
├─ chain_type (String) - Ethereum, Polygon, BSC, etc
├─ status (String) - PENDING, CONFIRMED, FAILED
├─ transaction_hash (String) - Hash na blockchain
├─ from_address (String) - Endereço que envia
├─ to_address (String) - Endereço que recebe
├─ confirmations (Number) - Quantas confirmações tem
├─ created_at (String) - Quando foi criado
├─ updated_at (String) - Última atualização
├─ idempotency_key (String) - Previne duplicatas
└─ ttl (Number) - Tempo para expirar (30 dias)

Índices Secundários:
├─ idempotency_key-index (Buscar por chave de idempotência)
└─ status-created_at-index (Filtrar por status e data)
```

**Analogia:** DynamoDB é como o **arquivo/banco de dados do banco** onde todas as transações são registradas e podem ser consultadas depois.

---

### **3. RPC Client (Comunicação com Blockchain)**

**Localização:** `internal/infrastructure/rpc/rpc_client.go`

```go
type RPCClient interface {
    SendTransaction(ctx context.Context, tx *models.Transaction) (string, error)
    GetTransactionReceipt(ctx context.Context, txHash string) (*models.TransactionReceipt, error)
    GetBlockNumber(ctx context.Context) (uint64, error)
}
```

**Fluxo:**

```
RPC Client (Ethereum Sepolia)
    │
    ├─ Valida conexão
    ├─ Envia transação (POST request para RPC endpoint)
    ├─ Recebe TX Hash
    ├─ Monitora confirmações
    └─ Retorna status final

Exemplo de Endpoints:
├─ Ethereum Sepolia: https://ethereum-sepolia-rpc.publicnode.com
├─ Polygon Amoy: https://polygon-amoy-rpc.publicnode.com
├─ BSC Testnet: https://bsc-testnet-rpc.publicnode.com
└─ etc...
```

**Analogia:** RPC Client é como um **telefone direto para a agência do banco**. Você liga, pede para processar a transação, e ele retorna o status.

---

### **4. Circuit Breaker (Proteção contra Falhas)**

**Localização:** `internal/infrastructure/rpc/circuit_breaker.go`

```go
type CircuitBreaker struct {
    maxRetries      int
    retryDelay      time.Duration
    backoffMultiplier float64
}

// Retry com exponential backoff
// Tentativa 1: Aguarda 100ms
// Tentativa 2: Aguarda 200ms
// Tentativa 3: Aguarda 400ms
// ...
```

**Analogia:** Circuit Breaker é como um **disjuntor automático**. Se o blockchain fica fora do ar:

```
Tentativa 1: Falha! ❌ Espera 100ms
Tentativa 2: Falha! ❌ Espera 200ms
Tentativa 3: Falha! ❌ Espera 400ms
Tentativa 4: Sucesso! ✅

Se muito falho:
  └─ Para de tentar e retorna erro
```

---

## 📈 Estados de Uma Transação

```
Estado Machine:

    ┌───────────────┐
    │   PENDING     │  (Criada no BD, não enviada)
    └───────┬───────┘
            │ (Enviada para blockchain)
            ▼
    ┌───────────────┐
    │  BROADCAST    │  (Está na mempool da blockchain)
    └───────┬───────┘
            │ (Minerada em um bloco)
            ▼
    ┌───────────────┐
    │  CONFIRMED    │  (Tem N confirmações - SUCESSO ✅)
    └───────────────┘
    
    Mas pode falhar:
    
    PENDING ─────────────────┐
                             │
                             ▼
    BROADCAST ──────────────┌┴──────────────┐
                            │               │
                            ▼               ▼
                      CONFIRMED      FAILED ❌
                         ✅
```

---

## 🔐 Proteções Implementadas

### **1. Idempotência**
```
Se receber 2 mensagens idênticas (mesmo IdempotencyKey):
├─ Primeira: Processa normalmente
└─ Segunda: Retorna resultado da primeira (sem reprocessar)

Previne: Transferências duplicadas!
```

### **2. Circuit Breaker**
```
Se o RPC falha 5 vezes seguidas:
├─ Para de tentar
├─ Retorna erro
└─ Não desperdiça recursos
```

### **3. Validação de Endereços**
```
Antes de enviar para blockchain:
├─ Verifica formato EVM (0x...)
├─ Valida checksum
├─ Confirma que não é endereço zero
└─ Aborta se inválido
```

### **4. TTL (Time To Live) no DynamoDB**
```
Transações expiram após 30 dias:
├─ Limpa dados antigos automaticamente
├─ Economiza espaço de armazenamento
└─ Mantém apenas transações recentes
```

---

## 🧪 Exemplo Completo: Uma Transação do Início ao Fim

### **CENÁRIO: Transferir 1 ETH no Ethereum Sepolia**

```
PASSO 1: Mensagem chega na SQS
──────────────────────────────

{
  "operation_id": "op_123456",
  "chain_type": "ethereum",
  "operation_type": "transfer",
  "from_address": "0xABCD...1234",
  "to_address": "0xEFGH...5678",
  "payload": "1000000000000000000",  // 1 ETH em wei
  "idempotency_key": "key_789"
}

╔═══════════════════════════════════════════════════════════════╗
║ Lambda é acionado, função handler() é executada             ║
╚═══════════════════════════════════════════════════════════════╝


PASSO 2: Validação
─────────────────

✓ operation_id = "op_123456" → OK (único)
✓ chain_type = "ethereum" → OK (conhecida)
✓ from_address = "0xABCD...1234" → OK (válido)
✓ to_address = "0xEFGH...5678" → OK (válido)
✓ operation_type = "transfer" → OK (conhecido)
✓ payload > 0 → OK (tem valor)

╔═══════════════════════════════════════════════════════════════╗
║ Todas as validações passaram! Prosseguindo...                ║
╚═══════════════════════════════════════════════════════════════╝


PASSO 3: Verificar Idempotência
──────────────────────────────

DynamoDB Query:
  SELECT * FROM transactions 
  WHERE idempotency_key = "key_789"
  
Resultado: Nenhum registro (primeira vez)

╔═══════════════════════════════════════════════════════════════╗
║ Primeira execução! Continuar...                              ║
╚═══════════════════════════════════════════════════════════════╝


PASSO 4: Salvar Status Inicial
──────────────────────────────

DynamoDB INSERT:
{
  "operation_id": "op_123456",
  "chain_type": "ethereum",
  "status": "PENDING",
  "from_address": "0xABCD...1234",
  "to_address": "0xEFGH...5678",
  "amount": "1000000000000000000",
  "confirmations": 0,
  "created_at": "2025-12-04T23:30:00Z",
  "updated_at": "2025-12-04T23:30:00Z",
  "idempotency_key": "key_789"
}

Log:
  INFO: Transação criada em PENDING
  
╔═══════════════════════════════════════════════════════════════╗
║ Transação gravada no banco. Agora enviar para blockchain...  ║
╚═══════════════════════════════════════════════════════════════╝


PASSO 5: Chamar RPC Client
──────────────────────────

RPC Request (POST para Ethereum Sepolia):
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "eth_sendTransaction",
  "params": [{
    "from": "0xABCD...1234",
    "to": "0xEFGH...5678",
    "value": "0xDE0B6B3A7640000"  // 1 ETH em hex
  }]
}

RPC Response:
{
  "jsonrpc": "2.0",
  "result": "0xTX_HASH_123456789"  ✅
}

Log:
  INFO: Transação broadcast para Ethereum
  INFO: TX Hash: 0xTX_HASH_123456789

╔═══════════════════════════════════════════════════════════════╗
║ Transação enviada! Agora aguardar confirmações...            ║
╚═══════════════════════════════════════════════════════════════╝


PASSO 6: Monitorar Confirmações
───────────────────────────────

Loop a cada 3 segundos:

Verificação 1 (3s):
  eth_getTransactionReceipt("0xTX_HASH_123456789")
  → Resultado: null (ainda na mempool)
  → Confirmações: 0
  → Status: BROADCAST

Verificação 2 (6s):
  eth_getTransactionReceipt("0xTX_HASH_123456789")
  → Resultado: {
      "blockNumber": "0x4B7C5A",
      "gasUsed": "0x5208"
    }
  → Confirmações: 1 (minerada em 1 bloco)
  → Status: CONFIRMING

Verificação 3 (9s):
  → Confirmações: 2

Verificação 4 (12s):
  → Confirmações: 3 ✅ (Atingiu N=3 confirmações!)

Status: CONFIRMED

Log:
  INFO: Transação confirmada após 3 confirmações
  INFO: TX Hash: 0xTX_HASH_123456789
  INFO: Block: 0x4B7C5A

╔═══════════════════════════════════════════════════════════════╗
║ Transação confirmada! Atualizar no BD...                     ║
╚═══════════════════════════════════════════════════════════════╝


PASSO 7: Atualizar DynamoDB
──────────────────────────

DynamoDB UPDATE:
{
  "operation_id": "op_123456",
  "status": "CONFIRMED",
  "transaction_hash": "0xTX_HASH_123456789",
  "confirmations": 3,
  "block_number": "0x4B7C5A",
  "updated_at": "2025-12-04T23:30:12Z"
}

Log:
  INFO: Transação atualizada para CONFIRMED no BD

╔═══════════════════════════════════════════════════════════════╗
║ Tudo registrado! Agora limpar a fila...                      ║
╚═══════════════════════════════════════════════════════════════╝


PASSO 8: Deletar Mensagem da SQS
────────────────────────────────

SQS DELETE:
  receiptHandle = "AQEBwJn8..."
  
Result: Mensagem removida da fila ✅

Log:
  INFO: Mensagem removida da SQS

╔═══════════════════════════════════════════════════════════════╗
║ ✅ SUCESSO TOTAL!                                            ║
║                                                               ║
║ Transação "op_123456":                                       ║
║ - Validada                                                   ║
║ - Processada                                                 ║
║ - Confirmada na blockchain                                  ║
║ - Registrada no BD                                           ║
║ - Removida da fila                                           ║
║                                                               ║
║ Cliente pode consultar status em qualquer hora!              ║
╚═══════════════════════════════════════════════════════════════╝

STATUS FINAL:
─────────────
GET /transactions/op_123456
→ {
    "operation_id": "op_123456",
    "status": "CONFIRMED",
    "transaction_hash": "0xTX_HASH_123456789",
    "confirmations": 3,
    "from_address": "0xABCD...1234",
    "to_address": "0xEFGH...5678",
    "created_at": "2025-12-04T23:30:00Z",
    "confirmed_at": "2025-12-04T23:30:12Z"
  }
```

---

## ⚙️ Configuração de Ambiente

### **Variáveis Necessárias:**

```bash
# Blockchain
RPC_URL_ETHEREUM = "https://ethereum-sepolia-rpc.publicnode.com"
RPC_URL_POLYGON = "https://polygon-amoy-rpc.publicnode.com"
RPC_URL_BSC = "https://bsc-testnet-rpc.publicnode.com"
RPC_URL_ARBITRUM = "https://arbitrum-sepolia-rpc.publicnode.com"
RPC_URL_OPTIMISM = "https://optimism-sepolia-rpc.publicnode.com"
RPC_URL_AVALANCHE = "https://api.avax-test.network/ext/bc/C/rpc"
RPC_TIMEOUT_SECONDS = 10

# AWS
DYNAMODB_TABLE_NAME = "chainevm-transactions-dev"
SQS_QUEUE_URL = "https://sqs.us-east-1.amazonaws.com/.../..."

# Blockchain
REQUIRED_CONFIRMATIONS = 1 (testnet) ou 12 (mainnet)
REQUEST_TIMEOUT_SECONDS = 30

# Log
ENVIRONMENT = "development"
```

---

## 🚀 Resumo Executivo

**ChainEVM Lambda é um processador de transações blockchain que:**

1. **Recebe** mensagens de uma fila SQS
2. **Valida** dados e endereços
3. **Garante** idempotência (sem duplicatas)
4. **Envia** transações para múltiplos blockchains
5. **Monitora** confirmações
6. **Registra** tudo em um banco de dados
7. **Limpa** a fila após sucesso

**Fluxo em 3 segundos:**
- 0.1s: Receber e validar
- 0.5s: Enviar para blockchain
- 2.4s: Aguardar confirmações
- ✅ Sucesso!

**Analogias úteis:**
- SQS = Fila de tickets do restaurante
- Lambda = Chef que processa tickets
- RPC Client = Telefone para agência do banco
- DynamoDB = Arquivo/registro de transações
- Circuit Breaker = Disjuntor de proteção

---

## 📚 Referências de Código

| Componente | Arquivo | Função |
|-----------|---------|--------|
| Handler Principal | `cmd/lambda/main.go` | `handler()` |
| Use Case | `internal/application/usecases/execute_evm_transaction.go` | `Execute()` |
| SQS Consumer | `internal/infrastructure/eventbus/sqs_consumer.go` | `DeleteMessage()` |
| DynamoDB Repo | `internal/infrastructure/database/transaction_repository.go` | `Create()`, `Update()` |
| RPC Client | `internal/infrastructure/rpc/rpc_client.go` | `SendTransaction()` |
| Circuit Breaker | `internal/infrastructure/rpc/circuit_breaker.go` | Retry logic |

---

**Última atualização:** 2025-12-04  
**Versão:** 1.0  
**Status:** ✅ Produção
