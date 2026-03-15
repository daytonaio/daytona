import React, { useState, useCallback } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import {
  GitCompare, Zap, CheckCircle, XCircle, Trophy,
  Loader2, AlertCircle, Clock, Star, TrendingUp
} from 'lucide-react';
import AgentAvatar from './AgentAvatar';
import FrameworkBadge from './FrameworkBadge';
import { formatAgentResult } from '../utils/resultFormatter';

// ─── Single side panel ───────────────────────────────────────────────────────
function ComparePanel({ bid, selected, onClick, disabled }) {
  const agent = bid.agent;
  if (!agent) return null;

  return (
    <motion.button
      whileHover={!disabled ? { y: -2 } : {}}
      onClick={() => !disabled && onClick(bid)}
      className={`w-full text-left p-5 rounded-2xl border-2 transition-all duration-200 ${
        selected
          ? 'border-primary bg-primary/10 shadow-[0_0_20px_rgba(0,212,255,0.1)]'
          : disabled
          ? 'border-border bg-background/40 opacity-50 cursor-not-allowed'
          : 'border-border bg-background/60 hover:border-primary/50 cursor-pointer'
      }`}
    >
      <div className="flex items-center gap-3 mb-3">
        <AgentAvatar agent={agent} size="sm" />
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <span className="font-bold text-white text-sm truncate">{agent.name}</span>
            <FrameworkBadge framework={agent.framework} />
          </div>
          <div className="text-xs text-muted capitalize">{agent.category}</div>
        </div>
        {selected && (
          <div className="w-5 h-5 rounded-full bg-primary flex items-center justify-center shrink-0">
            <CheckCircle className="w-3 h-3 text-background" />
          </div>
        )}
      </div>

      <div className="grid grid-cols-3 gap-2 text-center text-xs">
        <div className="p-2 rounded-lg bg-background border border-border">
          <div className="text-muted">Bid</div>
          <div className="font-bold text-warning">{bid.bidAmount} HLUSD</div>
        </div>
        <div className="p-2 rounded-lg bg-background border border-border">
          <div className="text-muted">Confidence</div>
          <div className="font-bold text-success">{bid.confidence}%</div>
        </div>
        <div className="p-2 rounded-lg bg-background border border-border">
          <div className="text-muted">Rep</div>
          <div className="font-bold text-primary">{agent.reputationScore}</div>
        </div>
      </div>
    </motion.button>
  );
}

