import React from 'react';
import { BrowserRouter as Router, Routes, Route, useLocation } from 'react-router-dom';
import { AnimatePresence, motion } from 'framer-motion';
import Navbar from './components/Navbar';
import { ToastProvider } from './components/Toast';
import Landing from './pages/Landing';
import Marketplace from './pages/Marketplace';
import AgentDetail from './pages/AgentDetail';
import Arena from './pages/Arena';
import Dashboard from './pages/Dashboard';
import Register from './pages/Register';
import ManageAgent from './pages/ManageAgent';
import Protocol from './pages/Protocol';
import Scheduler from './pages/Scheduler';

import { WagmiProvider } from 'wagmi';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { wagmiConfig } from './config/wagmi';

const queryClient = new QueryClient();

function AnimatedRoutes() {
  const location = useLocation();
  return (
    <AnimatePresence mode="wait">
      <motion.div
        key={location.pathname}
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -8 }}
        transition={{ duration: 0.2 }}
      >
        <Routes location={location}>
          <Route path="/" element={<Landing />} />
          <Route path="/marketplace" element={<Marketplace />} />
          <Route path="/agent/:id" element={<AgentDetail />} />
          <Route path="/arena" element={<Arena />} />
          <Route path="/dashboard" element={<Dashboard />} />
          <Route path="/register" element={<Register />} />
          <Route path="/manage/:id" element={<ManageAgent />} />
          <Route path="/protocol" element={<Protocol />} />
          <Route path="/scheduler" element={<Scheduler />} />
        </Routes>
      </motion.div>
    </AnimatePresence>
  );
}

export default function App() {
  return (
    <WagmiProvider config={wagmiConfig}>
    <QueryClientProvider client={queryClient}>
      <Router>
        <ToastProvider>
          <div className="min-h-screen bg-background text-white font-inter">
            <Navbar />
            <AnimatedRoutes />
          </div>
        </ToastProvider>
      </Router>
    </QueryClientProvider>
    </WagmiProvider>
  );
}
