import { handleGlobalError } from '../utils/errors';

/**
 * Core service: x402 Payment Protocol flow for AgentArena.
 * 
 * Flow:
 * 1. PROBE:  Check backend is alive (GET /health)
 * 2. PAY:    Send native HLUSD to treasury via MetaMask
 * 3. EXECUTE: POST task to agent with payment tx hash as x-payment-tx header
 * 4. RETURN:  Display AI result to user
 */

// Treasury wallet (EOA) that receives agent payments
// Using a simple EOA address so native HLUSD transfers succeed
const TREASURY_WALLET = '0x7a3BD9C4C1F85F5e0b7a3bEF910000000000eF91';

export async function executeAgent({
  agentId,
  endpoint,
  price,
  input,
  walletClient,
  publicClient,
  onStateChange
}) {
  const isDemoMode = import.meta.env.VITE_DEMO_MODE === 'true';

  try {
    // Build the fetch URL
    let fetchUrl;
    if (endpoint.startsWith('/proxy/') || endpoint.startsWith('http')) {
      fetchUrl = endpoint;
    } else {
      fetchUrl = `http://localhost:8000${endpoint}`;
    }

    // ────────────────────────────────────────────
    // STEP 1: PROBE — Check backend is reachable
    // ────────────────────────────────────────────
    if (onStateChange) onStateChange('probe');
    
    if (!isDemoMode) {
      // We just check the health endpoint to confirm the Render server is awake
      const healthUrl = fetchUrl.replace(/\/api\/v1\/execute.*/, '/health');
      try {
        const probe = await fetch(healthUrl, { method: 'GET' });
        if (!probe.ok) {
          throw new Error(`Agent backend returned ${probe.status}`);
        }
      } catch (fetchErr) {
        if (fetchErr.message.includes('Agent backend')) throw fetchErr;
        throw new Error(
          'Cannot reach the AI agent backend. The Render server may be starting up (cold start ~30s). Please retry in a moment.'
        );
      }
    }

    // ────────────────────────────────────────────
    // STEP 2: PAY — x402 Payment via native HLUSD
    // ────────────────────────────────────────────
    let txHash = '0xDEMO_MODE';

    if (!isDemoMode && window.ethereum) {
      if (onStateChange) onStateChange('paying');
      
      const accounts = await window.ethereum.request({ method: 'eth_accounts' });
      if (!accounts || accounts.length === 0) {
        throw new Error('Wallet not connected. Please connect MetaMask first.');
      }
      
      const from = accounts[0];
      const priceFloat = parseFloat(price);
      const weiValue = BigInt(Math.floor(priceFloat * 1e18));
      const hexValue = '0x' + weiValue.toString(16);

      // Send native HLUSD to treasury — MetaMask popup appears here
      txHash = await window.ethereum.request({
        method: 'eth_sendTransaction',
        params: [{
          from: from,
          to: TREASURY_WALLET,
          value: hexValue,
        }]
      });
      
      // Wait for on-chain confirmation
      if (onStateChange) onStateChange('confirming');
      let receipt = null;
      for (let i = 0; i < 30; i++) {
        await new Promise(r => setTimeout(r, 2000));
        receipt = await window.ethereum.request({
          method: 'eth_getTransactionReceipt',
          params: [txHash]
        });
        if (receipt) break;
      }
      if (!receipt) {
        throw new Error('Transaction confirmation timed out. Check your wallet.');
      }
    }

    // ────────────────────────────────────────────
    // STEP 3: EXECUTE — Send task to AI agent
    // ────────────────────────────────────────────
    if (onStateChange) onStateChange('executing');
    
    const startTime = Date.now();
    const result = await fetch(fetchUrl, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'x-payment-tx': txHash
      },
      body: JSON.stringify(input)
    });
    const duration = ((Date.now() - startTime) / 1000).toFixed(1);
    
    if (!result.ok) {
      let errMsg = `Agent returned error ${result.status}`;
      try {
        const errBody = await result.json();
        errMsg = errBody.detail?.message || errBody.message || errBody.detail || errMsg;
      } catch { /* ignore */ }
      throw new Error(errMsg);
    }

    const rawResult = await result.json();

    // Normalize the response so the UI always has what it needs
    const normalizedResult = {
      taskId: txHash !== '0xDEMO_MODE' ? txHash : `0xTASK_${Date.now().toString(16)}`,
      agentId: agentId,
      status: rawResult.status || 'success',
      agent: rawResult.agent || 'agent',
      duration: duration,
      txHash: txHash,
      // The actual AI output — backends return it in `data` or `result`
      data: rawResult.data || rawResult.result || rawResult
    };

    try {
      const history = JSON.parse(localStorage.getItem('arenaHistory') || '[]');
      history.unshift({
        id: normalizedResult.taskId.substring(0, 10),
        agentId: agentId,
        cost: price,
        executionTime: Math.floor(Date.now() / 1000),
        statusNum: 1, // 1 = Success
        result: normalizedResult.data,
        txHash: txHash
      });
      localStorage.setItem('arenaHistory', JSON.stringify(history));
    } catch (e) {
      console.warn('Failed to save history', e);
    }

    return normalizedResult;

  } catch (err) {
    if (isDemoMode) {
      console.warn('Demo mode fallback');
      await new Promise(r => setTimeout(r, 1500));
      return {
        taskId: `0xDEMO_${Date.now().toString(16)}`,
        agentId, status: "success", agent: "demo",
        duration: "1.5", txHash: "0xDEMO_MODE",
        data: {
          messages: ["Demo mode: Agent executed successfully."],
          content: "### Demo Result\n\nThis is a simulated execution in demo mode."
        }
      };
    }
    handleGlobalError(err);
    throw err;
  }
}
