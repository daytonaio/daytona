import React, { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Wallet, Activity, DollarSign, CheckCircle, XCircle, AlertTriangle, ChevronDown, ChevronUp, Clock, ExternalLink } from 'lucide-react';
import { useReadContract } from 'wagmi';
import { useWallet } from '../hooks/useWallet';
import { formatEther } from 'viem';
import { CONTRACT_ADDRESSES, ABIS } from '../contracts/addresses';
import { agents as mockAgents } from '../data/agents';
import { mockSchedules } from '../data/mockTasks';
import { formatAgentResult } from '../utils/resultFormatter';

const statusLabels = {
  0: 'Pending',
  1: 'Success',
  2: 'Failed',
  3: 'Disputed'
};

export default function Dashboard() {
  const { isConnected, address } = useWallet();
  const [expandedRow, setExpandedRow] = useState(null);

  // Fetch Live Task History for the User
  const { data: userTasksRaw } = useReadContract({
    address: CONTRACT_ADDRESSES.taskLedger,
    abi: ABIS.TaskLedger,
    functionName: 'getUserTasks',
    args: [address],
    query: { enabled: !!address },
  });

  const { data: onChainAgents } = useReadContract({
    address: CONTRACT_ADDRESSES.registry,
    abi: ABIS.AgentRegistry,
    functionName: 'getAllAgents',
  });

  // Calculate stats & format tasks
  let tasksRun = 0;
  let hlusdSpent = 0;
  let successful = 0;
  let disputes = 0;
  
  const formattedTasks = [];

  if (userTasksRaw && Array.isArray(userTasksRaw)) {
    // Reverse mapping to show newest first
    const reversed = [...userTasksRaw].reverse();
    
    reversed.forEach(raw => {
      const stringId = raw.id.toString();
      const costRaw = Number(formatEther(raw.paymentAmount));
      
      const statusNum = raw.status;

      // Find Agent Name
      const aId = Number(raw.agentId);
      let aName = `Agent #${aId}`;
      let aCategory = 'wildcard';
      
      const chainAgent = onChainAgents?.find(a => Number(a.id) === aId);
      if (chainAgent) {
        aName = chainAgent.name;
        aCategory = chainAgent.category;
      } else {
         const mockA = mockAgents.find(a => a.id === aId);
         if (mockA) {
           aName = mockA.name;
           aCategory = mockA.category;
         }
      }

      // Try parse resultCID as JSON
      let parsedResultCID = raw.resultCID;
      try {
         if (raw.resultCID.startsWith('{')) {
           parsedResultCID = JSON.parse(raw.resultCID);
         }
      } catch (e) {}

      formattedTasks.push({
        id: stringId,
        agentId: aId,
        agentName: aName,
        category: aCategory,
        cost: costRaw.toFixed(2),
        executionTime: Number(raw.executionTime),
        statusNum: statusNum,
        result: parsedResultCID
      });
    });
  }

  // Merge with local storage history
  try {
    const localHistory = JSON.parse(localStorage.getItem('arenaHistory') || '[]');
    localHistory.forEach(raw => {
      // don't duplicate if for some reason it's already in the smart contract results
      if (formattedTasks.find(t => t.id === raw.id)) return;

      const aId = Number(raw.agentId);
      let aName = `Agent #${aId}`;
      let aCategory = 'wildcard';
      
      const mockA = mockAgents.find(a => a.id === aId);
      if (mockA) {
        aName = mockA.name;
        aCategory = mockA.category;
      }

      formattedTasks.push({
        id: raw.id,
        agentId: aId,
        agentName: aName,
        category: aCategory,
        cost: Number(raw.cost || 0).toFixed(2),
        executionTime: Number(raw.executionTime),
        statusNum: raw.statusNum,
        result: raw.result
      });
    });
  } catch (e) {
    console.error('Failed to parse local history', e);
  }

  // Sort unified tasks by execution time descending
  formattedTasks.sort((a, b) => b.executionTime - a.executionTime);

  // Recalculate stats cleanly based on unified history
  formattedTasks.forEach(t => {
    tasksRun++;
    hlusdSpent += Number(t.cost);
    if (t.statusNum === 1) successful++;
    if (t.statusNum === 3) disputes++;
  });

  if (!isConnected) {
    return (
      <div className="min-h-screen pt-24 flex items-center justify-center">
        <motion.div initial={{ opacity: 0 }} animate={{ opacity: 1 }} className="text-center glass-card p-12 max-w-md mx-4">
          <Wallet className="w-16 h-16 text-primary mx-auto mb-4 animate-pulse" />
          <h2 className="text-2xl font-bold text-white mb-2">Connect Your Wallet</h2>
          <p className="text-muted mb-6">Connect your Web3 wallet to securely view your private TaskLedger history.</p>
        </motion.div>
      </div>
    );
  }

  const stats = [
    { icon: <Activity className="w-5 h-5 text-primary" />, label: 'Tasks Run', value: tasksRun.toLocaleString(), color: 'text-primary' },
    { icon: <DollarSign className="w-5 h-5 text-warning" />, label: 'HLUSD Spent', value: hlusdSpent.toFixed(2), color: 'text-warning' },
    { icon: <CheckCircle className="w-5 h-5 text-success" />, label: 'Successful', value: successful.toLocaleString(), color: 'text-success' },
    { icon: <AlertTriangle className="w-5 h-5 text-danger" />, label: 'Disputes/Failed', value: disputes.toLocaleString(), color: 'text-danger' },
  ];

  return (
    <div className="min-h-screen pt-24 pb-12">
      <div className="max-w-6xl mx-auto px-4">
        <motion.div initial={{ opacity: 0, y: -10 }} animate={{ opacity: 1, y: 0 }} className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2">My Dashboard</h1>
          <p className="text-muted">On-chain activity overview sourced from TaskLedger.sol</p>
        </motion.div>

        {/* Stats */}
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
          {stats.map((stat, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ delay: i * 0.1 }}
              className="glass-card p-5"
            >
              <div className="flex items-center gap-3 mb-3">
                <div className="w-10 h-10 rounded-xl bg-background border border-border flex items-center justify-center">{stat.icon}</div>
              </div>
              <div className={`text-2xl font-bold ${stat.color}`}>{stat.value}</div>
              <div className="text-sm text-muted">{stat.label}</div>
            </motion.div>
          ))}
        </div>

        {/* Task History */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.2 }}
          className="glass-card mb-8"
        >
          <div className="p-6 border-b border-border flex justify-between items-center bg-[#0A0A0F]/50 rounded-t-2xl">
            <h2 className="text-xl font-semibold text-white flex items-center gap-2">
               <Activity className="w-5 h-5 text-primary" /> Immutable Task History
            </h2>
            <div className="text-xs text-[#00FF94] font-mono tracking-widest bg-[#00FF94]/10 px-2 py-1 rounded">LIVE LEDGER</div>
          </div>
          <div className="divide-y divide-border">
            {formattedTasks.length === 0 ? (
                <div className="p-12 text-center text-muted">
                   <Activity className="w-12 h-12 mx-auto mb-3 opacity-20" />
                   <p>No tasks found for your wallet address yet.</p>
                   <p className="text-sm mt-1">Execute a task in the marketplace to see it permanently recorded here.</p>
                </div>
            ) : formattedTasks.map((task) => (
              <div key={task.id}>
                <div
                  className="p-4 flex items-center justify-between cursor-pointer hover:bg-white/[0.02] transition-colors"
                  onClick={() => setExpandedRow(expandedRow === task.id ? null : task.id)}
                >
                  <div className="flex items-center gap-4">
                    {task.statusNum === 1 ? (
                      <CheckCircle className="w-5 h-5 text-success shrink-0" />
                    ) : task.statusNum === 0 ? (
                      <Clock className="w-5 h-5 text-warning shrink-0" />
                    ) : (
                      <AlertTriangle className="w-5 h-5 text-danger shrink-0" />
                    )}
                    <div>
                      <div className="font-medium text-white">{task.agentName} <span className="text-xs text-muted ml-1">#{task.id}</span></div>
                      <div className="text-xs text-muted capitalize">{task.category}</div>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    <div className="text-right hidden sm:block">
                      <div className="text-sm font-bold text-white">{task.cost} HLUSD</div>
                      <div className="text-xs text-muted">{new Date(task.executionTime * 1000).toLocaleString()}</div>
                    </div>
                    {task.statusNum === 2 && (
                      <button 
                         onClick={(e) => { e.stopPropagation(); alert('Dispute protocol simulation triggered (Requires DisputeResolver connection)'); }}
                         className="px-3 py-1 rounded-lg bg-danger/10 text-danger text-xs font-semibold hover:bg-danger/20 transition-colors border border-danger/30"
                      >
                        Open Dispute
                      </button>
                    )}
                    {expandedRow === task.id ? (
                      <ChevronUp className="w-4 h-4 text-muted" />
                    ) : (
                      <ChevronDown className="w-4 h-4 text-muted" />
                    )}
                  </div>
                </div>

                <AnimatePresence>
                  {expandedRow === task.id && (
                    <motion.div
                      initial={{ height: 0, opacity: 0 }}
                      animate={{ height: 'auto', opacity: 1 }}
                      exit={{ height: 0, opacity: 0 }}
                      className="overflow-hidden"
                    >
                      <div className="px-4 pb-4 ml-9 mt-2">
                        <div className="bg-background rounded-xl p-4 text-sm border border-border">
                          
                          <div className="flex justify-between items-center mb-3">
                             <div className="text-xs text-muted uppercase font-bold tracking-wider">Execution Output</div>
                             <div className="text-[10px] bg-primary/20 text-primary px-2 py-0.5 rounded border border-primary/30">STATUS: {statusLabels[task.statusNum].toUpperCase()}</div>
                          </div>

                          <div className="mb-2 w-full">
                            {formatAgentResult(task.category, typeof task.result === 'object' ? task.result : { content: task.result })}
                          </div>

                        </div>
                      </div>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>
            ))}
          </div>
        </motion.div>

        {/* Active Schedules - Leaving simulated for now since cron is off-chain */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ delay: 0.3 }}
          className="glass-card opacity-50 grayscale hover:grayscale-0 hover:opacity-100 transition-all"
        >
          <div className="p-6 border-b border-border flex justify-between items-center">
            <h2 className="text-xl font-semibold text-white flex items-center gap-2"><Clock /> Active Subscriptions</h2>
            <div className="text-xs text-warning tracking-widest uppercase">Mock UI</div>
          </div>
          <div className="divide-y divide-border pointer-events-none">
            {mockSchedules.map(s => (
              <div key={s.id} className="p-4 flex items-center justify-between">
                <div className="flex items-center gap-4">
                  <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center">
                    <Clock className="w-5 h-5 text-primary" />
                  </div>
                  <div>
                    <div className="font-medium text-white">{s.description}</div>
                    <div className="text-sm text-muted">{s.schedule} — {s.agentName}</div>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </motion.div>
      </div>
    </div>
  );
}
