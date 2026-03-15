import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { X, CheckCircle, ExternalLink } from 'lucide-react';
import { formatAgentResult } from '../utils/resultFormatter';

export default function ResultModal({ isOpen, onClose, result, agent }) {
  if (!isOpen || !result) return null;

  return (
    <AnimatePresence>
      <div className="fixed inset-0 z-50 flex items-center justify-center px-4">
        <div className="absolute inset-0 bg-black/60 backdrop-blur-sm" onClick={onClose} />
        <motion.div
          initial={{ scale: 0.95, opacity: 0, y: 10 }}
          animate={{ scale: 1, opacity: 1, y: 0 }}
          exit={{ scale: 0.95, opacity: 0, y: 10 }}
          className="relative bg-card border border-border rounded-2xl p-6 md:p-8 max-w-2xl w-full shadow-2xl max-h-[90vh] overflow-y-auto"
        >
          <button onClick={onClose} className="absolute top-4 right-4 text-muted hover:text-white transition-colors">
            <X className="w-6 h-6" />
          </button>

          <div className="flex flex-col items-center mb-6">
            <div className="w-16 h-16 rounded-full bg-success/20 flex items-center justify-center mb-4">
              <CheckCircle className="w-10 h-10 text-success" />
            </div>
            <h3 className="text-2xl font-bold text-white mb-1">Task Completed</h3>
            <p className="text-sm text-muted">Agent execution finished successfully</p>
          </div>

          {/* Details Row */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 mb-8">
             <div className="bg-background rounded-lg p-3 border border-border">
                <div className="text-[10px] uppercase tracking-wider text-muted mb-1">Agent</div>
                <div className="text-sm font-semibold truncate text-white">{result.agent || agent?.name}</div>
             </div>
             <div className="bg-background rounded-lg p-3 border border-border">
                <div className="text-[10px] uppercase tracking-wider text-muted mb-1">Task ID</div>
                <div className="text-sm font-mono text-primary truncate" title={result.taskId}>{result.taskId}</div>
             </div>
             <div className="bg-background rounded-lg p-3 border border-border">
                <div className="text-[10px] uppercase tracking-wider text-muted mb-1">Duration</div>
                <div className="text-sm font-semibold text-white">{result.duration || '0.0'}s</div>
             </div>
             <div className="bg-background rounded-lg p-3 border border-border">
                <div className="text-[10px] uppercase tracking-wider text-muted mb-1">On-Chain Tx</div>
                {result.txHash && result.txHash !== '0xDEMO_MODE' ? (
                  <a href={`https://testnet-scan.helachain.com/tx/${result.txHash}`} target="_blank" rel="noreferrer" className="text-sm font-mono text-[#00FF94] hover:underline flex items-center gap-1 truncate" title={result.txHash}>
                    {result.txHash.slice(0,8)}... <ExternalLink size={12}/>
                  </a>
                ) : (
                  <div className="text-sm text-muted italic">Simulated</div>
                )}
             </div>
          </div>

          {/* Result Content Formatted */}
          <div className="mb-8">
            <h4 className="text-sm font-bold text-white mb-3 flex items-center gap-2">
              <span className="w-1.5 h-1.5 rounded-full bg-primary" />
              Execution Output
            </h4>
            {formatAgentResult(agent?.category, result.data)}
          </div>

          <div className="flex justify-end gap-3 border-t border-border pt-6">
             <button onClick={onClose} className="px-6 py-2.5 rounded-xl font-semibold text-sm bg-background border border-border text-muted hover:text-white transition-colors">
               Close
             </button>
             <button onClick={onClose} className="px-6 py-2.5 rounded-xl font-semibold text-sm bg-primary/20 text-primary border border-primary/30 hover:bg-primary/30 transition-colors">
               Confirm & Submit Rating
             </button>
          </div>
        </motion.div>
      </div>
    </AnimatePresence>
  );
}
