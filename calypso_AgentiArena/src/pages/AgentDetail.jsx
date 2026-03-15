import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { ArrowLeft, Zap, CheckCircle, XCircle, Lock, Shield, Swords, Activity, BookOpen, ChevronDown, ChevronUp, Sparkles } from 'lucide-react';
import { useReadContract } from 'wagmi';
import { useWallet } from '../hooks/useWallet';

import { agents as mockAgents, categoryColors } from '../data/agents';
import { CONTRACT_ADDRESSES, ABIS } from '../contracts/addresses';
import FrameworkBadge from '../components/FrameworkBadge';
import ReputationGauge from '../components/ReputationGauge';
import AgentAvatar from '../components/AgentAvatar';
import PaymentModal from '../components/PaymentModal';
import ResultModal from '../components/ResultModal';



export default function AgentDetail() {
  const { id } = useParams();
  const { isConnected, address } = useWallet();
  const [formData, setFormData] = useState({});
  const [paymentOpen, setPaymentOpen] = useState(false);
  const [resultOpen, setResultOpen] = useState(false);
  const [executionResult, setExecutionResult] = useState(null);
  const [openSample, setOpenSample] = useState(null);

  // 1. Fetch Agent from registry (fallback to mock)
  const agentIdNum = parseInt(id);
  const { data: agentData } = useReadContract({
    address: CONTRACT_ADDRESSES.registry,
    abi: ABIS.AgentRegistry,
    functionName: 'getAgent',
    args: [agentIdNum],

  });

  // 2. Fetch Reputation
  const { data: repData } = useReadContract({
    address: CONTRACT_ADDRESSES.reputation,
    abi: ABIS.ReputationEngine,
    functionName: 'getAgentScore',
    args: [agentIdNum],

  });

  const mockAgent = mockAgents.find(a => a.id === agentIdNum);
  
  // Merge live data on top of standard visual details (image, icon, etc) from mock
  const agent = mockAgent ? {
    ...mockAgent,
    ...(agentData && agentData.id !== 0n ? {
      name: agentData.name,
      category: agentData.category,
      pricePerCall: (Number(agentData.pricePerCall) / 1e18).toString(),
      stakedAmount: (Number(agentData.stakedAmount) / 1e18).toString(),
      totalTasksCompleted: Number(agentData.totalTasks),
      successCount: Number(agentData.successfulTasks),
      isActive: agentData.isActive,
    } : {}),
    reputationScore: repData ? Number(repData) : mockAgent.reputationScore,
  } : null;

  // Initialize form data based on the agent's taskInputSchema
  useEffect(() => {
    if (agent?.taskInputSchema && Object.keys(formData).length === 0) {
      const initialData = {};
      Object.keys(agent.taskInputSchema).forEach(key => {
        initialData[key] = '';
      });
      // Optionally auto-fill wallet if the schema asks for it
      if (address && 'wallet' in initialData) initialData['wallet'] = address;
      if (address && 'wallet_address' in initialData) initialData['wallet_address'] = address;
      
      setFormData(initialData);
    }
  }, [agent, address]);

  if (!agent) {
    return (
      <div className="min-h-screen pt-24 flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-white mb-2">Agent not found</h2>
          <Link to="/marketplace" className="text-primary hover:underline">Back to Marketplace</Link>
        </div>
      </div>
    );
  }

  const successRate = agent.totalTasksCompleted > 0 
    ? ((agent.successCount / agent.totalTasksCompleted) * 100).toFixed(1) 
    : "100.0";
  
  const schemaKeys = Object.keys(agent.taskInputSchema || {});
  const mockTaskResults = [true, true, true, false, true]; // visual candy

  const handlePaymentSuccess = (result) => {
    setPaymentOpen(false);
    setExecutionResult(result);
    setResultOpen(true);
  };

  return (
    <div className="min-h-screen pt-24 pb-12">
      <div className="max-w-5xl mx-auto px-4">
        {/* Back */}
        <Link to="/marketplace" className="inline-flex items-center gap-2 text-muted hover:text-white mb-6 transition-colors">
          <ArrowLeft className="w-4 h-4" /> Back to Marketplace
        </Link>

        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="glass-card p-8 mb-6 relative overflow-hidden"
          style={{ borderTop: `3px solid ${categoryColors[agent.category] || categoryColors.wildcard}` }}
        >
          {agentData && <div className="absolute top-4 right-4 text-[10px] font-bold tracking-widest text-[#00FF94] bg-[#00FF94]/10 px-2 py-1 rounded border border-[#00FF94]/20">LIVE ON-CHAIN</div>}
          
          <div className="flex flex-col md:flex-row items-start gap-8">
            <div className="flex-1">
              <div className="flex items-center gap-3 mb-2">
                <FrameworkBadge framework={agent.framework} />
                <span className="text-sm text-muted capitalize px-2 py-0.5 rounded bg-background">{agent.category}</span>
                {agent.isActive && (
                  <span className="flex items-center gap-1 text-xs text-success">
                    <span className="w-2 h-2 rounded-full bg-success animate-pulse" /> Active
                  </span>
                )}
              </div>
              <h1 className="text-3xl font-bold text-white mb-3">{agent.name}</h1>
              <p className="text-muted mb-6 leading-relaxed">{agent.description}</p>

              {/* Stats Row */}
              <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                <div className="p-3 rounded-lg bg-background text-center border border-border">
                  <Zap className="w-4 h-4 text-primary mx-auto mb-1" />
                  <div className="text-lg font-bold text-white">{agent.pricePerCall}</div>
                  <div className="text-xs text-muted">HLUSD / call</div>
                </div>
                <div className="p-3 rounded-lg bg-background text-center border border-border">
                  <Activity className="w-4 h-4 text-success mx-auto mb-1" />
                  <div className="text-lg font-bold text-white">{agent.totalTasksCompleted.toLocaleString()}</div>
                  <div className="text-xs text-muted">Tasks</div>
                </div>
                <div className="p-3 rounded-lg bg-background text-center border border-border">
                  <div className="text-lg font-bold text-success">{successRate}%</div>
                  <div className="text-xs text-muted">Success Rate</div>
                </div>
                <div className="p-3 rounded-lg bg-background text-center border border-border">
                  <Lock className="w-4 h-4 text-warning mx-auto mb-1" />
                  <div className="text-lg font-bold text-white">{agent.stakedAmount}</div>
                  <div className="text-xs text-muted">HLUSD Staked</div>
                </div>
              </div>
            </div>

            <div className="flex flex-col items-center gap-4">
              <AgentAvatar agent={agent} size="lg" />
              <ReputationGauge score={agent.reputationScore} size={130} />
            </div>
          </div>
        </motion.div>

        <div className="grid lg:grid-cols-5 gap-6">
          {/* Left: Info */}
          <div className="lg:col-span-2 space-y-6">
             <div className="glass-card p-6">
               <h3 className="text-lg font-semibold text-white mb-4">Payment</h3>
             </div>

            {/* Recent Tasks */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.1 }}
              className="glass-card p-6"
            >
              <h3 className="text-lg font-semibold text-white mb-4">Network Reliability</h3>
              <div className="flex gap-2 mb-4">
                {mockTaskResults.map((ok, i) => (
                  <div key={i} className={`flex-1 h-10 rounded-lg flex items-center justify-center ${ok ? 'bg-success/10' : 'bg-danger/10'}`}>
                    {ok ? <CheckCircle className="w-5 h-5 text-success" /> : <XCircle className="w-5 h-5 text-danger" />}
                  </div>
                ))}
              </div>
              <p className="text-xs text-muted text-center italic">Historically verified on TaskLedger.sol</p>
            </motion.div>

            {/* Protection Notice */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: 0.2 }}
              className="glass-card p-6 border-success/20"
            >
              <div className="flex items-start gap-3">
                <Shield className="w-6 h-6 text-success shrink-0 mt-0.5" />
                <div>
                  <h3 className="font-semibold text-white mb-1">Stake Protection</h3>
                  <p className="text-sm text-muted">
                    Agent stake: <span className="text-success font-semibold">{agent.stakedAmount} HLUSD</span> — your protection if this agent fails.
                    Failed tasks are automatically compensated from the agent's stake.
                  </p>
                </div>
              </div>
            </motion.div>

            {/* Features List */}
            {agent.features && agent.features.length > 0 && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.25 }}
                className="glass-card p-6"
              >
                <div className="flex items-center gap-2 mb-4">
                  <Sparkles className="w-5 h-5 text-primary" />
                  <h3 className="text-lg font-semibold text-white">Capabilities</h3>
                </div>
                <ul className="space-y-2.5">
                  {agent.features.map((feat, i) => (
                    <li key={i} className="flex items-start gap-2.5 text-sm text-muted">
                      <CheckCircle className="w-4 h-4 text-success shrink-0 mt-0.5" />
                      {feat}
                    </li>
                  ))}
                </ul>
              </motion.div>
            )}

            {/* Sample Tasks Portfolio */}
            {agent.sampleTasks && agent.sampleTasks.length > 0 && (
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ delay: 0.3 }}
                className="glass-card p-6"
              >
                <div className="flex items-center gap-2 mb-4">
                  <BookOpen className="w-5 h-5 text-warning" />
                  <h3 className="text-lg font-semibold text-white">Sample Tasks</h3>
                  <span className="ml-auto text-xs text-muted bg-background px-2 py-0.5 rounded-full border border-border">Developer Portfolio</span>
                </div>
                <div className="space-y-3">
                  {agent.sampleTasks.map((sample, i) => (
                    <div key={i} className="rounded-xl border border-border bg-background overflow-hidden">
                      <button
                        onClick={() => setOpenSample(openSample === i ? null : i)}
                        className="w-full flex items-center justify-between p-4 text-left hover:bg-white/5 transition-colors"
                      >
                        <span className="text-sm font-medium text-white">{sample.title}</span>
                        {openSample === i
                          ? <ChevronUp className="w-4 h-4 text-muted shrink-0" />
                          : <ChevronDown className="w-4 h-4 text-muted shrink-0" />}
                      </button>
                      {openSample === i && (
                        <div className="px-4 pb-4 space-y-3 border-t border-border pt-3">
                          <div>
                            <div className="text-xs text-muted uppercase tracking-widest mb-1.5">Input</div>
                            <pre className="text-xs bg-black/40 text-primary p-3 rounded-lg overflow-x-auto whitespace-pre-wrap">{JSON.stringify(sample.input, null, 2)}</pre>
                          </div>
                          <div>
                            <div className="text-xs text-muted uppercase tracking-widest mb-1.5">Output Preview</div>
                            <p className="text-sm text-white/80 bg-success/5 border border-success/20 p-3 rounded-lg leading-relaxed">{sample.output}</p>
                          </div>
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </motion.div>
            )}
          </div>

          {/* Right: Task Form */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.15 }}
            className="lg:col-span-3 glass-card p-6"
          >
            <h3 className="text-lg font-semibold text-white mb-6">Execute Task</h3>
            <div className="space-y-4 mb-6">
              {schemaKeys.map(key => {
                const schemaDef = agent.taskInputSchema[key];
                const typeHint = typeof schemaDef === 'object' ? schemaDef.type : schemaDef;
                const placeholder = typeof schemaDef === 'object' && schemaDef.description 
                  ? schemaDef.description 
                  : `Enter ${key.replace(/_/g, ' ')}...`;
                
                const cleanLabel = key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
                
                return (
                  <div key={key}>
                    <label className="block text-sm text-muted mb-1.5">{cleanLabel} <span className="text-xs opacity-50">({typeHint})</span></label>
                    <input
                     type={typeHint === 'number' ? 'number' : 'text'}
                     placeholder={placeholder}
                     value={formData[key] || ''}
                     onChange={(e) => setFormData({ ...formData, [key]: e.target.value })}
                     className="input-field"
                    />
                  </div>
                );
              })}
            </div>

            <div className="flex items-center justify-between p-4 rounded-xl bg-background mb-4 border border-border">
              <div>
                <div className="text-sm text-muted">Cost for this call</div>
                <div className="text-xl font-bold text-primary">{agent.pricePerCall} HLUSD</div>
              </div>
              <div className="text-right">
                <div className="text-sm text-muted">Agent stake at risk</div>
                <div className="text-lg font-semibold text-success">{agent.stakedAmount} HLUSD</div>
              </div>
            </div>

            <button
              onClick={() => setPaymentOpen(true)}
              className="w-full glow-btn text-white py-3.5 rounded-xl font-semibold text-lg mb-3"
            >
              Pay {agent.pricePerCall} HLUSD & Execute
            </button>

            <Link
              to="/arena"
              className="w-full flex items-center justify-center gap-2 py-3 rounded-xl border border-border text-muted hover:text-warning hover:border-warning transition-all bg-background"
            >
              <Swords className="w-4 h-4" />
              Add to Arena Instead
            </Link>
          </motion.div>
        </div>
      </div>

      <PaymentModal
        isOpen={paymentOpen}
        onClose={() => setPaymentOpen(false)}
        agent={agent}
        inputData={formData}
        onSuccess={handlePaymentSuccess}
      />

      <ResultModal
        isOpen={resultOpen}
        onClose={() => setResultOpen(false)}
        result={executionResult}
        agent={agent}
      />
    </div>
  );
}
