import React, { useState, useEffect, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Clock, Plus, Trash2, ToggleLeft, ToggleRight, Zap, AlertCircle, CheckCircle, ChevronDown, Timer } from 'lucide-react';
import { useWallet } from '../hooks/useWallet';
import { useScheduler, INTERVALS } from '../hooks/useScheduler';
import { agents } from '../data/agents';
import { executeAgent } from '../services/agentService';
import AgentAvatar from '../components/AgentAvatar';
import { usePublicClient, useWalletClient } from 'wagmi';

// ─── Create Schedule Modal ────────────────────────────────────────────────────
function CreateModal({ onClose, onSave }) {
  const [step, setStep] = useState(1); // 1: pick agent, 2: configure inputs + interval
  const [selectedAgent, setSelectedAgent] = useState(null);
  const [interval, setInterval] = useState('daily');
  const [inputs, setInputs] = useState({});
  const [label, setLabel] = useState('');

  const handleAgentSelect = (agent) => {
    setSelectedAgent(agent);
    // Pre-fill inputs with empty strings
    const emptyInputs = {};
    Object.keys(agent.taskInputSchema || {}).forEach(k => { emptyInputs[k] = ''; });
    setInputs(emptyInputs);
    setLabel(`${agent.name} — auto-run`);
    setStep(2);
  };

  const handleSave = () => {
    if (!selectedAgent) return;
    onSave({
      agentId: selectedAgent.id,
      agentName: selectedAgent.name,
      agentCategory: selectedAgent.category,
      endpointUrl: selectedAgent.endpointUrl,
      pricePerCall: selectedAgent.pricePerCall,
      avatar: selectedAgent.avatar,
      interval,
      inputs,
      label: label || `${selectedAgent.name} — auto-run`,
    });
    onClose();
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-md">
      <motion.div
        initial={{ scale: 0.95, opacity: 0 }}
        animate={{ scale: 1, opacity: 1 }}
        exit={{ scale: 0.95, opacity: 0 }}
        className="glass-card w-full max-w-2xl max-h-[90vh] overflow-y-auto"
      >
        <div className="p-6">
          {/* Header */}
          <div className="flex items-center justify-between mb-6">
            <div>
              <h2 className="text-xl font-bold text-white">New Scheduled Task</h2>
              <p className="text-sm text-muted mt-0.5">
                {step === 1 ? 'Step 1: Choose an agent' : 'Step 2: Configure task'}
              </p>
            </div>
            <button onClick={onClose} className="text-muted hover:text-white transition-colors text-xl leading-none">✕</button>
          </div>

          {/* Step 1: Agent Picker */}
          {step === 1 && (
            <div className="grid sm:grid-cols-2 gap-3 max-h-[60vh] overflow-y-auto pr-1">
              {agents.filter(a => a.isActive).map(agent => (
                <button
                  key={agent.id}
                  onClick={() => handleAgentSelect(agent)}
                  className="flex items-center gap-3 p-4 rounded-xl border border-border bg-background hover:border-primary/50 hover:bg-primary/5 transition-all text-left group"
                >
                  <AgentAvatar agent={agent} size="sm" />
                  <div className="flex-1 min-w-0">
                    <div className="font-semibold text-white text-sm group-hover:text-primary transition-colors truncate">{agent.name}</div>
                    <div className="text-xs text-muted truncate">{agent.pricePerCall} HLUSD / call</div>
                  </div>
                </button>
              ))}
            </div>
          )}

          {/* Step 2: Configure */}
          {step === 2 && selectedAgent && (
            <div className="space-y-5">
              {/* Selected agent pill */}
              <div className="flex items-center gap-3 p-3 rounded-xl bg-primary/5 border border-primary/20">
                <AgentAvatar agent={selectedAgent} size="sm" />
                <div>
                  <div className="font-semibold text-white text-sm">{selectedAgent.name}</div>
                  <div className="text-xs text-primary">{selectedAgent.pricePerCall} HLUSD per execution</div>
                </div>
                <button onClick={() => setStep(1)} className="ml-auto text-xs text-muted hover:text-white border border-border px-2 py-1 rounded-lg">Change</button>
              </div>

              {/* Label */}
              <div>
                <label className="block text-sm text-muted mb-1.5">Schedule Name</label>
                <input
                  className="input-field"
                  value={label}
                  onChange={e => setLabel(e.target.value)}
                  placeholder="e.g. Daily portfolio check"
                />
              </div>

              {/* Interval */}
              <div>
                <label className="block text-sm text-muted mb-1.5">Run Frequency</label>
                <div className="relative">
                  <select
                    value={interval}
                    onChange={e => setInterval(e.target.value)}
                    className="input-field appearance-none pr-10 cursor-pointer"
                  >
                    {INTERVALS.map(iv => (
                      <option key={iv.value} value={iv.value}>{iv.label}</option>
                    ))}
                  </select>
                  <ChevronDown className="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted pointer-events-none" />
                </div>
              </div>

              {/* Input fields from schema */}
              {Object.keys(selectedAgent.taskInputSchema || {}).length > 0 && (
                <div>
                  <label className="block text-sm text-muted mb-2">Task Inputs (saved for each run)</label>
                  <div className="space-y-3">
                    {Object.entries(selectedAgent.taskInputSchema).map(([key, def]) => {
                      const typeHint = typeof def === 'object' ? def.type : def;
                      const placeholder = typeof def === 'object' && def.description ? def.description : `Enter ${key}...`;
                      const cleanLabel = key.replace(/_/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
                      return (
                        <div key={key}>
                          <label className="block text-xs text-muted mb-1">{cleanLabel} <span className="opacity-50">({typeHint})</span></label>
                          <input
                            type={typeHint === 'number' ? 'number' : 'text'}
                            placeholder={placeholder}
                            value={inputs[key] || ''}
                            onChange={e => setInputs({ ...inputs, [key]: e.target.value })}
                            className="input-field text-sm"
                          />
                        </div>
                      );
                    })}
                  </div>
                </div>
              )}

              {/* Cost warning */}
              <div className="flex items-start gap-2.5 p-3 rounded-xl bg-warning/5 border border-warning/20">
                <AlertCircle className="w-4 h-4 text-warning shrink-0 mt-0.5" />
                <p className="text-xs text-warning/80">
                  Each scheduled run will require you to <strong>approve a MetaMask transaction</strong> of {selectedAgent.pricePerCall} HLUSD. You must have sufficient balance when the schedule fires.
                </p>
              </div>

              {/* Actions */}
              <div className="flex gap-3 pt-2">
                <button onClick={() => setStep(1)} className="flex-1 py-2.5 rounded-xl border border-border text-muted hover:text-white transition-colors text-sm">Back</button>
                <button onClick={handleSave} className="flex-1 glow-btn text-white py-2.5 rounded-xl font-semibold text-sm">
                  Save Schedule
                </button>
              </div>
            </div>
          )}
        </div>
      </motion.div>
    </div>
  );
}

// ─── Schedule Card ────────────────────────────────────────────────────────────
function ScheduleCard({ sched, onToggle, onDelete, nextRunLabel, isFiring }) {
  const agent = agents.find(a => a.id === sched.agentId) || { name: sched.agentName, category: sched.agentCategory, avatar: sched.avatar };
  const intervalLabel = INTERVALS.find(iv => iv.value === sched.interval)?.label || sched.interval;

  return (
    <motion.div
      layout
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      exit={{ opacity: 0, scale: 0.95 }}
      className={`glass-card p-5 transition-all ${!sched.enabled ? 'opacity-50' : ''} ${isFiring ? 'ring-2 ring-primary/50 shadow-[0_0_20px_rgba(0,212,255,0.15)]' : ''}`}
    >
      <div className="flex items-start gap-4">
        <div className="relative">
          <AgentAvatar agent={agent} size="sm" />
          {isFiring && (
            <span className="absolute -top-1 -right-1 w-3 h-3 rounded-full bg-primary animate-ping" />
          )}
        </div>

        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-0.5">
            <h3 className="font-semibold text-white text-sm truncate">{sched.label}</h3>
            {isFiring && <span className="text-xs text-primary font-bold animate-pulse">FIRING</span>}
          </div>
          <div className="text-xs text-muted mb-3">{intervalLabel} · {sched.pricePerCall} HLUSD / run</div>

          <div className="grid grid-cols-3 gap-2 text-center">
            <div className="p-2 rounded-lg bg-background/60 border border-border">
              <div className="text-xs text-muted">Runs</div>
              <div className="text-sm font-bold text-white">{sched.runCount || 0}</div>
            </div>
            <div className="p-2 rounded-lg bg-background/60 border border-border">
              <div className="text-xs text-muted">Next run</div>
              <div className="text-xs font-semibold text-primary truncate">{nextRunLabel}</div>
            </div>
            <div className="p-2 rounded-lg bg-background/60 border border-border">
              <div className="text-xs text-muted">Status</div>
              <div className={`text-xs font-bold ${sched.enabled ? 'text-success' : 'text-muted'}`}>
                {sched.enabled ? 'Active' : 'Paused'}
              </div>
            </div>
          </div>
        </div>

        <div className="flex flex-col items-center gap-2 shrink-0">
          <button onClick={() => onToggle(sched.id)} className="text-muted hover:text-primary transition-colors">
            {sched.enabled
              ? <ToggleRight className="w-7 h-7 text-primary" />
              : <ToggleLeft className="w-7 h-7" />}
          </button>
          <button onClick={() => onDelete(sched.id)} className="text-muted hover:text-danger transition-colors">
            <Trash2 className="w-4 h-4" />
          </button>
        </div>
      </div>
    </motion.div>
  );
}

// ─── Main Scheduler Page ──────────────────────────────────────────────────────
export default function Scheduler() {
  const { isConnected } = useWallet();
  const { data: walletClient } = useWalletClient();
  const publicClient = usePublicClient();

  const [showCreate, setShowCreate] = useState(false);
  const [firingIds, setFiringIds] = useState(new Set());
  const [log, setLog] = useState([]);
  const [nextRunLabels, setNextRunLabels] = useState({});

  const handleFire = useCallback(async (sched) => {
    setFiringIds(prev => new Set([...prev, sched.id]));
    const agent = agents.find(a => a.id === sched.agentId);
    if (!agent) return;

    const logEntry = {
      id: Date.now(),
      schedId: sched.id,
      agentName: sched.agentName,
      time: new Date().toLocaleTimeString(),
      status: 'pending',
      message: 'Awaiting MetaMask payment approval...'
    };
    setLog(prev => [logEntry, ...prev]);

    try {
      const result = await executeAgent({
        agentId: sched.agentId,
        endpoint: sched.endpointUrl,
        price: sched.pricePerCall,
        input: sched.inputs,
        walletClient,
        publicClient,
        onStateChange: (state) => {
          setLog(prev => prev.map(l => l.id === logEntry.id
            ? { ...l, message: state === 'probe' ? 'Connecting to agent...' : state === 'pay' ? 'Processing x402 payment...' : 'Executing agent task...' }
            : l
          ));
        }
      });

      setLog(prev => prev.map(l => l.id === logEntry.id
        ? { ...l, status: 'success', message: 'Task completed successfully!' }
        : l
      ));
    } catch (err) {
      setLog(prev => prev.map(l => l.id === logEntry.id
        ? { ...l, status: 'error', message: err.message || 'Execution failed' }
        : l
      ));
    } finally {
      setFiringIds(prev => { const n = new Set(prev); n.delete(sched.id); return n; });
    }
  }, [walletClient, publicClient]);

  const { schedules, addSchedule, toggleSchedule, removeSchedule, getNextRunTime } = useScheduler({
    onFireSchedule: handleFire,
  });

  // Refresh countdown labels every second
  useEffect(() => {
    const tick = () => {
      const labels = {};
      schedules.forEach(s => { labels[s.id] = getNextRunTime(s); });
      setNextRunLabels(labels);
    };
    tick();
    const t = setInterval(tick, 1000);
    return () => clearInterval(t);
  }, [schedules, getNextRunTime]);

  return (
    <div className="min-h-screen pt-24 pb-12">
      <div className="max-w-5xl mx-auto px-4">

        {/* Header */}
        <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }} className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <Timer className="w-8 h-8 text-primary" />
            <h1 className="text-3xl font-bold text-white">Scheduler</h1>
            <span className="text-xs font-bold tracking-widest text-success bg-success/10 border border-success/20 px-2 py-1 rounded-full">BETA</span>
          </div>
          <p className="text-muted">Set any agent to auto-run on a schedule. An x402 payment fires each time — fully on-chain.</p>
        </motion.div>

        {!isConnected ? (
          <div className="glass-card p-12 text-center">
            <Zap className="w-12 h-12 text-muted mx-auto mb-4" />
            <h3 className="text-xl font-bold text-white mb-2">Connect Your Wallet</h3>
            <p className="text-muted">You need to connect MetaMask to create and run scheduled tasks.</p>
          </div>
        ) : (
          <div className="grid lg:grid-cols-5 gap-6">
            {/* Left: Schedule List */}
            <div className="lg:col-span-3 space-y-4">
              <div className="flex items-center justify-between">
                <h2 className="text-lg font-semibold text-white">Active Schedules</h2>
                <button
                  onClick={() => setShowCreate(true)}
                  className="flex items-center gap-2 glow-btn text-white px-4 py-2 rounded-xl text-sm font-semibold"
                >
                  <Plus className="w-4 h-4" />
                  New Schedule
                </button>
              </div>

              {schedules.length === 0 ? (
                <div className="glass-card p-12 text-center">
                  <Clock className="w-12 h-12 text-muted mx-auto mb-4" />
                  <h3 className="text-lg font-bold text-white mb-2">No schedules yet</h3>
                  <p className="text-muted text-sm mb-4">Create your first recurring agent task to get started.</p>
                  <button onClick={() => setShowCreate(true)} className="glow-btn text-white px-6 py-2.5 rounded-xl text-sm font-semibold">
                    + Create Schedule
                  </button>
                </div>
              ) : (
                <AnimatePresence mode="popLayout">
                  {schedules.map(sched => (
                    <ScheduleCard
                      key={sched.id}
                      sched={sched}
                      onToggle={toggleSchedule}
                      onDelete={removeSchedule}
                      nextRunLabel={nextRunLabels[sched.id] || '—'}
                      isFiring={firingIds.has(sched.id)}
                    />
                  ))}
                </AnimatePresence>
              )}
            </div>

            {/* Right: Execution Log */}
            <div className="lg:col-span-2">
              <div className="glass-card p-5 sticky top-24">
                <h2 className="text-base font-semibold text-white mb-4 flex items-center gap-2">
                  <span className="w-2 h-2 rounded-full bg-success animate-pulse" />
                  Execution Log
                </h2>
                {log.length === 0 ? (
                  <p className="text-sm text-muted text-center py-8">No executions yet. Log will appear here when a schedule fires.</p>
                ) : (
                  <div className="space-y-3 max-h-[60vh] overflow-y-auto pr-1">
                    <AnimatePresence>
                      {log.map(entry => (
                        <motion.div
                          key={entry.id}
                          initial={{ opacity: 0, x: 10 }}
                          animate={{ opacity: 1, x: 0 }}
                          className={`p-3 rounded-xl border text-sm ${
                            entry.status === 'success' ? 'bg-success/5 border-success/20' :
                            entry.status === 'error'   ? 'bg-danger/5 border-danger/20' :
                            'bg-primary/5 border-primary/20'
                          }`}
                        >
                          <div className="flex items-center justify-between mb-1">
                            <span className="font-medium text-white text-xs truncate">{entry.agentName}</span>
                            <span className="text-muted text-xs shrink-0 ml-2">{entry.time}</span>
                          </div>
                          <div className="flex items-center gap-1.5">
                            {entry.status === 'success' && <CheckCircle className="w-3 h-3 text-success shrink-0" />}
                            {entry.status === 'error'   && <AlertCircle className="w-3 h-3 text-danger shrink-0" />}
                            {entry.status === 'pending' && <span className="w-3 h-3 border-2 border-primary border-t-transparent rounded-full animate-spin shrink-0" />}
                            <span className={`text-xs ${entry.status === 'error' ? 'text-danger/80' : 'text-muted'}`}>{entry.message}</span>
                          </div>
                        </motion.div>
                      ))}
                    </AnimatePresence>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Create Modal */}
      <AnimatePresence>
        {showCreate && <CreateModal onClose={() => setShowCreate(false)} onSave={addSchedule} />}
      </AnimatePresence>
    </div>
  );
}
