# Sistema de Orquestração de Agentes - ChainConnector

## 🎯 Visão Geral

Este projeto utiliza um sistema avançado de orquestração de agentes especializados para coordenar todas as aspectos de desenvolvimento, deployment e operação do **ChainConnector**, um serviço Go de integração com blockchain e persistência em **PostgreSQL**.

## 👥 Time Disponível

### 1. **Orquestrador de Projeto** (Agente Principal)
- **Função**: Coordena todos os outros agentes
- **Invocável**: SIM (selecionável no menu de agentes)
- **Responsabilidades**: Analisar requisições, delegar para especialistas, consolidar respostas

### 2. **Project Manager (PM)**
- **Especialidade**: Gestão de escopo, cronograma, riscos
- **Invocável**: Apenas pelo orquestrador
- **Output**: Timelines, dependências, priorização

### 3. **Tech Lead**
- **Especialidade**: Golang, blockchain e PostgreSQL
- **Invocável**: Apenas pelo orquestrador
- **Output**: Arquitetura, padrões, decisões técnicas

### 4. **DevOps**
- **Especialidade**: Infraestrutura, CI/CD, Golang, blockchain e PostgreSQL
- **Invocável**: Apenas pelo orquestrador
- **Output**: Pipelines, deployment, monitoring

### 5. **QA Engineer**
- **Especialidade**: Testes, qualidade, validação
- **Invocável**: Apenas pelo orquestrador
- **Output**: Plano de testes, casos de teste, métricas

### 6. **CyberSecurity Specialist**
- **Especialidade**: Segurança, compliance, proteção
- **Invocável**: Apenas pelo orquestrador
- **Output**: Análise de ameaças, recomendações, checklist

### 7. **Blockchain Specialist**
- **Especialidade**: Web3, smart contracts, descentralização
- **Invocável**: Apenas pelo orquestrador
- **Output**: Seleção de chain, arquitetura de contratos, tokenomics

### 8. **Developer Golang**
- **Especialidade**: Implementação backend em Go
- **Invocável**: Apenas pelo orquestrador
- **Output**: Código, testes, documentação

### 9. **Developer PostgreSQL**
- **Especialidade**: Design e otimização de banco de dados
- **Invocável**: Apenas pelo orquestrador
- **Output**: Schema SQL, queries otimizadas, migrations

### 10. **Frontend Architect**
- **Especialidade**: Arquitetura frontend em React e integração com APIs
- **Invocável**: Apenas pelo orquestrador
- **Output**: Estrutura de componentes, organização de código, padrões de frontend

### 11. **Frontend Developer**
- **Especialidade**: Implementação de UI, componentes e lógica cliente
- **Invocável**: Apenas pelo orquestrador
- **Output**: Componentes, interações, chamadas API e testes de frontend

### 12. **UX/UI Designer**
- **Especialidade**: Experiência do usuário, layout e consistência visual
- **Invocável**: Apenas pelo orquestrador
- **Output**: Recomendações de design, fluxo de usuário e hierarquia visual

### 13. **Frontend QA**
- **Especialidade**: Testes de interface e regressão de frontend
- **Invocável**: Apenas pelo orquestrador
- **Output**: Casos de teste, validação de UX e relatórios de qualidade

### 14. **Accessibility Specialist**
- **Especialidade**: Acessibilidade para leitores de tela e navegação por teclado
- **Invocável**: Apenas pelo orquestrador
- **Output**: Checklist de acessibilidade e correções recomendadas

### 15. **Frontend Performance Specialist**
- **Especialidade**: Otimização de velocidade e performance de frontend
- **Invocável**: Apenas pelo orquestrador
- **Output**: Diagnóstico de performance e ações de otimização

## 🚀 Como Usar

### Passo 1: Selecione o Orquestrador
1. No Copilot Chat do VS Code, abra o seletor de agentes
2. Procure por "**Orquestrador de Projeto**"
3. Selecione-o como seu agente

### Passo 2: Descreva Sua Necessidade
Envie uma mensagem com sua requisição. Exemplos:

```
"Preciso implementar uma API REST em Golang que se conecta a PostgreSQL para
gerenciar contratos inteligentes. Quero começar com design e timeline."
```

```
"Temos um smart contract em Solidity que precisa ser integrado com backend em Go.
Defina a arquitetura completa, segurança e plano de deployment."
```

```
"Estamos migrando para Kubernetes. Quero um plano de CI/CD, disaster recovery
e considerações de segurança para aplicação Golang + PostgreSQL."
```

### Passo 3: Orquestrador Trabalha
O orquestrador irá:
1. ✅ Analisar sua requisição
2. ✅ Identificar especialistas relevantes
3. ✅ Delegar para cada agente
4. ✅ Agregar respostas
5. ✅ Apresentar plano consolidado

## 📋 Exemplos de Fluxos

### Fluxo 1: Nova Feature
```
Requisição do Usuário
    ↓
Orquestrador analisa
    ↓
PM: Define escopo e timeline
Tech Lead: Propõe arquitetura
Golang Dev: Planeja implementação
PostgreSQL Dev: Desenha schema
QA: Define testes
DevOps: Setup CI/CD
    ↓
Orquestrador consolida plano
    ↓
Resposta ao usuário com passos concretos
```

