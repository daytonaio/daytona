import React, { useState, useMemo, useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Search, SlidersHorizontal, Activity } from 'lucide-react';
import { useReadContract } from 'wagmi';
import AgentCard from '../components/AgentCard';
import { agents as mockAgents, categories } from '../data/agents';
import { CONTRACT_ADDRESSES, ABIS } from '../contracts/addresses';

export default function Marketplace() {
  const [search, setSearch] = useState('');
  const [activeCategory, setActiveCategory] = useState('all');
  const [sortBy, setSortBy] = useState('topRated');
  const [isFiltering, setIsFiltering] = useState(false);

  // Fetch Live Agents from HeLa Testnet
  const { data: onChainAgents, isLoading: isLoadingChain } = useReadContract({
    address: CONTRACT_ADDRESSES.registry,
    abi: ABIS.AgentRegistry,
    functionName: 'getAllAgents',

  });

  // Calculate merged agents (Chain Data + Mock UI fluff like images)
  const allAgents = useMemo(() => {
    let baseList = [...mockAgents];

    if (onChainAgents && Array.isArray(onChainAgents)) {
      // Create a map of onchain agents by ID
      const onChainMap = new Map();
      onChainAgents.forEach(ca => {
         const id = Number(ca.id);
         if (id > 0) onChainMap.set(id, ca);
      });

      // Update baseList with onChain data, and add completely new onChain agents if they don't exist
      baseList = baseList.map(mock => {
         const ca = onChainMap.get(mock.id);
         if (ca) {
           onChainMap.delete(mock.id);
           return {
             ...mock,
             name: ca.name,
             category: ca.category,
             framework: ca.framework,
             pricePerCall: (Number(ca.pricePerCall) / 1e18).toString(),
             stakedAmount: (Number(ca.stakedAmount) / 1e18).toString(),
             description: ca.description,
             isActive: ca.isActive,
             totalTasksCompleted: Number(ca.totalTasks),
             successCount: Number(ca.successfulTasks),
           };
         }
         return mock;
      });

      // Add newly registered agents from the chain that aren't in mocks
      onChainMap.forEach(ca => {
         baseList.push({
             id: Number(ca.id),
             name: ca.name,
             category: ca.category.toLowerCase(),
             framework: ca.framework,
             pricePerCall: (Number(ca.pricePerCall) / 1e18).toString(),
             stakedAmount: (Number(ca.stakedAmount) / 1e18).toString(),
             description: ca.description,
             isActive: ca.isActive,
             totalTasksCompleted: Number(ca.totalTasks),
             successCount: Number(ca.successfulTasks),
             reputationScore: 100, // Default for new
             imageUrl: "https://api.dicebear.com/7.x/bottts/svg?seed=" + ca.name,
         });
      });
    }
    return baseList;
  }, [onChainAgents]);

  // Simulate slight loading when filters change
  useEffect(() => {
    setIsFiltering(true);
    const t = setTimeout(() => setIsFiltering(false), 300);
    return () => clearTimeout(t);
  }, [activeCategory, sortBy]);

  const filtered = useMemo(() => {
    let result = [...allAgents];

    // Search
    if (search) {
      result = result.filter(a =>
        a.name.toLowerCase().includes(search.toLowerCase()) ||
        a.description.toLowerCase().includes(search.toLowerCase())
      );
    }

    // Category
    if (activeCategory !== 'all') {
      result = result.filter(a => a.category.toLowerCase() === activeCategory);
    }

    // Sort
    switch (sortBy) {
      case 'topRated':
        result.sort((a, b) => b.reputationScore - a.reputationScore);
        break;
      case 'cheapest':
        result.sort((a, b) => parseFloat(a.pricePerCall) - parseFloat(b.pricePerCall));
        break;
      case 'mostUsed':
        result.sort((a, b) => b.totalTasksCompleted - a.totalTasksCompleted);
        break;
    }

    return result;
  }, [search, activeCategory, sortBy, allAgents]);

  const showLoader = isFiltering;

  return (
    <div className="min-h-screen pt-24 pb-12">
      <div className="max-w-7xl mx-auto px-4">
        {/* Header */}
        <motion.div
          initial={{ opacity: 0, y: -10 }}
          animate={{ opacity: 1, y: 0 }}
          className="mb-8"
        >
          <div className="flex items-center gap-3 mb-2">
            <h1 className="text-3xl md:text-4xl font-bold text-white">Agent Marketplace</h1>
            {onChainAgents && (
               <span className="flex items-center gap-1.5 px-3 py-1 rounded-full bg-success/10 border border-success/30 text-success text-xs font-bold tracking-widest uppercase">
                 <Activity size={12} className="animate-pulse" /> Live
               </span>
            )}
          </div>
          <p className="text-muted">Discover and deploy AI agents that execute autonomously</p>
        </motion.div>

        {/* Search & Filters */}
        <div className="space-y-4 mb-8">
          <div className="relative">
            <Search className="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-muted" />
            <input
              type="text"
              placeholder="Search agents by name or description..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="w-full pl-12 pr-4 py-3.5 bg-card border border-border rounded-xl text-white placeholder-muted focus:outline-none focus:border-primary transition-colors focus:ring-1 focus:ring-primary/50"
            />
          </div>

          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4">
            <div className="flex flex-wrap gap-2">
              {categories.map(cat => (
                <button
                  key={cat.key}
                  onClick={() => setActiveCategory(cat.key)}
                  className={`px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 ${
                    activeCategory === cat.key
                      ? 'bg-primary/20 text-primary border border-primary/30'
                      : 'bg-card border border-border text-muted hover:text-white hover:border-muted/50'
                  }`}
                >
                  {cat.label}
                </button>
              ))}
            </div>

            <div className="flex items-center gap-2">
              <SlidersHorizontal className="w-4 h-4 text-muted" />
              <select
                value={sortBy}
                onChange={(e) => setSortBy(e.target.value)}
                className="bg-card border border-border rounded-lg px-3 py-2 text-sm text-white focus:outline-none focus:border-primary cursor-pointer hover:border-muted/50 transition-colors appearance-none"
              >
                <option value="topRated">Top Rated</option>
                <option value="cheapest">Cheapest</option>
                <option value="mostUsed">Most Used</option>
              </select>
            </div>
          </div>
        </div>

        <div className="text-sm text-muted mb-6">
          Showing {filtered.length} agent{filtered.length !== 1 ? 's' : ''}
        </div>

        {/* Agent Grid */}
        {showLoader ? (
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6 animate-pulse">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="glass-card p-5">
                <div className="flex justify-between mb-3">
                  <div className="space-y-2">
                    <div className="bg-white/5 h-5 w-20 rounded" />
                    <div className="bg-white/5 h-6 w-32 rounded" />
                  </div>
                  <div className="bg-white/5 h-14 w-14 rounded-full" />
                </div>
                <div className="bg-white/5 h-4 w-full rounded mb-2" />
                <div className="bg-white/5 h-4 w-3/4 rounded mb-4" />
                <div className="grid grid-cols-3 gap-3 mb-4">
                  {[1, 2, 3].map(j => <div key={j} className="bg-white/5 h-16 rounded-lg" />)}
                </div>
                <div className="flex gap-2">
                  <div className="bg-white/5 h-10 flex-1 rounded-lg" />
                  <div className="bg-white/5 h-10 w-20 rounded-lg" />
                </div>
              </div>
            ))}
          </div>
        ) : filtered.length === 0 ? (
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            className="text-center py-20"
          >
            <Search className="w-12 h-12 text-muted mx-auto mb-4" />
            <h3 className="text-xl font-bold text-white mb-2">No agents found</h3>
            <p className="text-muted">Try adjusting your search or filter criteria</p>
          </motion.div>
        ) : (
          <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
            <AnimatePresence mode="popLayout">
              {filtered.map((agent, i) => (
                <AgentCard key={agent.id} agent={agent} index={i} />
              ))}
            </AnimatePresence>
          </div>
        )}
      </div>
    </div>
  );
}
