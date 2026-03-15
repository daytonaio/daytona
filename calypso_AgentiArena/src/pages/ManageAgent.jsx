import React from 'react';
import { useParams, Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { ArrowLeft, CheckCircle, XCircle, AlertTriangle, TrendingUp, DollarSign, Shield } from 'lucide-react';
import { LineChart, Line, BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, Cell } from 'recharts';
import { agents, getReputationColor } from '../data/agents';
import { mockReputationTrend, mockManageTaskHistory } from '../data/mockTasks';
import FrameworkBadge from '../components/FrameworkBadge';
import ReputationGauge from '../components/ReputationGauge';
import { useToast } from '../components/Toast';

const CustomTooltip = ({ active, payload, label }) => {
  if (active && payload?.length) {
    return (
      <div className="bg-card border border-border rounded-lg px-3 py-2 text-sm">
        <p className="text-white font-medium">{label}: {payload[0].value}</p>
      </div>
    );
  }
  return null;
};

export default function ManageAgent() {
  const { id } = useParams();
  const agent = agents.find(a => a.id === parseInt(id));
  const { addToast } = useToast();

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

  const earningsData = [
    { name: 'Earnings', value: (agent.totalTasksCompleted * parseFloat(agent.pricePerCall) * 0.9).toFixed(1) },
    { name: 'Stake', value: agent.stakedAmount },
  ];
  const barColors = ['#00D4FF', '#8B5CF6'];

  const disputes = mockManageTaskHistory.filter(t => t.status === 'disputed');

  return (
    <div className="min-h-screen pt-24 pb-12">
      <div className="max-w-6xl mx-auto px-4">
        <Link to="/marketplace" className="inline-flex items-center gap-2 text-muted hover:text-white mb-6 transition-colors">
          <ArrowLeft className="w-4 h-4" /> Back to Marketplace
        </Link>

        {/* Header */}
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} className="glass-card p-6 mb-6">
          <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-4">
            <div className="flex items-center gap-4">
              <ReputationGauge score={agent.reputationScore} size={80} />
              <div>
                <div className="flex items-center gap-2 mb-1">
                  <h1 className="text-2xl font-bold text-white">{agent.name}</h1>
                  <span className="flex items-center gap-1 px-2 py-0.5 rounded-full bg-success/10 text-success text-xs font-semibold">
                    <span className="w-1.5 h-1.5 rounded-full bg-success animate-pulse" /> Active
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <FrameworkBadge framework={agent.framework} />
                  <span className="text-sm text-muted capitalize">{agent.category}</span>
                </div>
              </div>
            </div>
            <div className="flex gap-2">
              <button onClick={() => addToast('Withdrawal initiated (mock)', 'success')} className="px-4 py-2 rounded-lg bg-success/10 text-success text-sm font-semibold hover:bg-success/20 transition-colors">
                Withdraw Earnings
              </button>
              <button onClick={() => addToast('Stake added (mock)', 'info')} className="px-4 py-2 rounded-lg bg-primary/10 text-primary text-sm font-semibold hover:bg-primary/20 transition-colors">
                Add Stake
              </button>
              <button onClick={() => addToast('Price updated (mock)', 'info')} className="px-4 py-2 rounded-lg border border-border text-muted text-sm font-semibold hover:text-white hover:border-primary transition-colors">
                Update Price
              </button>
            </div>
          </div>
        </motion.div>

        {/* Charts */}
        <div className="grid md:grid-cols-2 gap-6 mb-6">
          {/* Reputation Trend */}
          <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.1 }} className="glass-card p-6">
            <div className="flex items-center gap-2 mb-4">
              <TrendingUp className="w-5 h-5 text-primary" />
              <h3 className="font-semibold text-white">7-Day Reputation Trend</h3>
            </div>
            <ResponsiveContainer width="100%" height={200}>
              <LineChart data={mockReputationTrend}>
                <XAxis dataKey="day" stroke="#8888AA" fontSize={12} />
                <YAxis domain={[85, 100]} stroke="#8888AA" fontSize={12} />
                <Tooltip content={<CustomTooltip />} />
                <Line
                  type="monotone"
                  dataKey="score"
                  stroke={getReputationColor(agent.reputationScore)}
                  strokeWidth={2}
                  dot={{ fill: getReputationColor(agent.reputationScore), r: 4 }}
                  activeDot={{ r: 6 }}
                />
              </LineChart>
            </ResponsiveContainer>
          </motion.div>

          {/* Earnings vs Stake */}
          <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.2 }} className="glass-card p-6">
            <div className="flex items-center gap-2 mb-4">
              <DollarSign className="w-5 h-5 text-warning" />
              <h3 className="font-semibold text-white">Earnings vs Stake</h3>
            </div>
            <ResponsiveContainer width="100%" height={200}>
              <BarChart data={earningsData}>
                <XAxis dataKey="name" stroke="#8888AA" fontSize={12} />
                <YAxis stroke="#8888AA" fontSize={12} />
                <Tooltip content={<CustomTooltip />} />
                <Bar dataKey="value" radius={[8, 8, 0, 0]}>
                  {earningsData.map((_, i) => (
                    <Cell key={i} fill={barColors[i]} />
                  ))}
                </Bar>
              </BarChart>
            </ResponsiveContainer>
          </motion.div>
        </div>

        {/* Task History */}
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.3 }} className="glass-card mb-6">
          <div className="p-6 border-b border-border">
            <h2 className="text-lg font-semibold text-white">Recent Tasks (Last 10)</h2>
          </div>
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead>
                <tr className="border-b border-border text-sm text-muted">
                  <th className="text-left p-4 font-medium">Task ID</th>
                  <th className="text-left p-4 font-medium">User</th>
                  <th className="text-left p-4 font-medium">Status</th>
                  <th className="text-left p-4 font-medium">Cost</th>
                  <th className="text-left p-4 font-medium">Time</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {mockManageTaskHistory.map(t => (
                  <tr key={t.id} className="hover:bg-white/[0.02] transition-colors">
                    <td className="p-4 text-sm text-muted font-mono">{t.id}</td>
                    <td className="p-4 text-sm text-white">{t.user}</td>
                    <td className="p-4">
                      <span className={`inline-flex items-center gap-1 text-xs font-semibold px-2 py-1 rounded-full ${
                        t.status === 'success' ? 'bg-success/10 text-success' :
                        t.status === 'failed' ? 'bg-danger/10 text-danger' :
                        'bg-warning/10 text-warning'
                      }`}>
                        {t.status === 'success' ? <CheckCircle className="w-3 h-3" /> :
                         t.status === 'failed' ? <XCircle className="w-3 h-3" /> :
                         <AlertTriangle className="w-3 h-3" />}
                        {t.status}
                      </span>
                    </td>
                    <td className="p-4 text-sm text-primary">{t.cost} HLUSD</td>
                    <td className="p-4 text-sm text-muted">{new Date(t.timestamp).toLocaleString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </motion.div>

        {/* Open Disputes */}
        <motion.div initial={{ opacity: 0, y: 20 }} animate={{ opacity: 1, y: 0 }} transition={{ delay: 0.4 }} className="glass-card">
          <div className="p-6 border-b border-border">
            <div className="flex items-center gap-2">
              <Shield className="w-5 h-5 text-warning" />
              <h2 className="text-lg font-semibold text-white">Open Disputes</h2>
              {disputes.length > 0 && (
                <span className="px-2 py-0.5 rounded-full bg-warning/10 text-warning text-xs font-semibold">{disputes.length}</span>
              )}
            </div>
          </div>
          {disputes.length === 0 ? (
            <div className="p-8 text-center text-muted">
              <CheckCircle className="w-8 h-8 mx-auto mb-2 text-success" />
              <p>No open disputes — great job!</p>
            </div>
          ) : (
            <div className="divide-y divide-border">
              {disputes.map(d => (
                <div key={d.id} className="p-4 flex items-center justify-between">
                  <div>
                    <div className="text-sm text-white font-medium">Task {d.id}</div>
                    <div className="text-xs text-muted">User: {d.user} · {new Date(d.timestamp).toLocaleDateString()}</div>
                  </div>
                  <button className="px-3 py-1.5 rounded-lg bg-warning/10 text-warning text-xs font-semibold hover:bg-warning/20 transition-colors">
                    Respond
                  </button>
                </div>
              ))}
            </div>
          )}
        </motion.div>
      </div>
    </div>
  );
}