### Fluxo 2: Segurança e Compliance
```
Requisição do Usuário (questão de segurança)
    ↓
Orquestrador analisa
    ↓
CyberSec: Análise de ameaças
Tech Lead: Avalia impacto arquitetura
DevOps: Implementação de controles
QA: Validação de segurança
    ↓
Orquestrador consolida recomendações
    ↓
Resposta ao usuário com ações necessárias
```

### Fluxo 3: Projeto Blockchain
```
Requisição do Usuário (integração Web3)
    ↓
Orquestrador analisa
    ↓
Blockchain Specialist: Recomenda chain e arquitetura
Tech Lead: Integração com backend Go
Golang Dev: Implementação de interação com contratos
CyberSec: Riscos Web3 específicos
    ↓
Orquestrador consolida plano
    ↓
Resposta ao usuário com estratégia blockchain
```

## 🎨 Princípios de Operação

### Cooperação
- Agentes trabalham juntos, não em silos
- Cada um traz sua perspectiva especializada
- Conflitos são resolvidos pelo orquestrador

### Transparência
- Você vê o que cada agente contribuiu
- Justificativas para decisões são claras
- Riscos e trade-offs são expostos

### Qualidade > Velocidade
- Decisões bem pensadas são preferidas
- Segurança nunca é sacrificada
- Arquitetura sustentável é importante

### Especialização
- Cada agente é especialista em seu domínio
- Ninguém faz o que não é seu domínio
- Conhecimento é profundo, não superficial

## 📁 Estrutura de Arquivos

```
.github/agents/
├── orquestrador.agent.md          # Agente orquestrador principal
├── pm.agent.md                    # Project Manager
├── tech-lead.agent.md             # Tech Lead (Golang/PostgreSQL)
├── devops.agent.md                # DevOps Engineer
├── qa.agent.md                    # QA Engineer
├── cybersec.agent.md              # CyberSecurity Specialist
├── blockchain-specialist.agent.md # Blockchain Specialist
├── golang-dev.agent.md            # Developer Golang
├── postgres-dev.agent.md          # Developer PostgreSQL
├── frontend-architect.agent.md    # Frontend Architect (não aplicável neste repositório)
├── frontend-developer.agent.md    # Frontend Developer (não aplicável neste repositório)
├── ux-ui-designer.agent.md        # UX/UI Designer (não aplicável neste repositório)
├── frontend-qa.agent.md           # Frontend QA (não aplicável neste repositório)
├── accessibility-specialist.agent.md # Accessibility Specialist (não aplicável neste repositório)
├── frontend-performance.agent.md  # Frontend Performance Specialist (não aplicável neste repositório)
└── README.md                       # Esta documentação
```

## 🔄 Ciclo de Vida

1. **Requisição**: Usuário seleciona Orquestrador e envia mensagem
2. **Análise**: Orquestrador compreende o contexto e escopo
3. **Delegação**: Orquestrador invoca agentes relevantes como subagentes
4. **Execução**: Cada agente trabalha em sua especialidade
5. **Consolidação**: Orquestrador agrega insights
6. **Comunicação**: Resposta clara e acionável ao usuário

## 🛠️ Customização

Todos os agentes são customizáveis. Para modificar:

1. Edite o arquivo `.agent.md` correspondente
2. Atualize a `description` (usada para descoberta)
3. Modifique o corpo de instrções conforme necessário
4. As mudanças são aplicadas automaticamente

## ⚙️ Configuração Técnica

- **Cada agente** é um arquivo `.agent.md` em `.github/agents/`
- **Orquestrador** pode invocar todos via field `agents: [...]`
- **Ferramentas disponíveis** são restritas por especialidade
- **Modelo** é configurado automaticamente (fallback disponível)

## 🧰 Ferramentas WSL Disponíveis

No ambiente WSL atual, os seguintes utilitários de terminal estão disponíveis e podem ser usados por agentes com a ferramenta `execute` habilitada:

- `trivy` — scanner de vulnerabilidades de contêiner e dependências
- `go` — compilação e análise de código Go
- `docker`, `docker compose`, `docker-compose` — containers e orquestração
- `node`, `npm` — frontend e build tools
- `git` — controle de versão
- `curl` — chamadas HTTP e verificações de healthcheck
- `python3` — scripts de automação e testes
- `jq` — processamento JSON
- `make` — automação de build
- `awk`, `sed` — manipulação de texto

Agentes como DevOps, CyberSec, Tech Lead, Golang Dev e QA agora podem usar essas ferramentas quando fizer sentido para a tarefa.

## 📞 Suporte

Cada agente tem restrições claras (o que NÃO fazer) e competências bem definidas. Se precisar de um agente específico para uma tarefa, pode tentar chamar através do orquestrador.

---

**Versão**: 1.0  
**Data**: Maio de 2026  
**Status**: Ativo e Operacional
