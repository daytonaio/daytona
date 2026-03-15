import { useState, useEffect, useCallback } from 'react';

const HELA_TESTNET_CHAIN_ID = '0xa2d08'; // 666888

export function useWallet() {
  const [address, setAddress] = useState(null);
  const [balance, setBalance] = useState('0.00');
  const [isConnected, setIsConnected] = useState(false);

  const fetchBalance = useCallback(async (addr) => {
    try {
      // Check Chain ID first
      const chainId = await window.ethereum.request({ method: 'eth_chainId' });
      if (chainId !== HELA_TESTNET_CHAIN_ID) {
        console.warn('Wrong network. Please switch to HeLa Official Runtime Testnet (666888)');
        setBalance('0.00');
        return;
      }

      // Use native eth_getBalance — this returns the NATIVE HLUSD balance
      const rawBalance = await window.ethereum.request({
        method: 'eth_getBalance',
        params: [addr, 'latest']
      });

      if (!rawBalance || rawBalance === '0x' || rawBalance === '0x0') {
        setBalance('0.00');
      } else {
        const wei = BigInt(rawBalance);
        const balanceInHLUSD = Number(wei) / 1e18;
        setBalance(balanceInHLUSD.toFixed(3));
      }
    } catch (err) {
      console.error('fetchBalance error:', err);
      setBalance('0.00');
    }
  }, []);

  const syncAccounts = useCallback(async () => {
    if (!window.ethereum) return;
    try {
      const accounts = await window.ethereum.request({ method: 'eth_accounts' });
      if (accounts && accounts.length > 0) {
        setAddress(accounts[0]);
        setIsConnected(true);
        fetchBalance(accounts[0]);
      } else {
        setAddress(null);
        setIsConnected(false);
        setBalance('0.00');
      }
    } catch {
      // silent fail
    }
  }, [fetchBalance]);

  useEffect(() => {
    // Check on mount
    syncAccounts();

    if (!window.ethereum) return;

    const handleAccountsChanged = (accounts) => {
      if (accounts && accounts.length > 0) {
        setAddress(accounts[0]);
        setIsConnected(true);
        fetchBalance(accounts[0]);
      } else {
        setAddress(null);
        setIsConnected(false);
        setBalance('0.00');
      }
    };

    window.ethereum.on('accountsChanged', handleAccountsChanged);
    window.ethereum.on('chainChanged', syncAccounts);

    return () => {
      window.ethereum.removeListener('accountsChanged', handleAccountsChanged);
      window.ethereum.removeListener('chainChanged', syncAccounts);
    };
  }, [syncAccounts, fetchBalance]);

  const connect = async () => {
    if (!window.ethereum) {
      alert('MetaMask not found. Please install MetaMask to continue.');
      return;
    }
    try {
      const accounts = await window.ethereum.request({ method: 'eth_requestAccounts' });
      if (accounts && accounts.length > 0) {
        
        // Enforce HeLa Testnet
        const chainId = await window.ethereum.request({ method: 'eth_chainId' });
        if (chainId !== HELA_TESTNET_CHAIN_ID) {
          try {
            await window.ethereum.request({
              method: 'wallet_switchEthereumChain',
              params: [{ chainId: HELA_TESTNET_CHAIN_ID }],
            });
          } catch (switchError) {
             // If chain doesn't exist, try adding it
             if (switchError.code === 4902) {
               await window.ethereum.request({
                 method: 'wallet_addEthereumChain',
                 params: [{
                   chainId: HELA_TESTNET_CHAIN_ID,
                   chainName: 'HeLa Official Runtime Testnet',
                   nativeCurrency: { name: 'HLUSD', symbol: 'HLUSD', decimals: 18 },
                   rpcUrls: ['https://testnet-rpc.helachain.com'],
                   blockExplorerUrls: ['https://testnet-explorer.helachain.com']
                 }]
               });
             } else {
                 console.error('Failed to switch network:', switchError);
             }
          }
        }

        setAddress(accounts[0]);
        setIsConnected(true);
        fetchBalance(accounts[0]);
      }
    } catch (err) {
      if (err.code === 4001) {
        console.log('User rejected the connection request.');
      } else {
        console.error('Connect error:', err);
      }
    }
  };

  const disconnect = () => {
    setAddress(null);
    setIsConnected(false);
    setBalance('0.00');
  };

  const shortAddress = address
    ? `${address.slice(0, 6)}...${address.slice(-4)}`
    : null;

  return { address, balance, isConnected, connect, disconnect, shortAddress };
}
