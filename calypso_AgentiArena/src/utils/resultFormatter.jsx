import React from 'react';

/**
 * Beautifully formats backend results based on agent category
 */
export const formatAgentResult = (agentCategory, resultData) => {
  if (!resultData) return null;

  switch (agentCategory?.toLowerCase()) {
    case 'defi':
      return (
        <div className="space-y-4">
          {/* Universal Agent Messages */}
          {resultData.messages && resultData.messages.length > 0 && (
             <div className="bg-agent-card rounded-lg p-4 border border-agent-border">
                <h4 className="text-sm font-semibold text-agent-muted mb-2">Agent Execution Log</h4>
                <ul className="text-xs text-agent-primary space-y-1 font-mono">
                  {resultData.messages.map((m, i) => (
                    <li key={i}>&gt; {m}</li>
                  ))}
                </ul>
             </div>
          )}
          
          {/* Sniper Agent Specific */}
          {resultData.execution_payload && (
            <div className="bg-[#00FF94]/10 border border-[#00FF94]/30 rounded-lg p-4 mb-2">
                <div className="text-[#00FF94] font-bold mb-2">Arbitrage Executed Successfully!</div>
                <div className="text-sm text-white mb-1">Buy: <span className="text-[#00D4FF]">{resultData.execution_payload.buy?.exchange}</span> @ ${resultData.execution_payload.buy?.price_usd}</div>
                <div className="text-sm text-white mb-2">Sell: <span className="text-[#00D4FF]">{resultData.execution_payload.sell?.exchange}</span> @ ${resultData.execution_payload.sell?.price_usd}</div>
                
                <div className="flex justify-between items-center bg-[#0A0A0F] p-3 rounded-md mt-3 border border-agent-border text-xs">
                   <div>
                      <span className="text-agent-muted">Net Profit:</span>
                      <span className="ml-2 font-bold text-[#00FF94]">${resultData.execution_payload.net_profit_usd}</span>
                   </div>
                   <div>
                      <span className="text-agent-muted">Spread:</span>
                      <span className="ml-2 font-bold text-[#FFB800]">{resultData.execution_payload.spread_pct}%</span>
                   </div>
                </div>
            </div>
          )}

          {/* Fallback for unrecognized schemas */}
          {!resultData.trade && !resultData.selectedFarm && !resultData.newAllocation && !resultData.messages && (
            <pre className="bg-[#0A0A0F] p-4 rounded-lg overflow-x-auto text-xs text-[#00D4FF] font-mono border border-agent-border">
              {JSON.stringify(resultData, null, 2)}
            </pre>
          )}

          {resultData.trade && (
            <div className="bg-agent-card rounded-lg p-4 border border-agent-border">
              <h4 className="text-sm font-semibold text-agent-muted mb-2">Trade Execution</h4>
              <div className="grid grid-cols-2 gap-2 text-sm">
                <span className="text-white">Route:</span>
                <span className="text-agent-primary break-all">{resultData.trade.from} → {resultData.trade.to}</span>
                <span className="text-white">Price:</span>
                <span className="text-[#00FF94]">${resultData.trade.executionPrice}</span>
                <span className="text-white">Slippage:</span>
                <span className="text-[#FFB800]">{resultData.trade.slippage}</span>
                <span className="text-white">TX:</span>
                <a href="#" className="text-agent-primary underline truncate">{resultData.trade.txHash}</a>
              </div>
            </div>
          )}
          {resultData.selectedFarm && (
            <div className="bg-agent-card rounded-lg p-4 border border-agent-border">
              <h4 className="text-sm font-semibold text-agent-muted mb-2">Selected Farm</h4>
              <div className="flex justify-between items-center bg-[#0A0A0F] p-3 rounded-md">
                <div>
                  <div className="font-bold text-white">{resultData.selectedFarm.protocol}</div>
                  <div className="text-xs text-agent-muted">{resultData.selectedFarm.pool}</div>
                </div>
                <div className="text-right">
                  <div className="font-bold text-[#00FF94]">{resultData.selectedFarm.apy} APY</div>
                  <div className="text-xs text-[#FFB800] capitalize">{resultData.selectedFarm.risk} Risk</div>
                </div>
              </div>
            </div>
          )}
          {resultData.newAllocation && (
             <div className="flex flex-wrap gap-2 mt-2">
                {Object.entries(resultData.newAllocation).map(([k,v]) => (
                  <span key={k} className="px-2 py-1 bg-[#1A1A24] rounded text-xs text-[#00D4FF]">{k}: {v}%</span>
                ))}
             </div>
          )}
        </div>
      );

    case 'portfolio':
    case 'finance':
      return (
        <div className="space-y-4">
          {/* Universal Agent Messages */}
          {resultData.messages && resultData.messages.length > 0 && (
             <div className="bg-agent-card rounded-lg p-4 border border-agent-border mb-4">
                <h4 className="text-sm font-semibold text-agent-muted mb-2">Execution Log</h4>
                <ul className="text-xs text-agent-primary space-y-1 font-mono">
                  {resultData.messages.map((m, i) => (
                    <li key={i}>&gt; {m}</li>
                  ))}
                </ul>
             </div>
          )}

          {/* Tax Agent Specific */}
          {resultData.tax_report && (
            <div className="bg-agent-card rounded-lg p-5 border border-agent-border">
              <h4 className="text-sm font-semibold text-[#00D4FF] mb-3">Generated Tax Report</h4>
              <div className="whitespace-pre-wrap text-sm text-gray-200 leading-relaxed font-sans">
                {resultData.tax_report}
              </div>
              <div className="mt-4 pt-4 border-t border-agent-border text-right">
                 <button onClick={() => navigator.clipboard.writeText(resultData.tax_report)} className="text-xs text-[#00FF94] hover:underline">Copy Document</button>
              </div>
            </div>
          )}

          {resultData.summary && (
            <div className="grid grid-cols-2 gap-4">
               <div className="bg-agent-card p-4 rounded-lg border border-agent-border">
                 <div className="text-xs text-agent-muted uppercase">Start Value</div>
                 <div className="text-lg font-bold text-white">{resultData.summary.startValue}</div>
               </div>
               <div className="bg-agent-card p-4 rounded-lg border border-agent-border">
                 <div className="text-xs text-agent-muted uppercase">End Value</div>
                 <div className="text-lg font-bold text-[#00FF94]">{resultData.summary.endValue}</div>
               </div>
               <div className="col-span-2 bg-agent-card p-4 rounded-lg border border-agent-border flex flex-col items-center">
                 <div className="text-xs text-agent-muted uppercase">Total PNL</div>
                 <div className="text-3xl font-bold text-[#00FF94]">{resultData.summary.pnl} <span className="text-sm font-normal">({resultData.summary.pnlPercent})</span></div>
               </div>
            </div>
          )}
          {resultData.riskScore && (
             <div className="bg-agent-card p-4 rounded-lg border border-agent-border flex items-center justify-between">
                <div>
                   <div className="text-xs text-agent-muted uppercase">Risk Score</div>
                   <div className="text-2xl font-bold text-[#FFB800]">{resultData.riskScore}/100</div>
                </div>
                <div className="text-right">
                    <div className="text-xs text-agent-muted uppercase">Level</div>
                    <div className="text-lg font-bold text-[#FFB800]">{resultData.riskLevel}</div>
                </div>
             </div>
          )}

          {/* Fallback for unrecognized finance schemas */}
          {!resultData.summary && !resultData.riskScore && !resultData.tax_report && !resultData.messages && (
            <pre className="bg-[#0A0A0F] p-4 rounded-lg overflow-x-auto text-xs text-[#00D4FF] font-mono border border-agent-border">
              {JSON.stringify(resultData, null, 2)}
            </pre>
          )}
        </div>
      );

    case 'content':
      return (
        <div className="bg-agent-card rounded-lg p-5 border border-agent-border">
          <div className="flex justify-between items-center mb-4">
            <span className="px-3 py-1 bg-[#00D4FF]/20 text-agent-primary text-xs rounded-full uppercase font-bold tracking-wider">
              {resultData.platform || 'Generated Content'}
            </span>
            {resultData.readingTime && <span className="text-xs text-agent-muted">{resultData.readingTime}</span>}
          </div>
          <div className="whitespace-pre-wrap text-sm text-gray-200 leading-relaxed font-sans">
            {resultData.content || resultData.optimizedVersion || JSON.stringify(resultData.repurposed, null, 2)}
          </div>
          <div className="mt-4 pt-4 border-t border-agent-border flex justify-between items-center">
             <div className="flex gap-2 text-xs text-agent-primary">
                {resultData.hashtags?.map(h => <span key={h}>{h}</span>)}
             </div>
             <button onClick={() => navigator.clipboard.writeText(resultData.content)} className="text-xs text-[#00FF94] hover:underline">Copy text</button>
          </div>
        </div>
      );

    case 'onchain':
      return (
        <div className="space-y-4">
           {resultData.recommendation && resultData.recommendation.savings && (
             <div className="bg-[#00FF94]/10 border border-[#00FF94]/30 rounded-lg p-4 text-center">
                <div className="text-[#00FF94] text-lg font-bold mb-1">Optimum Action Found</div>
                <div className="text-sm text-[#00FF94]/80">{resultData.recommendation.optimalWindow}</div>
                <div className="text-xs mt-2 text-white">Expected Savings: {resultData.recommendation.savings}</div>
             </div>
           )}
           {resultData.flags && (
             <div className="bg-[#FF4444]/10 border border-[##FF4444]/30 rounded-lg p-4">
                <div className="text-[#FF4444] font-bold mb-2">Audit Flags Discovered</div>
                <ul className="list-disc pl-5 text-sm space-y-1">
                  {resultData.flags.map((f, i) => (
                    <li key={i} className="text-red-200"><strong className="text-white">{f.severity}:</strong> {f.issue}</li>
                  ))}
                </ul>
             </div>
           )}
           {resultData.sentiment && (
             <div className="bg-agent-card rounded-lg p-4 border border-agent-border text-center">
                <div className="text-xs text-agent-muted uppercase mb-1">On-Chain Sentiment</div>
                <div className="text-xl font-bold text-agent-primary">{resultData.sentiment}</div>
                <div className="text-xs mt-2 text-[#00FF94]">{resultData.netFlow || resultData.topic}</div>
             </div>
           )}
        </div>
      );

    case 'analysis':
      return (
        <div className="space-y-4">
          {/* Universal Agent Messages */}
          {resultData.messages && resultData.messages.length > 0 && (
             <div className="bg-agent-card rounded-lg p-4 border border-agent-border mb-4">
                <h4 className="text-sm font-semibold text-agent-muted mb-2">Execution Log</h4>
                <ul className="text-xs text-agent-primary space-y-1 font-mono">
                  {resultData.messages.map((m, i) => (
                    <li key={i}>&gt; {m}</li>
                  ))}
                </ul>
             </div>
          )}

          {/* Guardian / Analysis Specific */}
          {resultData.risk_score !== undefined && (
             <div className="bg-agent-card p-4 rounded-lg border border-agent-border flex items-center justify-between mb-4">
                <div>
                   <div className="text-xs text-agent-muted uppercase">Security Risk Score</div>
                   <div className={`text-2xl font-bold ${resultData.risk_score > 70 ? 'text-[#FF4444]' : resultData.risk_score > 30 ? 'text-[#FFB800]' : 'text-[#00FF94]'}`}>
                     {resultData.risk_score}/100
                   </div>
                </div>
             </div>
          )}

          {resultData.vulnerabilities && resultData.vulnerabilities.length > 0 && (
             <div className="bg-[#FF4444]/10 border border-[#FF4444]/30 rounded-lg p-4 mb-4">
                <div className="text-[#FF4444] font-bold mb-2">Vulnerabilities Detected</div>
                <ul className="list-disc pl-5 text-sm space-y-1">
                  {resultData.vulnerabilities.map((v, i) => (
                    <li key={i} className="text-red-200">{v}</li>
                  ))}
                </ul>
             </div>
          )}

          {resultData.audit_report && (
            <div className="bg-agent-card rounded-lg p-5 border border-agent-border">
              <h4 className="text-sm font-semibold text-[#00D4FF] mb-3">AI Audit Report</h4>
              <div className="whitespace-pre-wrap text-sm text-gray-200 leading-relaxed font-sans">
                {resultData.audit_report}
              </div>
            </div>
          )}

          {/* Fallback for unrecognized analysis schemas */}
          {resultData.risk_score === undefined && !resultData.vulnerabilities && !resultData.audit_report && !resultData.messages && (
            <pre className="bg-[#0A0A0F] p-4 rounded-lg overflow-x-auto text-xs text-[#00D4FF] font-mono border border-agent-border">
              {JSON.stringify(resultData, null, 2)}
            </pre>
          )}
        </div>
      );

    case 'business':
    case 'dao':
    case 'wildcard':
    default:
      if (resultData.summary) {
        return (
          <div className="bg-agent-card rounded-lg p-5 border border-agent-border">
            <div className="flex justify-between items-center mb-4">
              <span className="px-3 py-1 bg-[#00D4FF]/20 text-agent-primary text-xs rounded-full uppercase font-bold tracking-wider">
                Execution Output
              </span>
            </div>
            <div className="whitespace-pre-wrap text-sm text-gray-200 leading-relaxed font-sans">
              {typeof resultData.summary === 'string' ? resultData.summary : JSON.stringify(resultData.summary, null, 2)}
            </div>
            {resultData.details && (
               <div className="mt-4 pt-4 border-t border-agent-border">
                  <pre className="text-xs text-agent-muted font-mono whitespace-pre-wrap">
                    {typeof resultData.details === 'string' ? resultData.details : JSON.stringify(resultData.details, null, 2)}
                  </pre>
               </div>
            )}
          </div>
        );
      }
      return (
        <pre className="bg-[#0A0A0F] p-4 rounded-lg overflow-x-auto text-xs text-[#00D4FF] font-mono border border-agent-border">
          {JSON.stringify(resultData, null, 2)}
        </pre>
      );
  }
};
