import React from 'react';
import { Link } from 'react-router-dom';
import { Zap, CheckCircle, Lock, Swords } from 'lucide-react';
import FrameworkBadge from './FrameworkBadge';
import ReputationGauge from './ReputationGauge';
import AgentAvatar from './AgentAvatar';
import { categoryColors } from '../data/agents';
import { motion } from 'framer-motion';

export default function AgentCard({ agent, index = 0 }) {
  const borderColor = categoryColors[agent.category] || '#1E1E2E';
  const successRate = ((agent.successCount / agent.totalTasksCompleted) * 100).toFixed(1);

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.3, delay: index * 0.05 }}
      className="glass-card overflow-hidden hover:border-primary/30 transition-all duration-300 group"
      style={{ borderTop: `3px solid ${borderColor}` }}
    >
      <div className="p-5">
        {/* Header */}
        <div className="flex items-start justify-between mb-3">
          <div className="flex-1">
            <div className="flex items-center gap-2 mb-1">
              <FrameworkBadge framework={agent.framework} />
              <span className="text-xs text-muted capitalize">{agent.category}</span>
            </div>
            <h3 className="text-lg font-bold text-white group-hover:text-primary transition-colors">
              {agent.name}
            </h3>
          </div>
          <div className="flex flex-col items-center gap-1.5">
            <AgentAvatar agent={agent} size="sm" />
            <ReputationGauge score={agent.reputationScore} size={48} />
          </div>
        </div>

        {/* Description */}
        <p className="text-sm text-muted mb-4 leading-relaxed">{agent.description}</p>

        {/* Stats */}
        <div className="grid grid-cols-3 gap-3 mb-4">
          <div className="text-center p-2 rounded-lg bg-background/50">
            <div className="flex items-center justify-center gap-1 mb-1">
              <Zap className="w-3 h-3 text-primary" />
            </div>
            <div className="text-sm font-semibold text-white">{agent.pricePerCall}</div>
            <div className="text-xs text-muted">HLUSD/call</div>
          </div>
          <div className="text-center p-2 rounded-lg bg-background/50">
            <div className="flex items-center justify-center gap-1 mb-1">
              <CheckCircle className="w-3 h-3 text-success" />
            </div>
            <div className="text-sm font-semibold text-white">{agent.totalTasksCompleted.toLocaleString()}</div>
            <div className="text-xs text-muted">Tasks</div>
          </div>
          <div className="text-center p-2 rounded-lg bg-background/50">
            <div className="flex items-center justify-center gap-1 mb-1">
              <Lock className="w-3 h-3 text-warning" />
            </div>
            <div className="text-sm font-semibold text-white">{agent.stakedAmount}</div>
            <div className="text-xs text-muted">Staked</div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex gap-2">
          <Link
            to={`/agent/${agent.id}`}
            className="flex-1 text-center py-2.5 rounded-lg bg-primary/10 text-primary font-semibold text-sm hover:bg-primary/20 transition-all duration-200"
          >
            Use Directly
          </Link>
          <Link
            to="/arena"
            className="flex items-center justify-center gap-1 px-4 py-2.5 rounded-lg border border-border text-muted hover:text-white hover:border-warning transition-all duration-200 text-sm"
          >
            <Swords className="w-4 h-4" />
            Arena
          </Link>
        </div>
      </div>
    </motion.div>
  );
}
