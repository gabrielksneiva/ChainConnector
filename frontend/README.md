# ChainConnector Frontend

Este é o front-end do ChainConnector, construído com React 18, Vite e JavaScript.

## Funcionalidades Planejadas

- Dashboard principal com overview
- Gerenciamento de transações (criar, listar, status)
- Controle de saldos offchain
- Monitoramento de endereços de interesse
- Visualização de logs de blockchain

## Estrutura do Projeto

```
frontend/
├── src/
│   ├── components/     # Componentes reutilizáveis
│   ├── pages/          # Páginas da aplicação
│   ├── hooks/          # Hooks customizados
│   ├── services/       # Serviços para API
│   ├── types/          # Tipos (futuro, quando migrar para TS)
│   ├── App.jsx         # Componente principal
│   ├── main.jsx        # Ponto de entrada
│   └── index.css       # Estilos globais
├── Dockerfile          # Build multi-stage
├── nginx.conf          # Configuração Nginx
├── package.json        # Dependências
├── vite.config.js      # Configuração Vite
└── index.html          # HTML base
```

## Desenvolvimento

### Pré-requisitos

- Node.js >= 18.0.0
- npm ou yarn

### Instalação

```bash
cd frontend
npm install
```

### Desenvolvimento Local

```bash
npm run dev
```

A aplicação estará disponível em `http://localhost:5173`.

### Build para Produção

```bash
npm run build
```

### Preview do Build

```bash
npm run preview
```

## Docker

### Build da Imagem

```bash
docker build -t chainconnector-frontend .
```

### Execução com Docker Compose

O frontend está integrado ao `docker-compose.yml` na raiz do projeto. Para executar toda a stack:

```bash
docker-compose up --build
```

O frontend estará disponível em `http://localhost:8080`.

## Integração com Backend

O Nginx está configurado para proxy requests `/api/*` para o serviço backend em `chainconnector:3000`.

## Notas

- Atualmente usando JavaScript devido a limitações de ambiente. Planeja-se migrar para TypeScript futuramente.
- O build requer Node.js >= 18. Certifique-se de ter a versão correta instalada.