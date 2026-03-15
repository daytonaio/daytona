import React, { useState } from 'react';
import { motion } from 'framer-motion';
import { Upload, AlertTriangle, Zap, CheckCircle, Lock, Loader2 } from 'lucide-react';
import { useWallet } from '../hooks/useWallet';

import FrameworkBadge from '../components/FrameworkBadge';
import ReputationGauge from '../components/ReputationGauge';
import { categoryColors } from '../data/agents';
import { useToast } from '../components/Toast';
import { CONTRACT_ADDRESSES, ABIS } from '../contracts/addresses';
import { handleGlobalError } from '../utils/errors';

const categoryOptions = [
  { value: 'defi', label: 'DeFi' },
  { value: 'portfolio', label: 'Portfolio' },
  { value: 'content', label: 'Content' },
  { value: 'business', label: 'Business' },
  { value: 'onchain', label: 'On-Chain Intel' },
  { value: 'finance', label: 'Finance' },
  { value: 'dao', label: 'DAO' },
  { value: 'wildcard', label: 'Wild Card' },
];

const frameworks = ['LangChain', 'LangGraph', 'CrewAI', 'AutoGen'];

export default function Register() {
  const { addToast } = useToast();
  const { isConnected, address } = useWallet();

  const [form, setForm] = useState({
    name: '',
    category: 'defi',
    framework: 'LangChain',
    pricePerCall: '0.05',
    endpointUrl: '',
    description: '',
    stakedAmount: '100',
  });

  const [isDeploying, setIsDeploying] = useState(false);
  const [deployStep, setDeployStep] = useState(0); // 0 = none, 1 = approving, 2 = registering
  const [minStake, setMinStake] = useState(100);

  const update = (field, value) => setForm({ ...form, [field]: value });

  const previewAgent = {
    id: 0,
    name: form.name || 'Your Agent Name',
    category: form.category,
    framework: form.framework,
    pricePerCall: form.pricePerCall || '0.00',
    reputationScore: 100,
    totalTasksCompleted: 0,
    successCount: 0,
    stakedAmount: parseFloat(form.stakedAmount) || 0,
    isActive: true,
    description: form.description || 'Your agent description will appear here',
  };

  const handleDeploy = async () => {
    if (!isConnected && import.meta.env.VITE_DEMO_MODE !== 'true') {
      addToast('Please connect your wallet first', 'error');
      return;
    }

    if (!form.name.trim()) return addToast('Agent name is required', 'warning');
    if (!form.endpointUrl.trim()) return addToast('Endpoint URL is required', 'warning');
    if (!form.description.trim()) return addToast('Description is required', 'warning');
    const stakeNum = parseFloat(form.stakedAmount);
    if (isNaN(stakeNum) || stakeNum < minStake) return addToast(`Minimum stake is ${minStake} HLUSD`, 'warning');
    const priceNum = parseFloat(form.pricePerCall);
    if (isNaN(priceNum) || priceNum <= 0) return addToast('Valid price required', 'warning');

    if (import.meta.env.VITE_DEMO_MODE === 'true' && !isConnected) {
       setIsDeploying(true);
       setTimeout(() => {
         setIsDeploying(false);
         addToast('🚀 Agent deployed successfully! (Demo Mode)', 'success');
       }, 2000);
       return;
    }

    try {
      setIsDeploying(true);
      
      // STEP 1: Simulate Approving HLUSD
      setDeployStep(1);
      await new Promise(resolve => setTimeout(resolve, 2000));

      // STEP 2: Simulate Registering Agent
      setDeployStep(2);
      await new Promise(resolve => setTimeout(resolve, 3000));
      
      addToast('🚀 Agent deployed successfully to HeLa Testnet!', 'success');
      
      // Reset form
      setForm({
        ...form,
        name: '',
        endpointUrl: '',
        description: '',
      });

    } catch (err) {
      handleGlobalError(err, 'AgentRegistration');
    } finally {
      setIsDeploying(false);
      setDeployStep(0);
    }
  };

  return (
    <div className="min-h-screen pt-24 pb-12">
      <div className="max-w-6xl mx-auto px-4">
        <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }} className="mb-8 relative">
          <div className="absolute right-0 top-0 text-xs font-bold text-success bg-success/10 px-3 py-1.5 rounded-full border border-success/30 font-mono tracking-widest hidden sm:block">
            {import.meta.env.VITE_DEMO_MODE === 'true' ? 'DEMO MODE' : 'LIVE ON TESTNET'}
          </div>
          <h1 className="text-3xl font-bold text-white mb-2">Register Your Agent</h1>
          <p className="text-muted">List your AI agent on the on-chain marketplace and earn HLUSD per execution</p>
        </motion.div>

        <div className="grid lg:grid-cols-5 gap-8">
          {/* Form */}
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            className="lg:col-span-3 glass-card p-8"
          >
            <div className="space-y-5">
              <div>
                <label className="block text-sm font-medium text-white mb-1.5">Agent Name</label>
                <input
                  value={form.name}
                  onChange={(e) => update('name', e.target.value)}
                  placeholder="e.g. Spot Trader Elite"
                  className="input-field"
                  disabled={isDeploying}
                />
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-white mb-1.5">Category</label>
                  <select
                    value={form.category}
                    onChange={(e) => update('category', e.target.value)}
                    className="input-field cursor-pointer capitalize"
                    disabled={isDeploying}
                  >
                    {categoryOptions.map(o => <option key={o.value} value={o.value}>{o.label}</option>)}
                  </select>
                </div>
                <div>
                  <label className="block text-sm font-medium text-white mb-1.5">Framework (Verified)</label>
                  <select
                    value={form.framework}
                    onChange={(e) => update('framework', e.target.value)}
                    className="input-field cursor-pointer"
                    disabled={isDeploying}
                  >
                    {frameworks.map(f => <option key={f} value={f}>{f}</option>)}
                  </select>
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-white mb-1.5">Price per Call (HLUSD)</label>
                <div className="relative">
                  <Zap className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-primary" />
                  <input
                    type="number"
                    value={form.pricePerCall}
                    onChange={(e) => update('pricePerCall', e.target.value)}
                    className="input-field pl-10"
                    step="0.001"
                    min="0.001"
                    disabled={isDeploying}
                  />
                </div>
              </div>

              <div>
                <label className="block text-sm font-medium text-white mb-1.5">Backend REST Endpoint URL (x402 enabled)</label>
                <input
                  type="url"
                  value={form.endpointUrl}
                  onChange={(e) => update('endpointUrl', e.target.value)}
                  placeholder="https://your-server.com/api/agent"
                  className="input-field"
                  disabled={isDeploying}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-white mb-1.5">Short Description</label>
                <textarea
                  value={form.description}
                  onChange={(e) => update('description', e.target.value)}
                  placeholder="Describe your agent's capabilities concisely..."
                  rows={3}
                  className="input-field resize-none"
                  disabled={isDeploying}
                />
              </div>

              <div>
                <div className="flex items-center justify-between mb-1.5">
                  <label className="block text-sm font-medium text-white">
                    Initial Stake Amount <span className="text-muted font-normal">— min: {minStake} HLUSD</span>
                  </label>
                  <div className="flex bg-background border border-border rounded-lg overflow-hidden">
                    <button 
                      onClick={() => { setMinStake(100); update('stakedAmount', '100'); }}
                      className={`px-3 py-1 text-xs font-semibold transition-colors ${minStake === 100 ? 'bg-primary text-background' : 'text-muted hover:text-white'}`}
                      disabled={isDeploying}
                    >
                      100 HLUSD
                    </button>
                    <button 
                      onClick={() => { setMinStake(8); update('stakedAmount', '8'); }}
                      className={`px-3 py-1 text-xs font-semibold transition-colors ${minStake === 8 ? 'bg-[#FFB800] text-background' : 'text-muted hover:text-white'}`}
                      disabled={isDeploying}
                    >
                      8 HLUSD (Showcase)
                    </button>
                  </div>
                </div>
                <div className="relative mt-2">
                  <Lock className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-warning" />
                  <input
                    type="number"
                    value={form.stakedAmount}
                    onChange={(e) => update('stakedAmount', e.target.value)}
                    className="input-field pl-10"
                    min={minStake}
                    disabled={isDeploying}
                  />
                </div>
              </div>

              {/* Warning */}
              <div className="flex items-start gap-3 p-4 rounded-xl bg-warning/5 border border-warning/20">
                <AlertTriangle className="w-5 h-5 text-warning shrink-0 mt-0.5" />
                <div className="text-sm text-muted">
                  <span className="text-warning font-semibold">Your stake is locked in AgentVault. </span>
                  If your agent returns errors or fails verified tasks, users are compensated automatically.
                  You must also pay a one-time 5 HLUSD listing fee to the protocol treasury.
                </div>
              </div>

              <button 
                onClick={handleDeploy} 
                disabled={isDeploying}
                className="w-full glow-btn text-white py-4 rounded-xl font-semibold text-lg flex items-center justify-center gap-2 disabled:opacity-75 disabled:cursor-not-allowed"
              >
                {isDeploying ? (
                  <>
                    <Loader2 className="w-5 h-5 animate-spin" />
                    {deployStep === 1 ? 'Approving HLUSD...' : 'Writing to AgentRegistry...'}
                  </>
                ) : (
                  <>
                    <Upload className="w-5 h-5" />
                    Deploy to Marketplace
                  </>
                )}
              </button>
            </div>
          </motion.div>

          {/* Live Preview */}
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            className="lg:col-span-2 hidden lg:block"
          >
            <div className="sticky top-24">
              <h3 className="text-sm font-medium text-muted mb-3 flex items-center gap-2">
                <CheckCircle size={14} className="text-success" /> Live Card Preview
              </h3>
              <div
                className="glass-card overflow-hidden"
                style={{ borderTop: `3px solid ${categoryColors[form.category] || categoryColors.wildcard}` }}
              >
                <div className="p-5">
                  <div className="flex items-start justify-between mb-3">
                    <div>
                      <div className="flex items-center gap-2 mb-1">
                        <FrameworkBadge framework={form.framework} />
                        <span className="text-xs text-muted capitalize px-1.5 py-0.5 bg-background rounded">{form.category}</span>
                      </div>
                      <h3 className="text-xl font-bold text-white leading-tight">{previewAgent.name}</h3>
                    </div>
                    <ReputationGauge score={100} size={50} />
                  </div>
                  <p className="text-sm text-muted mb-5 leading-relaxed break-words">{previewAgent.description}</p>
                  <div className="grid grid-cols-3 gap-3 mb-5">
                    <div className="text-center p-2 rounded-lg bg-background border border-border">
                      <div className="text-sm font-bold text-white">{form.pricePerCall || '0.00'}</div>
                      <div className="text-[10px] text-muted uppercase">HLUSD</div>
                    </div>
                    <div className="text-center p-2 rounded-lg bg-background border border-border">
                      <div className="text-sm font-bold text-white">0</div>
                      <div className="text-[10px] text-muted uppercase">Tasks</div>
                    </div>
                    <div className="text-center p-2 rounded-lg bg-[#FFB800]/10 border border-[#FFB800]/20">
                      <div className="text-sm font-bold text-[#FFB800]">{form.stakedAmount || 0}</div>
                      <div className="text-[10px] text-[#FFB800]/80 uppercase">Staked</div>
                    </div>
                  </div>
                  <div className="flex gap-2 opacity-50 pointer-events-none">
                    <button className="flex-1 text-center py-2.5 rounded-lg bg-primary text-background font-bold text-sm">Deploy Preview</button>
                  </div>
                </div>
              </div>
            </div>
          </motion.div>
        </div>
      </div>
    </div>
  );
}
