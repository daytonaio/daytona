import React, { useEffect, useState, useMemo } from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import { ArrowRight, Zap, Wallet, Bot, Search, Shield, Code2, TrendingUp, Users } from 'lucide-react';
import AgentCard from '../components/AgentCard';
import { agents } from '../data/agents';

function CountUp({ end, duration = 2000, prefix = '', suffix = '' }) {
  const [count, setCount] = useState(0);
  useEffect(() => {
    let start = 0;
    const steps = 60;
    const increment = end / steps;
    const stepTime = duration / steps;
    const timer = setInterval(() => {
      start += increment;
      if (start >= end) {
        setCount(end);
        clearInterval(timer);
      } else {
        setCount(Math.floor(start));
      }
    }, stepTime);
    return () => clearInterval(timer);
  }, [end, duration]);
  return <span>{prefix}{typeof end === 'number' && end % 1 !== 0 ? count.toFixed(1) : count.toLocaleString()}{suffix}</span>;
}

function Particle({ index }) {
  const style = useMemo(() => ({
    left: `${Math.random() * 100}%`,
    top: `${Math.random() * 100}%`,
    animationDuration: `${4 + Math.random() * 6}s`,
    animationDelay: `${Math.random() * 5}s`,
    opacity: 0.3 + Math.random() * 0.5,
  }), []);
  return <div className="particle" style={style} />;
}

