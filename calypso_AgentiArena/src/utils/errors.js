import toast from 'react-hot-toast';

export const handleGlobalError = (error, context = '') => {
  console.error(`[AgentArena Error] ${context}:`, error);

  const errorMsg = error.message || error.toString();
  
  if (errorMsg.includes('User denied transaction signature') || errorMsg.includes('Rejected')) {
    // Silent dismiss for user rejections to avoid annoyance
    return;
  }
  
  if (errorMsg.includes('insufficient funds')) {
    toast.error('Insufficient HLUSD or gas. Please use the faucet.', { duration: 5000 });
    return;
  }
  
  if (errorMsg.includes('unsupported network') || errorMsg.includes('chain ID')) {
    toast.error('Wrong network — please switch to HeLa Testnet.', { duration: 5000 });
    return;
  }

  if (errorMsg.includes('Failed to fetch')) {
    toast.error('Backend is unreachable. Are you in Demo Mode without the server running?', { duration: 6000 });
    return;
  }

  toast.error(`Error: ${errorMsg.substring(0, 60)}${errorMsg.length > 60 ? '...' : ''}`);
};
