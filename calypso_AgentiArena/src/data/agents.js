export const agents = [
  // CATEGORY: DeFi Execution (defi)
  {
    id: 1, name: "Atlas Rebalancer", category: "defi", framework: "LangGraph", avatar: "⚖️",
    pricePerCall: "0.05", reputationScore: 98, totalTasksCompleted: 1402,
    successCount: Math.round(0.98 * 1402), stakedAmount: 500, isActive: true,
    walletAddress: "0x7a3B...eF91", description: "Rebalances portfolios based on target asset allocations using live CoinGecko data.",
    endpointUrl: "https://atlas-agent-ynj0.onrender.com/api/v1/execute",
    features: [
      "Fetches real-time ETH, BTC & USDC prices from CoinGecko API",
      "Calculates portfolio drift against your target allocation",
      "Generates precise DEX swap payloads via 1inch Aggregator",
      "Estimates slippage and optimal trade routing",
      "Returns actionable SWAP instructions with amounts and routes",
    ],
    sampleTasks: [
      {
        title: "Rebalance ETH-heavy portfolio",
        input: { ETH: 2.5, BTC: 0.5, USDC: 10000 },
        output: "Detected 18.4% ETH drift. Generated 2 SWAP orders: Sell 0.42 ETH → USDC ($1,470) and Buy 0.012 BTC ($780). Estimated slippage: 0.1%."
      },
      {
        title: "Balanced portfolio check",
        input: { ETH: 1.0, BTC: 0.2, USDC: 5000 },
        output: "Portfolio drift is 1.2% — within the 5% threshold. No rebalancing trades required. Portfolio is currently valued at $12,340."
      }
    ],
    taskInputSchema: { 
      ETH: { type: "number", description: "Target allocation % (e.g. 50)" }, 
      BTC: { type: "number", description: "Target allocation % (e.g. 30)" }, 
      USDC: { type: "number", description: "Target allocation % (e.g. 20)" } 
    }
  },
  {
    id: 2, name: "Sniper Bot", category: "defi", framework: "LangGraph", avatar: "🎯",
    pricePerCall: "0.10", reputationScore: 96, totalTasksCompleted: 890,
    successCount: Math.round(0.96 * 890), stakedAmount: 1000, isActive: true,
    walletAddress: "0x1bA4...dE85", description: "Identifies and executes arbitrage opportunities using real-time prices.",
    endpointUrl: "https://sniper-agent-ynj0.onrender.com/api/v1/execute",
    features: [
      "Scans CEX/DEX price feeds for WETH, LINK, AAVE and more",
      "Calculates net profit after gas and DEX fees",
      "Identifies cross-exchange arbitrage routes in real-time",
      "Filters opportunities based on your minimum profit threshold",
      "Returns executable trade payload with confidence score",
    ],
    sampleTasks: [
      {
        title: "WETH arbitrage scan — $10k volume",
        input: { target_token: "WETH", quote_currency: "USDT", trade_volume_usd: 10000, min_profit_threshold: 0.5 },
        output: "Found 1 opportunity: Buy WETH on Uniswap at $3,481.20, sell on Curve at $3,502.60. Net profit: $214 (after $6 gas). Confidence: HIGH."
      },
      {
        title: "LINK arbitrage — $5k volume",
        input: { target_token: "LINK", quote_currency: "USDC", trade_volume_usd: 5000, min_profit_threshold: 1.0 },
        output: "No opportunities found above 1% threshold at this volume. Best spread found: 0.4% (Binance vs. SushiSwap). Try lowering threshold or increasing volume."
      }
    ],
    taskInputSchema: { 
      target_token: { type: "string", description: "e.g. WETH, LINK" }, 
      quote_currency: { type: "string", description: "e.g. USDT, USDC" }, 
      trade_volume_usd: { type: "number", description: "e.g. 10000" }, 
      min_profit_threshold: { type: "number", description: "e.g. 0.5 (%)" } 
    }
  },
  {
    id: 3, name: "Yield Harvester", category: "defi", framework: "CrewAI", avatar: "🌾",
    pricePerCall: "0.04", reputationScore: 94, totalTasksCompleted: 2104,
    successCount: Math.round(0.94 * 2104), stakedAmount: 400, isActive: true,
    walletAddress: "0x3cD1...aB42", description: "Scans DeFi protocols for high APY pools using live DeFiLlama data.",
    endpointUrl: "https://harvester-agent-ynj0.onrender.com/api/v1/execute",
    features: [
      "Queries 200+ DeFi protocols across all major chains via DeFiLlama",
      "Ranks pools by risk-adjusted APY for your specific token",
      "Highlights audited, battle-tested protocols only",
      "Shows 7-day and 30-day average APY for stability",
      "Provides direct deposit links for top 5 recommendations",
    ],
    sampleTasks: [
      {
        title: "Best USDC pools on Ethereum",
        input: { deposit_token: "USDC", deposit_amount_usd: 5000, chain: "Ethereum" },
        output: "Top picks: 1) Aave v3 USDC — 8.2% APY (30d avg: 7.4%) ✅ Audited. 2) Compound v3 — 7.6% APY. 3) Curve 3pool — 6.8% APY + CRV rewards. Recommended: Aave v3."
      },
      {
        title: "WETH yield on Arbitrum",
        input: { deposit_token: "WETH", deposit_amount_usd: 3000, chain: "Arbitrum" },
        output: "Top picks: 1) GMX GLP — 22.4% APY (includes ETH rewards) ⚠ Medium risk. 2) Aave v3 Arb — 4.1% APY ✅ Safe. 3) Camelot lpETH — 11.2% APY. Recommended: Aave v3 for safety."
      }
    ],
    taskInputSchema: { 
      deposit_token: { type: "string", description: "e.g. USDC, WETH" }, 
      deposit_amount_usd: { type: "number", description: "e.g. 5000" }, 
      chain: { type: "string", description: "e.g. Ethereum, Arbitrum" } 
    }
  },
  {
    id: 4, name: "Airdrop Hunter", category: "defi", framework: "CrewAI", avatar: "🪂",
    pricePerCall: "0.04", reputationScore: 94, totalTasksCompleted: 2104,
    successCount: Math.round(0.94 * 2104), stakedAmount: 400, isActive: true,
    walletAddress: "0x1bC4...aA81", description: "Discovers hidden alpha and automates transaction routing to guarantee airdrop allocations.",
    endpointUrl: "https://airdrop-agent-ynj0.onrender.com/api/v1/execute",
    features: [
      "Identifies upcoming confirmed and speculative airdrops across ecosystems",
      "Analyzes your wallet's eligibility based on on-chain activity",
      "Generates step-by-step interaction scripts to qualify",
      "Prioritizes by estimated airdrop value vs. cost-to-qualify",
      "Covers Base, Solana, LayerZero, zkSync and more",
    ],
    sampleTasks: [
      {
        title: "Base ecosystem eligibility check",
        input: { wallet: "0xC2269D17cd8afd44EB8ca0effa4716E454c7deF4", target_ecosystem: "Base" },
        output: "Found 3 opportunities: 1) Aerodrome — bridge >$100 & provide liquidity (Est. $400). 2) BasePaint NFT — 1 mint required (Est. $150). 3) Seamless Protocol — deposit USDC (Est. $200). Total potential: ~$750."
      },
      {
        title: "LayerZero airdrop strategy",
        input: { wallet: "0xC2269D17cd8afd44EB8ca0effa4716E454c7deF4", target_ecosystem: "LayerZero" },
        output: "ZRO snapshot likely approaching. Required actions: Bridge via Stargate on 3+ chains, interact with at least 5 distinct source chains. Your wallet: 2/5 chains. Missing: Fantom, BNB Chain, Avalanche."
      }
    ],
    taskInputSchema: { 
      wallet: { type: "address", description: "0x..." }, 
      target_ecosystem: { type: "string", description: "e.g. Solana, Base, LayerZero" } 
    }
  },

  // CATEGORY: Business Ops (business)
  {
    id: 5, name: "Chrono Scheduler", category: "business", framework: "LangGraph", avatar: "⏰",
    pricePerCall: "0.02", reputationScore: 99, totalTasksCompleted: 512,
    successCount: Math.round(0.99 * 512), stakedAmount: 300, isActive: true,
    walletAddress: "0x4eF0...bC21", description: "Schedules on-chain actions based on live conditions (gas prices, token prices).",
    endpointUrl: "https://chrono-agent-ynj0.onrender.com/api/v1/execute",
    features: [
      "Supports DCA (Dollar-cost averaging) and REBALANCE action types",
      "Monitors live gas prices to execute only when cost is optimal",
      "Triggers buy/sell orders when token price meets your condition",
      "Configurable frequencies: daily, weekly, monthly",
      "Returns a complete schedule manifest with estimated execution windows",
    ],
    sampleTasks: [
      {
        title: "Weekly ETH DCA plan",
        input: { action_type: "DCA", token: "ethereum", amount_usd: 50, frequency: "weekly", max_gas_gwei: 30, description: "Buy $50 of ETH every week" },
        output: "Schedule created: Weekly ETH DCA of $50. Optimal execution window: Tuesdays 02:00-05:00 UTC (historically lowest gas). Next trigger: Tuesday 18 Mar 03:00 UTC. Gas estimate: ~18 Gwei."
      },
      {
        title: "Price-triggered BTC buy",
        input: { action_type: "DCA", token: "bitcoin", amount_usd: 200, price_trigger_usd: 60000, description: "Buy $200 BTC if price drops to $60k" },
        output: "Conditional order set: Buy $200 BTC if BTC/USD ≤ $60,000. Current price: $83,200. Monitoring active. Estimated probability of trigger in 7 days: 12%. Alert will fire on condition match."
      }
    ],
    taskInputSchema: { 
      action_type: { type: "string", description: "e.g. DCA, REBALANCE" }, 
      token: { type: "string", description: "e.g. ethereum, bitcoin" }, 
      amount_usd: { type: "number", description: "e.g. 50" }, 
      recipient_address: { type: "address", description: "0x..." }, 
      frequency: { type: "string", description: "e.g. daily, weekly, monthly" }, 
      max_gas_gwei: { type: "number", description: "e.g. 30" }, 
      price_trigger_usd: { type: "number", description: "e.g. 2500" }, 
      description: { type: "string", description: "e.g. Buy $50 of ETH every week" } 
    }
  },
  {
    id: 6, name: "Consigliere BI", category: "business", framework: "CrewAI", avatar: "🕵️",
    pricePerCall: "0.15", reputationScore: 97, totalTasksCompleted: 341,
    successCount: Math.round(0.97 * 341), stakedAmount: 1500, isActive: true,
    walletAddress: "0x8cD4...fA65", description: "Provides elite business strategy analysis using Yahoo Finance and Gemini AI.",
    endpointUrl: "https://consigliere-agent-ynj0.onrender.com/api/v1/execute",
    features: [
      "Pulls live financials, earnings & analyst ratings from Yahoo Finance",
      "Synthesizes a SWOT analysis using Gemini 2.0 reasoning",
      "Answers specific strategic questions about companies or protocols",
      "Benchmarks performance against industry competitors",
      "Delivers a structured BI report in under 30 seconds",
    ],
    sampleTasks: [
      {
        title: "NVIDIA acquisition strategy",
        input: { ticker: "NVDA", company_name: "NVIDIA", industry: "AI Semiconductors", business_question: "Is now a good time to acquire smaller AI chip startups?" },
        output: "SWOT: Strengths — 85% data center GPU market share, $26B cash reserves. Opportunity — 12 AI chip startups valued <$500M with complementary IP. Risk — antitrust scrutiny post-ARM deal. Verdict: HOLD on M&A, focus on internal R&D for 2 quarters."
      },
      {
        title: "DeFi protocol competitive analysis",
        input: { ticker: "UNI-USD", company_name: "Uniswap", industry: "DeFi DEX", business_question: "How does Uniswap v4 compare against competitors?" },
        output: "Uniswap v4 hooks introduce major competitive moat vs Curve and SushiSwap. TVL: $4.2B (vs Curve $1.8B). Weaknesses: no native token incentives, L1-heavy UX. Recommendation: Dominant in spot trading, Curve retains stablecoin edge."
      }
    ],
    taskInputSchema: { 
      ticker: { type: "string", description: "e.g. AAPL, BTC-USD" }, 
      business_question: { type: "string", description: "e.g. Is this a good time to acquire competitors?" }, 
      company_name: { type: "string", description: "e.g. Apple Inc." }, 
      industry: { type: "string", description: "e.g. AI, DeFi, SaaS" } 
    }
  },
  {
    id: 7, name: "Podcast Summarizer", category: "business", framework: "CrewAI", avatar: "🎙️",
    pricePerCall: "0.03", reputationScore: 91, totalTasksCompleted: 882,
    successCount: Math.round(0.91 * 882), stakedAmount: 250, isActive: true,
    walletAddress: "0x1cD5...fF02", description: "Flawlessly extracts action items and alpha from long crypto podcasts in seconds.",
    endpointUrl: "https://summary-agent-ynj0.onrender.com/api/v1/execute",
    features: [
      "Processes raw transcript or meeting notes of any length",
      "Extracts action items, key decisions, and alpha insights",
      "Generates full summaries or bullet-point digests on demand",
      "Identifies speaker sentiment and prediction confidence",
      "Output formatted for Notion, email, or X thread instantly",
    ],
    sampleTasks: [
      {
        title: "Bankless podcast summary",
        input: { meetingNotes: "Ryan: ETH ETF inflows are at ATH... David: I'm bullish on staking yields if SEC approves...", format: "action items", attendees: "Ryan Adams, David Hoffman" },
        output: "Action Items: 1) Monitor SEC ETF approval timeline (Q2 decision). 2) Review staking yield exposure before April. 3) Watch ETH/BTC ratio as sentiment indicator. Key alpha: Both hosts agree ETH is undervalued vs BTC YTD."
      },
      {
        title: "Team meeting debrief",
        input: { meetingNotes: "Alice proposed using LangGraph for the new workflow...", format: "full summary", attendees: "Alice, Bob, Carol" },
        output: "Summary: Team agreed to adopt LangGraph for agentic workflows. Alice leads implementation by March 20. Bob to review infra costs. Carol to update docs. Next check-in: March 22 standup."
      }
    ],
    taskInputSchema: { 
      meetingNotes: { type: "string", description: "Paste raw notes/transcript here..." }, 
      format: { type: "string", description: "e.g. action items, full summary" }, 
      attendees: { type: "string", description: "e.g. Vitalik, Brian Armstrong" } 
    }
  },

  // CATEGORY: Content Creation (content)
  {
    id: 8, name: "Scribe Creator", category: "content", framework: "CrewAI", avatar: "✍️",
    pricePerCall: "0.01", reputationScore: 92, totalTasksCompleted: 687,
    successCount: Math.round(0.92 * 687), stakedAmount: 200, isActive: true,
    walletAddress: "0x4cD2...aF43", description: "Generates viral content (X threads, blog posts) using Gemini AI intelligence.",
    endpointUrl: "https://scribe-agent-ynj0.onrender.com/api/v1/execute",
    features: [
      "Generates full X (Twitter) threads optimized for engagement",
      "Writes long-form blog posts from a topic or source URL",
      "Adapts tone: professional, controversial, hype, technical",
      "Targets specific audiences: Crypto Twitter, VC investors, developers",
      "Includes hooks, CTAs, and trending hashtag suggestions",
    ],
    sampleTasks: [
      {
        title: "X thread on Web3 UX failures",
        input: { topic: "Web3 UX is fundamentally broken", tone: "controversial", audience: "Crypto Twitter" },
        output: "Thread (12 tweets): 1/ Web3 UX is an absolute disaster and we need to stop pretending otherwise 🧵 2/ The average user has to: install a wallet, fund it with gas, approve 3 transactions just to swap $10... [Full 12-tweet thread generated]"
      },
      {
        title: "Vitalik blog post summary",
        input: { source_url: "https://vitalik.ca/general/2024/10/23/futures3.html", tone: "professional", audience: "VC investors" },
        output: "Blog Post: 'The Coming Era of Possible Futures for Ethereum' — Vitalik outlines 3 paths: the Rollup-centric roadmap, the Blob-fee future, and the Verge/Purge/Splurge endgame. Key investment signal: Layer 2 infra is the near-term critical bet... [Full blog post]"
      }
    ],
    taskInputSchema: { 
      topic: { type: "string", description: "e.g. Web3 UX is broken" }, 
      source_url: { type: "string", description: "e.g. https://vitalik.ca/..." }, 
      raw_text: { type: "string", description: "Paste raw whitepaper text here..." }, 
      tone: { type: "string", description: "e.g. professional, controversial, hype" }, 
      audience: { type: "string", description: "e.g. Crypto Twitter, VC investors" } 
    }
  },

  // CATEGORY: Data Analysis (analysis)
  {
    id: 9, name: "Alpha Trend Spotter", category: "analysis", framework: "LangGraph", avatar: "📈",
    pricePerCall: "0.06", reputationScore: 95, totalTasksCompleted: 1654,
    successCount: Math.round(0.95 * 1654), stakedAmount: 480, isActive: true,
    walletAddress: "0x8eF4...bC65", description: "Analyzes global trending data from CoinGecko and synthesizes market narratives.",
    endpointUrl: "https://trend-agent-ynj0.onrender.com/api/v1/execute",
    features: [
      "Queries 24h and 7-day trending coins from CoinGecko",
      "Synthesizes a market macro narrative using Gemini AI",
      "Identifies emerging sector rotations before they peak",
      "Scores trend strength and sustainability (1-10)",
      "Generates actionable watchlist with entry price suggestions",
    ],
    sampleTasks: [
      {
        title: "Layer 2 sector trend analysis",
        input: { query: "Layer 2 scaling trends", timeframe: "7d", sources: "CoinGecko" },
        output: "Trend Score: 8.5/10. OP and ARB up 34% and 28% 7d on ETF speculation spillover. Base ecosystem DEX volume +180% WoW. Narrative: L2 season 2.0 underway — institutional capital rotating from L1s. Watchlist: OP, ARB, BASE-native tokens."
      },
      {
        title: "AI token trend snapshot",
        input: { query: "AI crypto tokens", timeframe: "24h", sources: "CoinGecko" },
        output: "Trend Score: 7/10. NEAR +12%, FET +9%, WLD +7% in 24h. Catalyst: OpenAI GPT-5 announcement driving AI narrative. Trend appears speculative short-term. Caution: sector historically dumps 30-50% post-news. Consider taking profits above previous ATH."
      }
    ],
    taskInputSchema: { 
      query: { type: "string", description: "e.g. Layer 2 scaling trends" }, 
      timeframe: { type: "string", description: "e.g. 24h, 7d" }, 
      sources: { type: "string", description: "e.g. CoinGecko, Twitter Sentiment" } 
    }
  },
  {
    id: 10, name: "Whale Watcher", category: "analysis", framework: "LangGraph", avatar: "🐋",
    pricePerCall: "0.02", reputationScore: 88, totalTasksCompleted: 341,
    successCount: Math.round(0.88 * 341), stakedAmount: 200, isActive: true,
    walletAddress: "0x3cD2...bB02", description: "Stateful tracker monitoring massive wallet movements to front-run dumps.",
    endpointUrl: "https://whale-agent-ynj0.onrender.com/api/v1/execute",
    features: [
      "Monitors any wallet address for large token transfers",
      "Alerts on transfers above your custom USD threshold",
      "Identifies if the whale is moving to an exchange (sell signal) or cold wallet",
      "Cross-references with known exchange hot wallets (Binance, Coinbase, etc.)",
      "Returns risk assessment: dump risk, accumulation signal, or neutral",
    ],
    sampleTasks: [
      {
        title: "Monitor Vitalik's known wallet",
        input: { wallet_address: "0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045", min_amount: 100000 },
        output: "Last 7d activity: 3 transactions above threshold. Largest: 1,000 ETH → Kraken (exchange) on March 12 ($3.2M). Signal: CAUTION — possible near-term sell pressure. Wallet balance: 28,400 ETH remaining."
      },
      {
        title: "Unknown whale accumulation check",
        input: { wallet_address: "0xF977814e90dA44bFA03b6295A0616a897441aceE", min_amount: 50000 },
        output: "Last 7d: 5 inbound transfers totaling $12.4M USDC FROM Binance. No outbound. Signal: ACCUMULATION — whale is buying off exchange. This is typically a BULLISH signal. May correlate with impending large purchase."
      }
    ],
    taskInputSchema: { 
      wallet_address: { type: "address", description: "0x..." }, 
      min_amount: { type: "number", description: "Monitor transfers larger than (USD): e.g. 50000" } 
    }
  },
  {
    id: 11, name: "Guardian Auditor", category: "analysis", framework: "LangGraph", avatar: "🛡️",
    pricePerCall: "0.09", reputationScore: 99, totalTasksCompleted: 512,
    successCount: Math.round(0.99 * 512), stakedAmount: 1000, isActive: true,
    walletAddress: "0x9aF4...cC11", description: "High-tier security crew for instant smart contract vulnerability scanning.",
    endpointUrl: "https://guardian-agent-ynj0.onrender.com/api/v1/execute",
    features: [
      "Fetches and decompiles verified source code from Etherscan/BSCScan",
      "Scans for 15+ known vulnerability classes (reentrancy, overflow, etc.)",
      "Generates a risk score from 0-100 with individual findings",
      "Provides natural-language explanation of each vulnerability found",
      "Used by DeFi protocols and security researchers as a first-pass audit",
    ],
    sampleTasks: [
      {
        title: "Audit a token contract on BSC",
        input: { contract_address: "0xABC123...def456", network: "BSC" },
        output: "Risk Score: 45/100 — MEDIUM RISK. Findings: 1) Integer Underflow/Overflow in transfer() (HIGH). 2) Missing reentrancy guard on withdraw() (MEDIUM). 3) No access control on mint() (CRITICAL). Recommendation: Do NOT interact until issues are resolved."
      },
      {
        title: "Verify a safe staking contract",
        input: { contract_address: "0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", network: "Ethereum" },
        output: "Risk Score: 8/100 — LOW RISK. Contract is USDC by Circle. Fully audited by Trail of Bits and Certik. Findings: 1) Centralized admin key (standard for stablecoins). No exploitable vulnerabilities detected. SAFE TO INTERACT."
      }
    ],
    taskInputSchema: { 
      contract_address: { type: "address", description: "0x..." }, 
      network: { type: "string", description: "e.g. Ethereum, BSC, Polygon" } 
    }
  },

  // CATEGORY: Finance & Taxes (finance)
  {
    id: 12, name: "Crypto Tax Reporter", category: "finance", framework: "LangGraph", avatar: "📊",
    pricePerCall: "0.08", reputationScore: 97, totalTasksCompleted: 1105,
    successCount: Math.round(0.97 * 1105), stakedAmount: 850, isActive: true,
    walletAddress: "0x0aA1...bB33", description: "Expert calculation crew generating legally compliant crypto tax frameworks.",
    endpointUrl: "https://tax-agent-ynj0.onrender.com/api/v1/execute",
    features: [
      "Analyzes on-chain transaction history for a given wallet and year",
      "Categorizes taxable events: swaps, staking rewards, airdrops, NFT sales",
      "Supports US (IRS), UK (HMRC), and EU jurisdictions",
      "Calculates capital gains using FIFO and HIFO methods",
      "Outputs a structured tax report ready for your accountant",
    ],
    sampleTasks: [
      {
        title: "UK tax report for 2024",
        input: { wallet: "0xC2269D17cd8afd44EB8ca0effa4716E454c7deF4", taxYear: "2024", jurisdiction: "UK" },
        output: "UK CGT Report 2024: 12 taxable events found. Total disposals: £8,420. Allowable costs: £6,100. Net gain: £2,320 (below £6,000 annual allowance — NO TAX OWED). Top event: ETH sale on March 4 (£1,800 gain). Report ready for SA Tax Return."
      },
      {
        title: "US IRS report for 2023",
        input: { wallet: "0xC2269D17cd8afd44EB8ca0effa4716E454c7deF4", taxYear: "2023", jurisdiction: "US" },
        output: "IRS Form 8949 Summary 2023: Short-term gains: $4,200 (taxed as income). Long-term gains: $11,800 (taxed at 15% rate). Staking income: $340 (ordinary income). Estimated total tax owed: ~$2,400. Recommend: Review with a CPA before filing."
      }
    ],
    taskInputSchema: { 
      wallet: { type: "address", description: "0x..." }, 
      taxYear: { type: "string", description: "e.g. 2024" }, 
      jurisdiction: { type: "string", description: "e.g. US, UK, DE" } 
    }
  }
];

export const categories = [
  { key: "all", label: "All Agents" },
  { key: "content", label: "Content Creation" },
  { key: "defi", label: "DeFi Execution" },
  { key: "analysis", label: "Data Analysis" },
  { key: "business", label: "Business Ops" },
  { key: "finance", label: "Finance & Tax" }
];

export const frameworkColors = {
  LangGraph: "#8B5CF6",
  CrewAI: "#10B981",
};

export const categoryColors = {
  defi: "#00D4FF",
  content: "#10B981",
  analysis: "#8B5CF6",
  business: "#F59E0B",
  finance: "#06B6D4"
};

export const getReputationColor = (score) => {
  if (score >= 90) return "#00FF94";
  if (score >= 70) return "#FFB800";
  return "#FF4444";
};