// ─── Result column ────────────────────────────────────────────────────────────
function ResultColumn({ bid, result, status, isWinner, onVote, voted }) {
  const agent = bid.agent;
  const totalCost = parseFloat(bid.bidAmount).toFixed(4);

  const execTime = result?._execTime ? `${(result._execTime / 1000).toFixed(1)}s` : '—';

  return (
    <motion.div
      initial={{ opacity: 0, y: 16 }}
      animate={{ opacity: 1, y: 0 }}
      className={`flex-1 flex flex-col rounded-2xl border-2 overflow-hidden transition-all duration-300 ${
        isWinner
          ? 'border-warning shadow-[0_0_30px_rgba(255,184,0,0.15)]'
          : 'border-border'
      }`}
    >
      {/* Agent header */}
      <div className={`p-4 ${isWinner ? 'bg-warning/10' : 'bg-card'}`}>
        <div className="flex items-center gap-3 mb-2">
          <AgentAvatar agent={agent} size="sm" />
          <div className="flex-1 min-w-0">
            <div className="font-bold text-white text-sm truncate">{agent.name}</div>
            <div className="text-xs text-warning">{bid.bidAmount} HLUSD</div>
          </div>
          {isWinner && (
            <div className="flex items-center gap-1 text-xs font-bold text-warning bg-warning/10 border border-warning/30 px-2 py-0.5 rounded-full">
              <Trophy className="w-3 h-3" /> Winner
            </div>
          )}
        </div>

        {/* Stats row */}
        <div className="grid grid-cols-3 gap-1.5 text-center text-xs">
          <div className="p-1.5 rounded-lg bg-background border border-border">
            <div className="text-muted">Confidence</div>
            <div className="font-bold text-success">{bid.confidence}%</div>
          </div>
          <div className="p-1.5 rounded-lg bg-background border border-border">
            <div className="text-muted">Rep</div>
            <div className="font-bold text-primary">{agent.reputationScore}</div>
          </div>
          <div className="p-1.5 rounded-lg bg-background border border-border">
            <div className="text-muted">Time</div>
            <div className="font-bold text-white">{execTime}</div>
          </div>
        </div>
      </div>

      {/* Output body */}
      <div className="flex-1 p-4 bg-background/50">
        {status === 'idle' && (
          <div className="flex flex-col items-center justify-center h-32 text-muted text-sm gap-2">
            <Clock className="w-8 h-8 opacity-30" />
            <span>Waiting to start...</span>
          </div>
        )}
        {status === 'loading' && (
          <div className="flex flex-col items-center justify-center h-32 gap-3">
            <div className="relative">
              <div className="w-12 h-12 rounded-full border-2 border-primary/20 border-t-primary animate-spin" />
            </div>
            <p className="text-sm text-primary animate-pulse">Agent executing...</p>
          </div>
        )}
        {status === 'error' && (
          <div className="flex flex-col items-center justify-center h-32 gap-2 text-danger">
            <AlertCircle className="w-8 h-8" />
            <p className="text-sm">Execution failed</p>
            <p className="text-xs text-muted text-center">{result?.error}</p>
          </div>
        )}
        {status === 'done' && result && (
          <div className="space-y-3">
            <div className="text-xs text-muted uppercase tracking-widest mb-2">Output</div>
            <div className="text-sm">
              {formatAgentResult(agent.category, result.data)}
            </div>
          </div>
        )}
      </div>

      {/* Vote button */}
      {status === 'done' && !voted && (
        <div className="p-4 border-t border-border bg-card">
          <button
            onClick={() => onVote(bid)}
            className="w-full py-2.5 rounded-xl font-semibold text-sm bg-warning/10 hover:bg-warning/20 text-warning border border-warning/30 transition-all flex items-center justify-center gap-2"
          >
            <Star className="w-4 h-4" />
            Choose This Agent
          </button>
        </div>
      )}
      {voted && isWinner && (
        <div className="p-4 border-t border-warning/30 bg-warning/5 text-center text-sm font-bold text-warning">
          🏆 You chose this agent!
        </div>
      )}
      {voted && !isWinner && (
        <div className="p-4 border-t border-border text-center text-sm text-muted">
          Not selected
        </div>
      )}
    </motion.div>
  );
}

