import React, { useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { Zap, Menu, X, Swords, LayoutDashboard, Store, BarChart3, Timer } from 'lucide-react';
import { motion, AnimatePresence } from 'framer-motion';
import { useWallet } from '../hooks/useWallet';


export default function Navbar() {
  const [mobileOpen, setMobileOpen] = useState(false);
  const location = useLocation();

  const { isConnected, connect, disconnect, shortAddress, balance } = useWallet();

  const navLinks = [
    { to: '/marketplace', label: 'Marketplace', icon: <Store className="w-4 h-4" /> },
    { to: '/arena', label: '⚔️ Arena', icon: <Swords className="w-4 h-4" /> },
    { to: '/dashboard', label: 'Dashboard', icon: <LayoutDashboard className="w-4 h-4" /> },
    { to: '/protocol', label: 'Protocol', icon: <BarChart3 className="w-4 h-4" /> },
    { to: '/scheduler', label: '⏱ Scheduler', icon: <Timer className="w-4 h-4" /> },
  ];

  const isActive = (path) => location.pathname === path;

  return (
    <nav className="fixed top-0 left-0 right-0 z-50 bg-background/80 backdrop-blur-xl border-b border-border">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex items-center justify-between h-16">
          {/* Logo */}
          <Link to="/" className="flex items-center gap-2 group shrink-0">
            <Zap className="w-6 h-6 text-primary group-hover:drop-shadow-[0_0_8px_rgba(0,212,255,0.6)] transition-all" />
            <span className="text-xl font-bold text-primary tracking-tight hidden sm:block">AgentArena</span>
          </Link>

          {/* Desktop Nav */}
          <div className="hidden lg:flex items-center gap-1 mx-2">
            {navLinks.map(link => (
              <Link
                key={link.to}
                to={link.to}
                className={`px-3 py-2 rounded-lg text-sm font-medium transition-all duration-200 whitespace-nowrap ${
                  isActive(link.to)
                    ? 'bg-primary/10 text-primary'
                    : 'text-muted hover:text-white hover:bg-white/5'
                }`}
              >
                {link.label}
              </Link>
            ))}
          </div>

          {/* Wallet + Actions */}
          <div className="hidden md:flex items-center gap-3 shrink-0">


            <button
              onClick={() => isConnected ? disconnect() : connect()}
              className={`px-4 py-2 rounded-lg text-sm font-semibold transition-all duration-300 flex items-center gap-2 ${
                isConnected
                  ? 'bg-card border border-border text-white hover:border-primary/50'
                  : 'glow-btn text-white'
              }`}
            >
              {isConnected ? (
                <>
                  <span className="w-2 h-2 rounded-full bg-success" />
                  <span className="truncate max-w-[100px]">{shortAddress}</span>
                  <span className="text-muted">|</span>
                  <span className="text-primary truncate max-w-[80px]">
                    {balance} HLUSD
                  </span>
                </>
              ) : (
                'Connect Wallet'
              )}
            </button>
          </div>

          {/* Mobile hamburger */}
          <div className="md:hidden flex items-center gap-2">
            {!isConnected && (
               <button onClick={connect} className="glow-btn text-white text-xs px-3 py-1">Connect</button>
            )}
            <button
              className="text-muted hover:text-white p-2"
              onClick={() => setMobileOpen(!mobileOpen)}
            >
              {mobileOpen ? <X className="w-6 h-6" /> : <Menu className="w-6 h-6" />}
            </button>
          </div>
        </div>
      </div>

      {/* Mobile Menu */}
      <AnimatePresence>
        {mobileOpen && (
          <motion.div
            initial={{ opacity: 0, height: 0 }}
            animate={{ opacity: 1, height: 'auto' }}
            exit={{ opacity: 0, height: 0 }}
            className="md:hidden bg-card border-b border-border overflow-hidden shadow-2xl"
          >
            <div className="px-4 py-4 space-y-3">
              {isConnected && (
                 <div className="flex items-center justify-between p-3 bg-background rounded-lg border border-border">
                    <div className="flex items-center gap-2">
                      <span className="w-2 h-2 rounded-full bg-success" />
                      <span className="text-sm">{shortAddress}</span>
                    </div>
                    <div className="text-sm font-bold text-primary">{balance} HLUSD</div>
                 </div>
              )}



              {navLinks.map(link => (
                <Link
                  key={link.to}
                  to={link.to}
                  onClick={() => setMobileOpen(false)}
                  className={`flex items-center gap-3 px-4 py-3 rounded-lg text-sm font-medium transition-all ${
                    isActive(link.to)
                      ? 'bg-primary/10 text-primary'
                      : 'text-muted hover:text-white hover:bg-white/5'
                  }`}
                >
                  {link.icon}
                  {link.label}
                </Link>
              ))}
              <div className="pt-2">
                <button
                  onClick={() => {
                    isConnected ? disconnect() : connect();
                    setMobileOpen(false);
                  }}
                  className={`w-full py-3 rounded-lg text-sm font-semibold transition-all flex justify-center items-center gap-2 ${
                    isConnected ? 'bg-border text-white hover:bg-white/10' : 'glow-btn text-white'
                  }`}
                >
                  {isConnected ? 'Disconnect' : 'Connect Wallet'}
                </button>
              </div>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </nav>
  );
}