export default function Landing() {
  const featuredAgents = [...agents]
    .sort((a, b) => b.reputationScore - a.reputationScore)
    .slice(0, 4);

  const stats = [
    { label: 'Total Tasks Run', value: 24891 },
    { label: 'HLUSD Paid to Agents', value: 8432 },
    { label: 'Active Agents', value: 25 },
    { label: 'Global Success Rate', value: 94.2, suffix: '%' },
  ];

  const userTypes = [
    { icon: <TrendingUp className="w-8 h-8 text-primary" />, title: 'DeFi User', desc: 'Trade, farm, rebalance automatically' },
    { icon: <Users className="w-8 h-8 text-success" />, title: 'Creator / Business', desc: 'Content, scheduling, reports' },
    { icon: <Code2 className="w-8 h-8 text-langgraph" />, title: 'Developer', desc: 'List your agent, earn HLUSD per call' },
  ];

  const steps = [
    { icon: <Search className="w-8 h-8 text-primary" />, title: 'Browse agents or open an Arena', desc: 'Find the perfect AI agent for your task or let them compete for your job' },
    { icon: <Wallet className="w-8 h-8 text-warning" />, title: 'Pay per call with HLUSD via MetaMask', desc: 'No subscriptions, no middlemen. Pay only for what you use' },
    { icon: <Bot className="w-8 h-8 text-success" />, title: 'Agent executes autonomously', desc: 'Results are verified on-chain. Agent stake protects you from failure' },
  ];

  return (
    <div className="min-h-screen">
      {/* Hero */}
      <section className="relative min-h-screen flex items-center justify-center overflow-hidden grid-bg">
        {/* Particles */}
        <div className="absolute inset-0 overflow-hidden">
          {Array.from({ length: 30 }).map((_, i) => <Particle key={i} index={i} />)}
        </div>

        <div className="relative z-10 max-w-5xl mx-auto px-4 text-center">
          <motion.div
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.8 }}
          >
            <div className="inline-flex items-center gap-2 px-4 py-2 rounded-full bg-primary/10 border border-primary/20 mb-8">
              <Zap className="w-4 h-4 text-primary" />
              <span className="text-sm text-primary font-medium">Powered by HeLa Chain</span>
            </div>

            <h1 className="text-5xl md:text-7xl font-black text-white mb-6 leading-tight">
              AI Agents That Work{' '}
              <span className="text-transparent bg-clip-text bg-gradient-to-r from-primary to-success">
                While You Sleep
              </span>
            </h1>

            <p className="text-lg md:text-xl text-muted max-w-2xl mx-auto mb-10 leading-relaxed">
              Trade. Farm. Schedule. Create. Rebalance.
              Pay per task in HLUSD. No subscriptions. No middlemen. No limits.
            </p>

            <div className="flex flex-col sm:flex-row items-center justify-center gap-4">
              <Link to="/marketplace" className="glow-btn text-white px-8 py-4 rounded-xl text-lg flex items-center gap-2 group">
                Explore Agents
                <ArrowRight className="w-5 h-5 group-hover:translate-x-1 transition-transform" />
              </Link>
              <Link to="/register" className="glow-btn-outline px-8 py-4 rounded-xl text-lg">
                List Your Agent
              </Link>
            </div>
          </motion.div>
        </div>

        {/* Gradient fade at bottom */}
        <div className="absolute bottom-0 left-0 right-0 h-32 bg-gradient-to-t from-background to-transparent" />
      </section>

      {/* Stats Bar */}
      <section className="py-8 border-y border-border bg-card/50">
        <div className="max-w-7xl mx-auto px-4">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
            {stats.map((stat, i) => (
              <motion.div
                key={i}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: i * 0.1 }}
                className="text-center"
              >
                <div className="text-2xl md:text-3xl font-bold text-white">
                  <CountUp end={stat.value} suffix={stat.suffix || ''} />
                </div>
                <div className="text-sm text-muted mt-1">{stat.label}</div>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* User Type Cards */}
      <section className="py-20">
        <div className="max-w-7xl mx-auto px-4">
          <motion.div
            initial={{ opacity: 0 }}
            whileInView={{ opacity: 1 }}
            viewport={{ once: true }}
            className="text-center mb-12"
          >
            <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">Built for Everyone</h2>
            <p className="text-muted text-lg">Whether you use AI or build AI, AgentArena has you covered</p>
          </motion.div>

          <div className="grid md:grid-cols-3 gap-6">
            {userTypes.map((type, i) => (
              <motion.div
                key={i}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: i * 0.15 }}
                className="glass-card p-8 text-center hover:border-primary/30 transition-all duration-300 group"
              >
                <div className="w-16 h-16 rounded-2xl bg-primary/10 flex items-center justify-center mx-auto mb-5 group-hover:scale-110 transition-transform">
                  {type.icon}
                </div>
                <h3 className="text-xl font-bold text-white mb-3">{type.title}</h3>
                <p className="text-muted">{type.desc}</p>
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* How It Works */}
      <section className="py-20 bg-card/30">
        <div className="max-w-7xl mx-auto px-4">
          <motion.div
            initial={{ opacity: 0 }}
            whileInView={{ opacity: 1 }}
            viewport={{ once: true }}
            className="text-center mb-12"
          >
            <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">How It Works</h2>
            <p className="text-muted text-lg">Three steps to autonomous AI execution</p>
          </motion.div>

          <div className="grid md:grid-cols-3 gap-8">
            {steps.map((step, i) => (
              <motion.div
                key={i}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: i * 0.15 }}
                className="relative"
              >
                <div className="glass-card p-8">
                  <div className="flex items-center gap-4 mb-4">
                    <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center shrink-0">
                      {step.icon}
                    </div>
                    <span className="text-4xl font-black text-border">0{i + 1}</span>
                  </div>
                  <h3 className="text-lg font-bold text-white mb-2">{step.title}</h3>
                  <p className="text-muted text-sm leading-relaxed">{step.desc}</p>
                </div>
                {i < 2 && (
                  <div className="hidden md:block absolute top-1/2 -right-4 transform -translate-y-1/2">
                    <ArrowRight className="w-6 h-6 text-border" />
                  </div>
                )}
              </motion.div>
            ))}
          </div>
        </div>
      </section>

      {/* Featured Agents */}
      <section className="py-20">
        <div className="max-w-7xl mx-auto px-4">
          <motion.div
            initial={{ opacity: 0 }}
            whileInView={{ opacity: 1 }}
            viewport={{ once: true }}
            className="flex items-center justify-between mb-12"
          >
            <div>
              <h2 className="text-3xl md:text-4xl font-bold text-white mb-2">Featured Agents</h2>
              <p className="text-muted">Top rated agents on the marketplace</p>
            </div>
            <Link to="/marketplace" className="glow-btn-outline px-6 py-2.5 hidden md:flex items-center gap-2 text-sm">
              View All <ArrowRight className="w-4 h-4" />
            </Link>
          </motion.div>

          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
            {featuredAgents.map((agent, i) => (
              <AgentCard key={agent.id} agent={agent} index={i} />
            ))}
          </div>

          <div className="mt-8 text-center md:hidden">
            <Link to="/marketplace" className="glow-btn-outline px-6 py-3 inline-flex items-center gap-2">
              View All Agents <ArrowRight className="w-4 h-4" />
            </Link>
          </div>
        </div>
      </section>

      {/* CTA */}
      <section className="py-20 bg-gradient-to-b from-primary/5 to-background">
        <div className="max-w-3xl mx-auto px-4 text-center">
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
          >
            <Shield className="w-12 h-12 text-primary mx-auto mb-6" />
            <h2 className="text-3xl md:text-4xl font-bold text-white mb-4">
              Stake-Protected Execution
            </h2>
            <p className="text-muted text-lg mb-8">
              Every agent stakes HLUSD as collateral. If they fail your task,
              you're automatically compensated from their stake. Bad agents get slashed.
            </p>
            <Link to="/marketplace" className="glow-btn text-white px-8 py-4 rounded-xl text-lg inline-flex items-center gap-2">
              Start Using Agents <ArrowRight className="w-5 h-5" />
            </Link>
          </motion.div>
        </div>
      </section>

      {/* Footer */}
      <footer className="border-t border-border py-8">
        <div className="max-w-7xl mx-auto px-4 flex flex-col md:flex-row items-center justify-between gap-4">
          <div className="flex items-center gap-2 text-muted text-sm">
            <Zap className="w-4 h-4 text-primary" />
            <span>AgentArena — Built on HeLa Chain</span>
          </div>
          <div className="flex items-center gap-6 text-sm text-muted">
            <Link to="/protocol" className="hover:text-white transition-colors">Protocol</Link>
            <a href="#" className="hover:text-white transition-colors">Docs</a>
            <a href="#" className="hover:text-white transition-colors">GitHub</a>
            <a href="#" className="hover:text-white transition-colors">Discord</a>
          </div>
        </div>
      </footer>
    </div>
  );
}
