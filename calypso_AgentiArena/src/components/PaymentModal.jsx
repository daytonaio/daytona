import React, { useState, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { X, Shield, Loader2, CheckCircle, AlertCircle } from 'lucide-react';
import { executeAgent } from '../services/agentService';

import { useWallet } from '../hooks/useWallet';

export default function PaymentModal({ isOpen, onClose, agent, inputData, onSuccess }) {
  const [state, setState] = useState('confirm'); // confirm | approving | paying | executing | success | error
  const [errorMsg, setErrorMsg] = useState('');
  const [result, setResult] = useState(null);

  const { isConnected } = useWallet();

  // Reset state when opened
  useEffect(() => {
    if (isOpen) {
      setState('confirm');
      setErrorMsg('');
      setResult(null);
    }
  }, [isOpen]);

  const handleExecute = async () => {
    if (!isConnected && import.meta.env.VITE_DEMO_MODE !== 'true') {
      setErrorMsg('Please connect your wallet first');
      setState('error');
      return;
    }

    // Pass exactly what the user typed in the dynamic agent schema form
    const agentCategory = (agent.category || '').toLowerCase();
    const payload = { ...inputData };

    try {
      const execResult = await executeAgent({
        agentId: agent.id || agent.agentId,
        endpoint: agent.endpointUrl || `/api/agents/${agentCategory}/fake`,
        price: agent.pricePerCall,
        input: payload,
        walletClient: null,
        publicClient: null,
        onStateChange: (newState) => setState(newState)
      });

      setResult(execResult);
      setState('success');
    } catch (err) {
      setErrorMsg(err.message || 'Transaction failed');
      setState('error');
    }
  };


  const handleClose = () => {
    if (state === 'success') {
      onSuccess?.(result);
    }
    onClose();
  };

  if (!isOpen || !agent) return null;

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 flex items-center justify-center px-4">
        <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={state === 'success' || state === 'error' || state === 'confirm' ? handleClose : undefined} />
        
        <motion.div
          initial={{ scale: 0.9, opacity: 0, y: 20 }}
          animate={{ scale: 1, opacity: 1, y: 0 }}
          exit={{ scale: 0.9, opacity: 0, y: 20 }}
          className="relative bg-card border border-border rounded-2xl p-6 max-w-md w-full shadow-2xl overflow-hidden"
        >
          {(state === 'confirm' || state === 'success' || state === 'error') && (
            <button onClick={handleClose} className="absolute top-4 right-4 text-muted hover:text-white transition-colors z-10">
              <X className="w-5 h-5" />
            </button>
          )}

          {/* STATE: CONFIRM */}
          {state === 'confirm' && (
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }}>
              <h3 className="text-xl font-bold text-white mb-1">Confirm Payment</h3>
              <p className="text-sm text-muted mb-6">x402 Payment Protocol • {import.meta.env.VITE_DEMO_MODE === 'true' ? 'DEMO MODE' : 'LIVE'}</p>

              <div className="space-y-4 mb-6">
                <div className="flex justify-between items-center p-3 rounded-lg bg-background border border-border">
                  <span className="text-muted text-sm">Agent</span>
                  <span className="text-white font-semibold flex items-center gap-2">
                    <img src={agent.imageUrl} alt="" className="w-5 h-5 rounded-full" />
                    {agent.name}
                  </span>
                </div>
                <div className="flex justify-between items-center p-3 rounded-lg bg-background border border-border">
                  <span className="text-muted text-sm">Cost</span>
                  <span className="text-primary font-bold text-lg">{agent.pricePerCall} HLUSD</span>
                </div>
                <div className="flex items-center gap-2 p-3 rounded-lg bg-success/5 border border-success/20">
                  <Shield className="w-4 h-4 text-success" />
                  <span className="text-sm text-success">
                    Protected by {agent.stakedAmount} HLUSD agent stake
                  </span>
                </div>
              </div>

              {!isConnected && import.meta.env.VITE_DEMO_MODE !== 'true' && (
                <div className="mb-4">
                  <p className="text-danger text-sm text-center mb-2">Wallet not connected</p>
                </div>
              )}

              <button 
                onClick={handleExecute} 
                disabled={!isConnected && import.meta.env.VITE_DEMO_MODE !== 'true'}
                className="w-full glow-btn text-white py-3 rounded-xl font-semibold disabled:opacity-50 disabled:cursor-not-allowed"
              >
                Confirm & Pay
              </button>
            </motion.div>
          )}

          {/* STATE: APPROVING */}
          {state === 'approving' && (
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="flex flex-col items-center py-8 text-center">
              <Loader2 className="w-12 h-12 text-primary animate-spin mb-6" />
              <h3 className="text-lg font-bold text-white mb-2">Step 1/2: Approving HLUSD...</h3>
              <p className="text-sm text-muted">Please confirm the spending cap in your wallet.</p>
            </motion.div>
          )}

          {/* STATE: PAYING */}
          {state === 'paying' && (
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="flex flex-col items-center py-8 text-center">
              <Loader2 className="w-12 h-12 text-[#FFB800] animate-spin mb-6" />
              <h3 className="text-lg font-bold text-[#FFB800] mb-2">Step 2/2: Sending payment...</h3>
              <p className="text-sm text-muted">Please confirm the transaction to AgentVault.</p>
            </motion.div>
          )}

          {/* STATE: PROBING / EXECUTING */}
          {(state === 'executing' || state === 'probe') && (
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="flex flex-col items-center py-8 text-center">
              <div className="relative mb-6">
                <div className="absolute inset-0 bg-primary/20 rounded-full animate-ping" />
                <div className="relative bg-card border-2 border-primary rounded-full p-4">
                  <img src={agent.imageUrl} alt="" className="w-12 h-12 rounded-full animate-pulse" />
                </div>
              </div>
              <h3 className="text-lg font-bold text-white mb-2">Agent is working...</h3>
              <p className="text-sm text-primary animate-pulse">{state === 'probe' ? 'Handshaking...' : 'Executing task on-chain'}</p>
            </motion.div>
          )}

          {/* STATE: SUCCESS */}
          {state === 'success' && (
            <motion.div initial={{ opacity: 0, scale: 0.9 }} animate={{ opacity: 1, scale: 1 }} className="flex flex-col items-center py-6 text-center">
              <div className="w-20 h-20 rounded-full bg-success/20 flex items-center justify-center mb-6">
                <CheckCircle className="w-12 h-12 text-success" />
              </div>
              <h3 className="text-2xl font-bold text-white mb-2">Task Complete!</h3>
              <p className="text-sm text-muted mb-6">The agent successfully returned your result.</p>
              
              <div className="w-full bg-background rounded-lg p-3 text-left border border-border mb-6">
                <div className="text-xs text-muted mb-1">Task ID</div>
                <div className="font-mono text-[#00D4FF] text-sm break-all">{result?.taskId}</div>
              </div>

              <button onClick={handleClose} className="w-full glow-btn text-white py-3 rounded-xl font-semibold">
                View Result
              </button>
            </motion.div>
          )}

          {/* STATE: ERROR */}
          {state === 'error' && (
            <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="flex flex-col items-center py-6 text-center">
              <div className="w-20 h-20 rounded-full bg-danger/20 flex items-center justify-center mb-6">
                <AlertCircle className="w-12 h-12 text-danger" />
              </div>
              <h3 className="text-xl font-bold text-white mb-2">Execution Failed</h3>
              <p className="text-sm text-danger/80 mb-6 bg-danger/5 p-3 rounded-lg border border-danger/20 w-full">{errorMsg}</p>
              
              

              <div className="flex gap-3 w-full">
                <button onClick={handleClose} className="flex-1 px-4 py-3 rounded-xl border border-border text-white hover:bg-white/5 font-semibold">
                  Cancel
                </button>
                <button onClick={() => setState('confirm')} className="flex-1 glow-btn text-white py-3 rounded-xl font-semibold">
                  Retry
                </button>
              </div>
            </motion.div>
          )}

        </motion.div>
      </div>
    </AnimatePresence>
  );
}
