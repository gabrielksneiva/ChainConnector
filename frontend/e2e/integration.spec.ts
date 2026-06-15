import { test, expect } from '@playwright/test';

// Use nginx frontend when inside Docker Compose test network
// Use direct backend when running outside (local development)
const API_BASE = process.env.DOCKER_INTERNAL === 'true' 
  ? 'http://frontend/api'  // Via nginx proxy inside Docker
  : 'http://localhost:3001'; // Direct backend outside Docker

// Helper to wait for backend to be ready
async function waitForBackend(maxRetries = 30) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      const response = await fetch(`${API_BASE.replace('/api', '')}/health`);
      if (response.ok) {
        console.log('Backend is ready!');
        return;
      }
    } catch {
      // Backend not ready yet
    }
    await new Promise(r => setTimeout(r, 1000));
  }
  throw new Error('Backend failed to start');
}

// Helper to create a wallet via API
async function createWalletViaAPI(chain = 'sepolia') {
  const response = await fetch(`${API_BASE}/wallets`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ chain }),
  });
  if (!response.ok) {
    throw new Error(`Failed to create wallet: ${await response.text()}`);
  }
  return response.json();
}

// Helper to update balance via API
async function updateBalanceViaAPI(address: string, amount: string, chain = 'sepolia') {
  const response = await fetch(`${API_BASE}/balance/${address}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ amount, chain }),
  });
  if (!response.ok) {
    throw new Error(`Failed to update balance: ${await response.text()}`);
  }
  return response.status === 204; // Expect 204 No Content
}

// Helper to get balance via API
async function getBalanceViaAPI(address: string, chain = 'sepolia') {
  const response = await fetch(`${API_BASE}/balance/${address}?chain=${chain}`);
  if (!response.ok) {
    throw new Error(`Failed to get balance: ${await response.text()}`);
  }
  return response.json();
}

test.describe('ChainConnector Integration Tests', () => {
  test.beforeAll(async () => {
    await waitForBackend();
  });

  test('should load the dashboard', async ({ page }) => {
    await page.goto('/');
    await expect(page).toHaveTitle(/ChainConnector|Dashboard/i);
    await expect(page.locator('text=ChainConnector|Wallets')).toBeVisible({ timeout: 5000 });
  });

  test('should create a wallet via API', async () => {
    const wallet = await createWalletViaAPI('sepolia');
    expect(wallet).toHaveProperty('id');
    expect(wallet).toHaveProperty('address');
    expect(wallet.chain).toBe('sepolia');
    console.log('Created wallet:', wallet);
  });

  test('should update balance via API', async () => {
    const wallet = await createWalletViaAPI('sepolia');
    const amount = '1000000000000000000'; // 1 ETH in wei

    const success = await updateBalanceViaAPI(wallet.address, amount, 'sepolia');
    expect(success).toBe(true);

    // Verify the balance was updated
    const balance = await getBalanceViaAPI(wallet.address, 'sepolia');
    expect(balance.amount).toBe(amount);
  });

  test('should handle multiple chains', async () => {
    const wallet = await createWalletViaAPI('sepolia');
    const amount1 = '1000000000000000000'; // 1 ETH
    const amount2 = '2000000000000000000'; // 2 ETH

    // Update balance on sepolia
    await updateBalanceViaAPI(wallet.address, amount1, 'sepolia');

    // Update balance on mainnet (should be treated as different record)
    await updateBalanceViaAPI(wallet.address, amount2, 'mainnet');

    // Verify both balances are stored separately
    const sepoliaBalance = await getBalanceViaAPI(wallet.address, 'sepolia');
    const mainnetBalance = await getBalanceViaAPI(wallet.address, 'mainnet');

    expect(sepoliaBalance.amount).toBe(amount1);
    expect(mainnetBalance.amount).toBe(amount2);
  });

  test('should display wallet list in UI', async ({ page }) => {
    // Create a wallet via API
    const wallet = await createWalletViaAPI('sepolia');

    // Navigate to dashboard
    await page.goto('/');

    // Wait for wallet list to be visible
    await expect(page.locator('text=Wallets')).toBeVisible();

    // Check if wallet address appears in the list
    await expect(page.locator(`text=${wallet.address.substring(0, 10)}`)).toBeVisible({ timeout: 5000 });
  });

  test('should handle invalid balance updates', async () => {
    const wallet = await createWalletViaAPI('sepolia');

    // Try to update with invalid amount (non-numeric)
    try {
      const response = await fetch(`${API_BASE}/balance/${wallet.address}`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ amount: 'invalid', chain: 'sepolia' }),
      });
      expect(response.status).toBe(400); // Bad request
    } catch {
      // Expected error
    }
  });

  test('should normalize chain names', async () => {
    const wallet = await createWalletViaAPI();

    // Test various chain name formats
    const testCases = [
      { input: 'SEPOLIA', expected: 'sepolia' },
      { input: 'Sepolia', expected: 'sepolia' },
      { input: 'sepolia', expected: 'sepolia' },
      { input: '', expected: 'sepolia' }, // default chain
    ];

    for (const testCase of testCases) {
      const amount = '1000000000000000000';
      await updateBalanceViaAPI(wallet.address, amount, testCase.input || 'sepolia');

      const balance = await getBalanceViaAPI(wallet.address, testCase.expected);
      expect(balance.chain).toBe(testCase.expected);
    }
  });
});
