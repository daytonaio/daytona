import React, { useState, useEffect, useRef } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Timer, Swords, Trophy, Loader2, GitCompare } from 'lucide-react';
import FrameworkBadge from './FrameworkBadge';
import { agents, getReputationColor } from '../data/agents';
import confetti from 'canvas-confetti';

export default function ArenaLive({ task, category, maxBudget, duration, onWinnerSelected, onCompareSelected, onReset }) {
  const [timeLeft, setTimeLeft] = useState(duration);
  const [bids, setBids] = useState([]);
  const [arenaEnded, setArenaEnded] = useState(false);
  const [winner, setWinner] = useState(null);
  const [executing, setExecuting] = useState(false);
  const bidIndexRef = useRef(0);

  // Get eligible agents
  const eligible = agents.filter(a =>
    category === 'all' ? true : a.category === category
  );

  // Timer countdown
  useEffect(() => {
    if (timeLeft <= 0) {
      setArenaEnded(true);
      return;
    }
    const t = setInterval(() => setTimeLeft(prev => prev - 1), 1000);
    return () => clearInterval(t);
  }, [timeLeft]);

  // Fetch real AI-generated dynamic bids from the Central Bid Engine (Port 8012)
  useEffect(() => {
    if (bids.length > 0 || arenaEnded) return;
    if (eligible.length === 0) return;

    const fetchBids = async () => {
      try {
        const payload = {
          task,
          category,
          maxBudget,
          available_agents: eligible.map(a => ({
            id: a.id,
            name: a.name,
            category: a.category,
            pricePerCall: a.pricePerCall,
            description: a.description
          }))
        };

        const res = await fetch('https://bid-engine-ynj0.onrender.com/api/v1/arena/bids', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });

        if (!res.ok) throw new Error('Failed to fetch AI bids');
        const data = await res.json();

        // Stagger the AI bids into the UI for the "live auction" effect
        const intervals = [800, 2000, 3500];
        
        data.bids.forEach((bidData, i) => {
          setTimeout(() => {
            setBids(prev => {
              if (prev.some(b => b.agent.id === bidData.agent_id)) return prev;
              
              const matchedAgent = eligible.find(a => a.id === bidData.agent_id);
              if (!matchedAgent) return prev; // failsafe

              return [...prev, {
                agent: matchedAgent,
                bidAmount: bidData.bid_amount.toFixed(4),
                confidence: bidData.confidence_score,
                rationale: bidData.rationale,
                timestamp: Date.now()
              }].sort((a, b) => parseFloat(a.bidAmount) - parseFloat(b.bidAmount));
            });
          }, intervals[i] || 4000);
        });
      } catch (err) {
        console.error("Bid Engine Error:", err);
      }
    };

    fetchBids();
  }, [eligible, task, maxBudget, category, arenaEnded]);

  const getTimerColor = () => {
    if (timeLeft <= 10) return 'text-danger';
    if (timeLeft <= 30) return 'text-warning';
    return 'text-success';
  };

  const formatTime = (s) => {
    const m = Math.floor(s / 60);
    const sec = s % 60;
    return `${m}:${sec.toString().padStart(2, '0')}`;
  };

  const handleSelectWinner = (bid) => {
    setWinner(bid);
    setExecuting(true);
    confetti({
      particleCount: 100,
      spread: 70,
      origin: { y: 0.6 },
      colors: ['#00D4FF', '#00FF94', '#8B5CF6'],
    });
    setTimeout(() => {
      setExecuting(false);
      onWinnerSelected?.(bid, bids);
    }, 2000);
  };

  return (
    <div className="space-y-6">
      {/* Task Card */}
      <div className="glass-card p-6">
        <div className="flex items-center gap-2 mb-2 text-sm text-muted">
          <Swords className="w-4 h-4 text-warning" />
          Active Arena
        </div>
        <p className="text-white font-medium mb-3">{task}</p>
        <div className="flex items-center gap-4 text-sm text-muted">
          <span>Category: <span className="text-white capitalize">{category}</span></span>
          <span>Max Budget: <span className="text-primary">{maxBudget} HLUSD</span></span>
        </div>
      </div>

      {/* Timer */}
      <div className="text-center">
        <div className={`text-6xl font-bold font-mono ${getTimerColor()} transition-colors`}>
          {formatTime(timeLeft)}
        </div>
        <div className="flex flex-col items-center justify-center gap-3 mt-2 min-h-[60px]">
          <p className="text-muted">
            {arenaEnded ? '⏰ Arena closed — select your winner or compare' : 'Time remaining'}
          </p>
          {arenaEnded && bids.length >= 2 && !winner && (
            <motion.button
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              onClick={() => onCompareSelected?.(bids)}
              className="flex items-center justify-center gap-2 px-6 py-2.5 rounded-xl font-semibold border-2 border-primary/40 text-primary bg-primary/10 hover:bg-primary/20 transition-all shadow-[0_0_15px_rgba(0,212,255,0.15)]"
            >
              <GitCompare className="w-5 h-5" /> Compare Agents
            </motion.button>
          )}
        </div>
      </div>

      {/* Bids */}
      <div className="space-y-3">
        {bids.length === 0 && (
          <div className="glass-card p-8 text-center text-muted">
            <Loader2 className="w-8 h-8 text-primary animate-spin mx-auto mb-3" />
            Waiting for agents to submit bids...
          </div>
        )}

        <AnimatePresence>
          {bids.map((bid, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, x: 100 }}
              animate={{ opacity: 1, x: 0 }}
              transition={{ duration: 0.4, ease: 'easeOut' }}
              className={`glass-card p-4 ${winner?.agent.id === bid.agent.id ? 'border-success ring-1 ring-success/30' : ''}`}
            >
              <div className="flex items-center justify-between flex-wrap gap-3">
                <div className="flex items-center gap-3">
                  <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center text-primary font-bold">
                    #{i + 1}
                  </div>
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="font-semibold text-white">{bid.agent.name}</span>
                      <FrameworkBadge framework={bid.agent.framework} />
                    </div>
                    <div className="flex items-center gap-3 mt-1 text-xs text-muted">
                      <span>Rep: <span style={{ color: getReputationColor(bid.agent.reputationScore) }}>{bid.agent.reputationScore}</span></span>
                      <span>Tasks: {bid.agent.totalTasksCompleted.toLocaleString()}</span>
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <div className="text-right">
                    <div className="text-lg font-bold text-primary">{bid.bidAmount} HLUSD</div>
                    <div className="flex flex-col text-[10px] text-muted text-right">
                      {bid.confidence && <span className="text-success">{bid.confidence}% Confidence</span>}
                      {bid.rationale && <span className="max-w-[150px] truncate" title={bid.rationale}>"{bid.rationale}"</span>}
                    </div>
                  </div>
                  {!winner && (
                    <button
                      onClick={() => handleSelectWinner(bid)}
                      className="px-4 py-2 rounded-lg bg-primary/10 text-primary font-semibold text-sm hover:bg-primary/20 transition-all"
                    >
                      Hire
                    </button>
                  )}
                </div>
              </div>
            </motion.div>
          ))}
        </AnimatePresence>
      </div>

      {/* Winner State */}
      {winner && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          className="glass-card p-6 border-success text-center"
        >
          {executing ? (
            <div className="flex flex-col items-center">
              <Trophy className="w-10 h-10 text-warning mb-3" />
              <h3 className="text-lg font-bold text-white mb-2">
                ✅ Winner selected! Agent executing...
              </h3>
              <Loader2 className="w-6 h-6 text-primary animate-spin" />
            </div>
          ) : (
            <div className="flex flex-col items-center">
              <Trophy className="w-10 h-10 text-warning mb-3" />
              <h3 className="text-lg font-bold text-white mb-2">Task Complete!</h3>
              <p className="text-muted text-sm mb-4">
                {winner.agent.name} executed your task for {winner.bidAmount} HLUSD
              </p>
              <button onClick={onReset} className="glow-btn text-white px-6 py-2 rounded-lg text-sm">
                Open New Arena
              </button>
            </div>
          )}
        </motion.div>
      )}
    </div>
  );
}
