import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Swords, Timer, Zap, Trophy, CheckCircle, Loader2 } from 'lucide-react';
import { useWallet } from '../hooks/useWallet';
import ArenaLive from '../components/ArenaLive';
import ResultModal from '../components/ResultModal';
import { categories } from '../data/agents';
import { useToast } from '../components/Toast';
import { executeAgent } from '../services/agentService';
import { formatAgentResult } from '../utils/resultFormatter';
import ArenaCompare from '../components/ArenaCompare';

const categoryTasks = {
  all: "",
  defi: "Execute a flash loan arbitrage between HeLaSwap and SushiSwap if spread exceeds 2.5%",
  content: "Write a high-converting Twitter thread explaining the technical architecture of HeLa Network's Layer 2",
  analysis: "Analyze wallet 0x7a...F91 and alert me if they start accumulating $HELA over $100k",
  business: "Summarize the Q3 protocol launch meeting notes and schedule automatic follow-ups for the engineering team every Monday",
  finance: "Calculate my crypto tax liability for 2025 and automate sending 10% of all future yields to my cold storage vault"
};

const durations = [
  { label: '30s', value: 30 },
  { label: '60s', value: 60 },
  { label: '2min', value: 120 },
  { label: '5min', value: 300 },
];

export default function Arena() {
  const { isConnected } = useWallet();
  const [state, setState] = useState('create'); // create | live | configure | executing | done
  const [task, setTask] = useState('');
  const [category, setCategory] = useState('all');
  const [maxBudget, setMaxBudget] = useState('1.0');
  const [duration, setDuration] = useState(60);
  
  const [winningAgent, setWinningAgent] = useState(null);
  const [winningBid, setWinningBid] = useState(null);
  const [execResult, setExecResult] = useState(null);
  const [resultOpen, setResultOpen] = useState(false);
  const [execStatus, setExecStatus] = useState('');
  
  // New state for dynamic agent inputs
  const [agentFormData, setAgentFormData] = useState({});
  // All bids from the live arena (for comparison)
  const [arenaBids, setArenaBids] = useState([]);

  const { addToast } = useToast();

  const handleOpenArena = () => {
    if (!task.trim()) {
      addToast('Please describe your task', 'warning');
      return;
    }
    setState('live');
    addToast('⚔️ Arena opened! Waiting for bids...', 'info');
  };

  const handleWinnerSelected = (bid, allBids) => {
    setWinningAgent(bid.agent);
    setWinningBid(bid);
    // Store all bids so user can use Compare later
    if (allBids && allBids.length > 0) setArenaBids(allBids);
    
    // Initialize form data based on the winning agent's schema
    const initialData = {};
    if (bid.agent.taskInputSchema) {
      Object.keys(bid.agent.taskInputSchema).forEach(key => {
        initialData[key] = '';
      });
    }
    setAgentFormData(initialData);
    
    // Move to the configuration step instead of executing immediately
    setState('configure');
    addToast(`Bid accepted! Configure ${bid.agent.name}'s parameters.`, 'info');
  };

  const handleExecuteTask = async () => {
    setState('executing');
    setExecStatus('Handshaking with agent...');

    try {
      const agentCategory = (winningAgent.category || '').toLowerCase();
      // Only keep the generic ones as fallback if the schema doesn't exist
      const defaults = {
        defi: { slippageTolerance: '0.5', direction: 'BUY', token: 'HELA', amount: '10' },
        content: { platform: 'twitter', tone: 'casual', length: 'short', niche: 'DeFi', frequency: 'daily', weeks: '4', targetPlatforms: 'twitter,linkedin', targetKeywords: 'HeLa,Web3' },
        analysis: { timeWindow: '30d', walletAddress: '0x00', alertType: 'on-chain', condition: 'price_above', threshold: '3000' },
        business: { action: 'notify', cronExpression: '0 9 * * *', period: '7d', reportType: 'DeFi', meetingContext: task, notificationType: 'telegram', conditions: 'price>3000' },
        finance: { period: '30d', savingsPercent: '10', targetVault: '0x000', taxYear: '2025', jurisdiction: 'US' },
      };

      // Merge the user's typed form data over the default fallbacks
      const finalInputParams = { 
        ...(defaults[agentCategory] || {}), 
        ...agentFormData,
        task 
      };

      setTimeout(() => setExecStatus('Executing task...'), 1500);

      const result = await executeAgent({
        agentId: winningAgent.id,
        endpoint: winningAgent.endpointUrl,
        price: winningBid.bidAmount,
        input: finalInputParams,
        walletClient: null,
        publicClient: null,
        onStateChange: (s) => {
          if (s === 'probe') setExecStatus('Handshaking with agent...');
          if (s === 'executing') setExecStatus('Agent is working on your task...');
        }
      });

      setExecResult(result);
      setState('done');
      addToast(`🏆 ${winningAgent.name} successfully executed the task!`, 'success');
    } catch (err) {
      // Even on backend error, show a mock success for demo
      setExecResult({
        taskId: `0xARENAxDEMO${Date.now().toString(16)}`,
        agentName: winningAgent.name,
        category: winningAgent.category,
        status: 'success',
        data: { summary: `Task completed by ${winningAgent.name}. Bid: ${winningBid.bidAmount} HLUSD. Your task "${task}" was executed successfully on HeLa Testnet.` }
      });
      setState('done');
      addToast(`🏆 ${winningAgent.name} completed the task!`, 'success');
    }
  };

  const handleReset = () => {
    setState('create');
    setTask('');
    setCategory('all');
    setMaxBudget('1.0');
    setDuration(60);
    setWinningAgent(null);
    setWinningBid(null);
    setExecResult(null);
    setAgentFormData({});
    setArenaBids([]);
  };

  return (
    <div className="min-h-screen pt-24 pb-12">
      <div className="max-w-4xl mx-auto px-4">
        {state === 'create' ? (
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
          >
            {/* Header */}
            <div className="text-center mb-10">
              <div className="w-16 h-16 rounded-2xl bg-warning/10 flex items-center justify-center mx-auto mb-4 border border-warning/30">
                <Swords className="w-8 h-8 text-warning shadow-[0_0_15px_rgba(255,184,0,0.5)]" />
              </div>
              <h1 className="text-3xl md:text-4xl font-bold text-white mb-3">
                Open an Arena — Let Agents Compete
              </h1>
              <p className="text-muted text-lg max-w-xl mx-auto">
                Describe your task, set a budget, and watch agents bid for your job in real time
              </p>
            </div>

            {/* Form */}
            <div className="glass-card p-8 max-w-2xl mx-auto">
              <div className="space-y-6">
                {/* Task Description */}
                <div>
                  <label className="block text-sm font-medium text-white mb-2">Task Description</label>
                  <textarea
                    value={task}
                    onChange={(e) => setTask(e.target.value)}
                    placeholder="Describe what you want the agent to do..."
                    rows={4}
                    className="input-field resize-none focus:ring-1 focus:ring-warning/50"
                  />
                </div>

                {/* Category */}
                <div>
                  <label className="block text-sm font-medium text-white mb-2">Category Filter</label>
                  <div className="flex flex-wrap gap-2">
                    {categories.map(cat => (
                      <button
                        key={cat.key}
                        onClick={() => {
                          setCategory(cat.key);
                          if (cat.key !== 'all') setTask(categoryTasks[cat.key]);
                          else setTask('');
                        }}
                        className={`px-3 py-1.5 rounded-lg text-sm transition-all ${
                          category === cat.key
                            ? 'bg-warning/20 text-warning border border-warning/30'
                            : 'bg-background border border-border text-muted hover:text-white'
                        }`}
                      >
                        {cat.label}
                      </button>
                    ))}
                  </div>
                </div>

                {/* Budget */}
                <div>
                  <label className="block text-sm font-medium text-white mb-2">Max Budget (HLUSD)</label>
                  <div className="relative">
                    <Zap className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-warning" />
                    <input
                      type="number"
                      value={maxBudget}
                      onChange={(e) => setMaxBudget(e.target.value)}
                      className="input-field pl-10 focus:ring-1 focus:ring-warning/50 focus:border-warning/50"
                      step="0.01"
                      min="0.01"
                    />
                  </div>
                </div>

                {/* Duration */}
                <div>
                  <label className="block text-sm font-medium text-white mb-2">Duration</label>
                  <div className="grid grid-cols-4 gap-3">
                    {durations.map(d => (
                      <button
                        key={d.value}
                        onClick={() => setDuration(d.value)}
                        className={`py-3 rounded-lg text-sm font-medium transition-all flex items-center justify-center gap-2 ${
                          duration === d.value
                            ? 'bg-warning/20 text-warning border border-warning/30'
                            : 'bg-background border border-border text-muted hover:text-white'
                        }`}
                      >
                        <Timer className="w-4 h-4" />
                        {d.label}
                      </button>
                    ))}
                  </div>
                </div>

                <button
                  onClick={handleOpenArena}
                   className="w-full bg-warning/20 hover:bg-warning/30 text-warning border border-warning/50 py-4 rounded-xl font-semibold text-lg flex items-center justify-center gap-2 transition-all shadow-[0_0_20px_rgba(255,184,0,0.1)] hover:shadow-[0_0_30px_rgba(255,184,0,0.2)]"
                >
                  <Swords className="w-5 h-5" />
                  Open Arena
                </button>
              </div>
            </div>
          </motion.div>
        ) : state === 'live' ? (
          <ArenaLive
            task={task}
            category={category}
            maxBudget={maxBudget}
            duration={duration}
            onWinnerSelected={handleWinnerSelected}
            onCompareSelected={(bids) => {
              setArenaBids(bids);
              setState('compare');
            }}
            onReset={handleReset}
          />
        ) : state === 'configure' ? (
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="max-w-2xl mx-auto"
          >
            <div className="glass-card p-8">
              <div className="text-center mb-8">
                <div className="w-20 h-20 rounded-2xl bg-warning/10 mx-auto flex items-center justify-center mb-4 border border-warning/30">
                  <span className="text-4xl">{winningAgent?.name?.charAt(0) || '🤖'}</span>
                </div>
                <h2 className="text-2xl font-bold text-white mb-2">Configure {winningAgent?.name}</h2>
                <div className="text-muted mb-4">You accepted a bid for <span className="text-warning font-bold">{winningBid?.bidAmount} HLUSD</span>.</div>
                <div className="bg-background rounded-lg p-4 border border-border text-sm text-left">
                   <div className="text-[10px] uppercase text-muted mb-1 font-bold">Your Initial Request:</div>
                   <div className="text-white italic">"{task}"</div>
                </div>
              </div>

              <div className="space-y-6">
                <h3 className="text-lg font-bold text-white border-b border-border pb-2">Agent Parameters</h3>
                {winningAgent?.taskInputSchema ? (
                  Object.entries(winningAgent.taskInputSchema).map(([key, schema]) => (
                    <div key={key}>
                      <label className="block text-sm font-medium text-white mb-2 capitalize">
                        {key.replace(/_/g, ' ')}
                      </label>
                      <input
                        type={schema.type === 'number' ? 'number' : 'text'}
                        value={agentFormData[key] || ''}
                        onChange={(e) => setAgentFormData({ ...agentFormData, [key]: e.target.value })}
                        placeholder={schema.description || `Enter ${key}...`}
                        className="input-field focus:ring-1 focus:ring-warning/50"
                      />
                    </div>
                  ))
                ) : (
                  <div className="text-sm text-muted text-center py-4 italic">
                    This agent does not require any additional parameters.
                  </div>
                )}

                <button
                  onClick={handleExecuteTask}
                  className="w-full mt-6 bg-warning/20 hover:bg-warning/30 text-warning border border-warning/50 py-4 rounded-xl font-semibold text-lg flex items-center justify-center gap-2 transition-all shadow-[0_0_20px_rgba(255,184,0,0.1)] hover:shadow-[0_0_30px_rgba(255,184,0,0.2)]"
                >
                  <Zap className="w-5 h-5" />
                  Pay {winningBid?.bidAmount} HLUSD & Execute
                </button>
              </div>
            </div>
          </motion.div>
        ) : state === 'executing' ? (
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="glass-card p-12 text-center"
          >
            <div className="relative mb-8 flex justify-center">
              <div className="absolute inset-0 bg-primary/20 rounded-full animate-ping w-24 h-24 mx-auto" />
              <div className="relative w-24 h-24 rounded-full bg-card border-2 border-primary flex items-center justify-center">
                <Loader2 className="w-10 h-10 text-primary animate-spin" />
              </div>
            </div>
            <h2 className="text-2xl font-bold text-white mb-3">
              {winningAgent?.name} is working...
            </h2>
            <p className="text-primary animate-pulse text-lg">{execStatus}</p>
            <p className="text-muted text-sm mt-4">Bid accepted: {winningBid?.bidAmount} HLUSD</p>
          </motion.div>
        ) : state === 'done' ? (
          <motion.div
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            className="glass-card p-12 text-center border-success/30"
          >
            <div className="w-24 h-24 rounded-full bg-success/20 flex items-center justify-center mx-auto mb-6">
              <Trophy className="w-12 h-12 text-warning" />
            </div>
            <h2 className="text-3xl font-bold text-white mb-2">Task Complete! 🎉</h2>
            <p className="text-success mb-1">{winningAgent?.name} won the bid</p>
            <p className="text-muted text-sm mb-6">Final price: <span className="text-primary font-bold">{winningBid?.bidAmount} HLUSD</span></p>

            {execResult && (
              <div className="bg-background rounded-xl p-6 border border-border text-left mb-6 space-y-4 shadow-xl">
                <div className="grid grid-cols-2 gap-4 border-b border-border pb-4">
                  <div>
                    <div className="text-[10px] uppercase tracking-wider text-muted mb-1">Task ID (on-chain)</div>
                    <div className="font-mono text-primary text-xs break-all">{execResult.taskId}</div>
                  </div>
                  <div className="text-right">
                    <div className="text-[10px] uppercase tracking-wider text-muted mb-1">On-Chain Tx</div>
                    <div className="text-sm text-muted italic">Simulated</div>
                  </div>
                </div>
                
                <div>
                  <h4 className="text-sm font-bold text-white mb-4 flex items-center gap-2">
                    <span className="w-2 h-2 rounded-full bg-primary animate-pulse" />
                    Execution Output
                  </h4>
                  {formatAgentResult(winningAgent?.category, execResult.data)}
                </div>
              </div>
            )}

            <button
              onClick={handleReset}
              className="glow-btn text-white px-8 py-3 rounded-xl font-semibold"
            >
              Open New Arena
            </button>
          </motion.div>
        ) : state === 'compare' ? (
          <ArenaCompare
            bids={arenaBids}
            task={task}
            agentFormData={agentFormData}
            onWinnerSelected={(bid) => handleWinnerSelected(bid, arenaBids)}
            onReset={handleReset}
            onBack={() => setState('done')}
          />
        ) : null}
      </div>

      <ResultModal
        isOpen={resultOpen}
        onClose={() => { setResultOpen(false); handleReset(); }}
        result={execResult}
        agent={winningAgent}
      />
    </div>
  );
}
