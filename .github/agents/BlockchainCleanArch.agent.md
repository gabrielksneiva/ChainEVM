# Blockchain Clean Architecture Agent — Instructions

Estas são as instruções oficiais que o GitHub Copilot (Chat, Workspace e Agents)
deve seguir **sempre que operar neste repositório**.

O objetivo é construir um sistema financeiro baseado em blockchain, com suporte
inicial a TRON (rede Shasta), seguido por BTC, ETH e SOL. A aplicação deve usar
Golang + Fiber, além de FX para injeção de dependências e ZAP para logs.

O agente deve atuar de forma **autônoma**, sem pedir confirmações, e com a
responsabilidade de manter o projeto limpo, testável e escalável.

---

## 🧱 Arquitetura Obrigatória

- O projeto **deve seguir Clean Architecture**.
- Aplicar **DDD (Domain-Driven Design)** em todos os domínios críticos.
- Utilizar **Event-Driven Architecture** para fluxos assíncronos e callbacks.
- Utilizar **FX** como mecanismo padrão de dependency injection e lifecycle.
- Utilizar **ZAP** para logs estruturados.
- Usar **Golang + Fiber** como stack principal.
- Evitar camadas desnecessárias; abstrair apenas o necessário para futuras blockchains.
- Criar arquitetura onde cada blockchain seja um módulo independente.

---

## 🔐 Blockchains – Ordem de Implementação

O agente deve implementar **nesta ordem**:

1. **TRON (Shasta)** — obrigatório para testes
2. **Bitcoin**
3. **Ethereum**
4. **Solana**

### Para cada blockchain:
Implementar 100% das funcionalidades essenciais:

- Criação de carteiras  
- Geração de chaves  
- Assinatura de transações  
- Envio de transações  
- Consulta de saldo  
- Consulta de status  
- Callbacks de atualização de status

---

## 📄 Documentação Obrigatória por Blockchain

Após finalizar cada integraçao:

O agente deve criar um arquivo em:

/docs/blockchains/<nome-da-chain>.md


Esse arquivo deve conter:

- Como conectar à chain  
- Network usada (ex: TRON Shasta)  
- Protocolos utilizados  
- Endpoints  
- Modelos de transação  
- Fluxos de assinatura e envio  
- Considerações de segurança  
- Como estender o módulo  

---

## 🧪 Regras de Testes e Coverage

### O agente deve sempre:

- **Aplicar TDD rigoroso** — escrever testes primeiro.
- Manter **coverage ≥ 90% em TODO o código novo e existente**.
- Não permitir:
  - `TODO`
  - `not implemented`
  - stubs artificiais
  - mocks desnecessários
- Refatorar automaticamente qualquer código que não seja testável.

---

## ⚙️ Regras de Implementação

O agente deve:

- Agir **sem pedir confirmação**.
- Poder excluir, recriar ou reorganizar arquivos/pastas sempre que necessário.
- Refatorar para manter testabilidade, consistência e separação de domínios.
- Utilizar Redis e PostgreSQL quando necessário.
- Implementar handlers do Fiber de forma robusta, retornando erros estruturados.
- Garantir que cada módulo esteja independentemente testável.

---

## 🎯 Objetivo Final

O agente deve garantir:

- Arquitetura Clean Architecture bem definida e modular.
- TRON (Shasta) implementada com 100% das funcionalidades essenciais.
- BTC → ETH → SOL implementadas na ordem especificada.
- Testes com cobertura mínima de 90%.
- Documentação completa e gerada após cada implementação.
- Código limpo, escalável e sem partes não implementadas.
- Logs estruturados via ZAP e DI via FX.

---

## 📌 Nota Final

Estas instruções servem como **sistema de regras permanente** para este repositório.
O Copilot deve seguir rigorosamente cada item acima ao gerar código, revisar PRs,
explicar decisões ou efetuar qualquer ação dentro deste repo.

