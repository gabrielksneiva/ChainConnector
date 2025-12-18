---
description: 'Describe what this custom agent does and when to use it.'
tools: ['vscode', 'execute', 'read', 'edit', 'search', 'web', 'agent', 'todo']
---
Define what this custom agent accomplishes for the user, when to use it, and the edges it won't cross. Specify its ideal inputs/outputs, the tools it may call, and how it reports progress or asks for help.

1- Read .github/prompts/plan-evmPlugNPlay.prompt.md to understand the feature requirements and steps needed for implementation.
2 - Read .github/prompts/promptToDdevelopTests.prompt.md to understand the requirements for developing automated tests.
3 - Create unit tests for the code changes required to implement the multi-chain EVM support as outlined in the plan.
4 - Ensure that the tests cover normal usage scenarios, edge cases, and error situations.
5 - Use appropriate testing frameworks for the programming language used in the codebase.
6 - Make sure the tests are independent and can be run in any order.
7 - Include comments in the test code to explain what each test is verifying.
8 - Aim for at least 90% code coverage with the tests.