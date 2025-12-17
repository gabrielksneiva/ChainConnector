## Plan: Suporte multi-chain EVM

Resumo rápido: Modificar configuração, providers e adapters para carregar um mapeamento cadeia→env/url, construir instâncias  por cadeia, introduzir um router/adapter que roteia chamadas por , e ajustar as interfaces/serviço para propagar seleção de cadeia. Seria interessante, também, criar uma gofunc para cada rede que possa ser chamada conforme necessário.

### Rules
- Não criar novos arquivos, utilizar os arquivos já existentes para implementar as mudanças.
- Manter compatibilidade retroativa: se nenhuma cadeia for especificada, usar a configuração padrão (atual).
- Garantir que todas as chamadas de envio de transação possam especificar a cadeia lógica.
- Atualizar testes unitários para cobrir os novos fluxos multi-chain.
- Seguir as convenções de código e padrões já estabelecidos no projeto, assim como SOLID, Clean Architecture, DDD, Event-Driven Architecture, Hexagonal Architecture and Clean Code.

### Steps
1. Atualizar `Config` em [internal/config/config.go] para suportar map[string]string `ChainRPC` e método de carregamento.
2. Construir instâncias  por cadeia em [internal/app/fx_adapters.go] usando o mapeamento de `Config`.
3. Criar [internal/adapters/ethereum_rpc/chain_router.go] com `ChainRouter` que implementa  e .
4. Modificar  em [internal/app/fx_adapters.go] para retornar o `ChainRouter` configurado.
5. Alterar `BlockchainTxSenderPort` em [internal/domain/ports/blockchain_txcx_sender_port.go] e `SignAndSend` em [internal/domain/service/transaction_service.go] para aceitar/propagar  ao enviar.
6. Atualizar  em [internal/app/fx_modules.go] para prover o mapeamento, registrar `ChainRouter` e manter fallback noop.
7. Ajustar workers em [internal/app/fx_workers.go] para passar  ao chamar `SignAndSend`.
8. Atualizar testes unitários em [internal/app/fx_workers_test.go], [internal/adapters/ethereum_rpc/noop_test.go], etc., para refletir as mudanças (regra: lembre de não criar nenhum arquivo novo, utilize os arquivos já criados para isso).

### Further Considerations
1. Decidir formato da variável: JSON `EVM_RPC_MAP` ou convenção `ETH_RPC_URL`/`POLYGON_RPC_URL`.
2. Implementar health-check, retries e timeouts por endpoint no `ChainRouter`.
3. Atualizar testes unitários que assumem  hardcoded e documentar a configuração.