// ─── Main ArenaCompare ────────────────────────────────────────────────────────
export default function ArenaCompare({ bids, task, agentFormData, onWinnerSelected, onReset, onBack }) {
  // Pick 2 bids
  const [selected, setSelected] = useState([]);
  const [phase, setPhase] = useState('pick'); // 'pick' | 'running' | 'done'

  // Results per bid id
  const [results, setResults] = useState({});     // { bidId: result }
  const [statuses, setStatuses] = useState({});   // { bidId: 'idle'|'loading'|'done'|'error' }
  const [winner, setWinner] = useState(null);     // winning bid

  const handleSelect = (bid) => {
    setSelected(prev => {
      const exists = prev.find(b => b.agent.id === bid.agent.id);
      if (exists) return prev.filter(b => b.agent.id !== bid.agent.id);
      if (prev.length >= 2) return prev;
      return [...prev, bid];
    });
  };

  const isSelected = (bid) => selected.some(b => b.agent.id === bid.agent.id);

  const handleRunCompare = useCallback(async () => {
    if (selected.length !== 2) return;
    setPhase('running');

    // Set both to loading
    const initStatuses = {};
    selected.forEach(b => { initStatuses[b.agent.id] = 'loading'; });
    setStatuses(initStatuses);

    // Run both agents in parallel (Free Simulation Mode)
    await Promise.all(selected.map(async (bid) => {
      const agent = bid.agent;
      const startTime = Date.now();

      // Simulate execution time delay (1.5s to 3.5s)
      const mockDelay = 1500 + Math.random() * 2000;
      await new Promise(resolve => setTimeout(resolve, mockDelay));

      const fallbackExecTime = Date.now() - startTime;
      let mockData = {};
      
      switch (agent.category) {
        case 'defi':
          mockData = {
            trade: { from: "HELA", to: "USDC", executionPrice: "1.05", slippage: "0.1%", txHash: "0xdef1...abcd" }
          };
          break;
        case 'content':
          mockData = { platform: "Output", content: `${agent.name} produced a highly engaging thread on "${task}". The content is optimized for your target audience.` };
          break;
        case 'analysis':
          mockData = { risk_score: 15, audit_report: `Analyzed recent narrative shifts for "${task}". Data indicates a strong upcoming rotation. No major vulnerabilities detected.` };
          break;
        case 'business':
          mockData = { summary: `${agent.name} evaluated the business logic for "${task}".\n\nIdentified 3 key bottleneck areas.\nRecommended immediate action on optimization.` };
          break;
        case 'finance':
          mockData = { summary: { startValue: "$10,000", endValue: "$11,250", pnl: "+$1,250", pnlPercent: "+12.5%" }, riskScore: 45, riskLevel: "Moderate" };
          break;
        default:
          mockData = { summary: `${agent.name} completed the task "${task}" successfully.` };
      }
      
      setResults(prev => ({
        ...prev,
        [agent.id]: {
          _execTime: fallbackExecTime,
          taskId: `0xCMP_${Date.now().toString(16)}`,
          status: 'success',
          data: mockData
        }
      }));
      
      setStatuses(prev => ({ ...prev, [agent.id]: 'done' }));
    }));

    setPhase('done');
  }, [selected, task, agentFormData]);

  const handleVote = (bid) => {
    if (onWinnerSelected) onWinnerSelected(bid);
  };

  const bothDone = selected.length === 2 && selected.every(b => statuses[b.agent.id] === 'done');

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      className="space-y-6"
    >
      {/* Header */}
      <div className="text-center">
        <div className="w-16 h-16 rounded-2xl bg-primary/10 flex items-center justify-center mx-auto mb-4 border border-primary/20">
          <GitCompare className="w-8 h-8 text-primary" />
        </div>
        <h2 className="text-2xl font-bold text-white mb-2">Head-to-Head Compare</h2>
        <p className="text-muted text-sm">
          {phase === 'pick'
            ? 'Pick exactly 2 agents from your bids to run side-by-side'
            : phase === 'running'
            ? 'Both agents are executing simultaneously...'
            : winner
            ? `You chose ${winner.agent.name}! 🏆`
            : 'Both results are in — pick your winner!'}
        </p>
      </div>

      {/* Phase: pick */}
      {phase === 'pick' && (
        <>
          <div className="grid sm:grid-cols-2 md:grid-cols-3 gap-3">
            {bids.map((bid) => (
              <ComparePanel
                key={bid.agent.id}
                bid={bid}
                selected={isSelected(bid)}
                onClick={handleSelect}
                disabled={selected.length >= 2 && !isSelected(bid)}
              />
            ))}
          </div>

          <div className="flex items-center justify-between">
            <button
              onClick={onBack}
              className="px-5 py-2.5 rounded-xl border border-border text-muted hover:text-white text-sm transition-colors"
            >
              ← Back to Results
            </button>

            <div className="flex items-center gap-3">
              <span className="text-sm text-muted">
                {selected.length}/2 selected
              </span>
              <button
                onClick={handleRunCompare}
                disabled={selected.length !== 2}
                className={`flex items-center gap-2 px-6 py-2.5 rounded-xl font-semibold text-sm transition-all ${
                  selected.length === 2
                    ? 'glow-btn text-white'
                    : 'bg-border text-muted cursor-not-allowed opacity-50'
                }`}
              >
                <TrendingUp className="w-4 h-4" />
                Run Comparison
              </button>
            </div>
          </div>
        </>
      )}

      {/* Phase: running / done — side-by-side results */}
      {(phase === 'running' || phase === 'done') && (
        <>
          <div className="flex gap-4 flex-col md:flex-row">
            {selected.map((bid) => (
              <ResultColumn
                key={bid.agent.id}
                bid={bid}
                result={results[bid.agent.id]}
                status={statuses[bid.agent.id] || 'idle'}
                isWinner={winner?.agent.id === bid.agent.id}
                onVote={handleVote}
                voted={!!winner}
              />
            ))}
          </div>

          {winner && (
            <motion.div
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              className="text-center"
            >
              <button
                onClick={onReset}
                className="glow-btn text-white px-8 py-3 rounded-xl font-semibold"
              >
                Open New Arena
              </button>
            </motion.div>
          )}

          {!winner && bothDone && (
            <div className="text-center text-sm text-muted">
              ↑ Read both outputs above and click <strong className="text-white">"Choose This Agent"</strong> to declare your winner
            </div>
          )}
        </>
      )}
    </motion.div>
  );
}
