export const mockTaskHistory = [
  {
    id: "task_001",
    agentId: 1,
    agentName: "Spot Trader",
    category: "defi",
    status: "success",
    timestamp: "2026-03-14T08:30:00Z",
    cost: "0.05",
    input: { token: "HLUSD", amount: 100, action: "buy" },
    result: { txHash: "0xabc123...def456", amountReceived: "99.85 HELA", gasUsed: "0.002" }
  },
  {
    id: "task_002",
    agentId: 8,
    agentName: "Content Writer",
    category: "content",
    status: "success",
    timestamp: "2026-03-14T07:15:00Z",
    cost: "0.01",
    input: { topic: "DeFi Yield Farming Guide", platform: "Twitter", tone: "professional" },
    result: { content: "🧵 Thread: DeFi Yield Farming in 2026 — A complete guide...", wordCount: 1200 }
  },
  {
    id: "task_003",
    agentId: 5,
    agentName: "Portfolio Monitor",
    category: "portfolio",
    status: "success",
    timestamp: "2026-03-13T22:00:00Z",
    cost: "0.001",
    input: { wallet: "0x7a3B...eF91", condition: "price_drop", threshold: 5 },
    result: { alertsTriggered: 0, portfolioValue: "$12,430", change24h: "+2.1%" }
  },
  {
    id: "task_004",
    agentId: 16,
    agentName: "Whale Watcher",
    category: "onchain",
    status: "failed",
    timestamp: "2026-03-13T18:45:00Z",
    cost: "0.01",
    input: { walletAddress: "0x1234...5678", minAmount: 100000 },
    result: { error: "RPC timeout — agent stake slashed 0.5 HLUSD to your wallet" }
  },
  {
    id: "task_005",
    agentId: 7,
    agentName: "PnL Reporter",
    category: "portfolio",
    status: "success",
    timestamp: "2026-03-13T09:00:00Z",
    cost: "0.02",
    input: { wallet: "0x7a3B...eF91", timePeriod: "7d" },
    result: { totalPnL: "+$234.50", winRate: "73%", bestTrade: "HELA/USDC +12.3%" }
  },
];

export const mockSchedules = [
  {
    id: "sched_001",
    agentId: 7,
    agentName: "PnL Reporter",
    schedule: "Every day at 9:00 AM",
    description: "Daily PnL Report",
    nextRun: "2026-03-15T09:00:00Z",
    isActive: true,
  },
  {
    id: "sched_002",
    agentId: 3,
    agentName: "Portfolio Rebalancer",
    schedule: "Every Monday at 10:00 AM",
    description: "Weekly Rebalance",
    nextRun: "2026-03-16T10:00:00Z",
    isActive: true,
  },
];

export const mockDashboardStats = {
  tasksRun: 12,
  hlusdSpent: 0.48,
  successful: 11,
  disputes: 1,
};

export const mockReputationTrend = [
  { day: "Mon", score: 91 },
  { day: "Tue", score: 92 },
  { day: "Wed", score: 90 },
  { day: "Thu", score: 93 },
  { day: "Fri", score: 95 },
  { day: "Sat", score: 94 },
  { day: "Sun", score: 96 },
];

export const mockEarningsData = [
  { name: "Earnings", value: 124.5, fill: "#00D4FF" },
  { name: "Stake", value: 480, fill: "#8B5CF6" },
];

export const mockManageTaskHistory = [
  { id: "mt_001", user: "0x1234...5678", status: "success", cost: "0.05", timestamp: "2026-03-14T10:00:00Z" },
  { id: "mt_002", user: "0x9abc...def0", status: "success", cost: "0.05", timestamp: "2026-03-14T09:30:00Z" },
  { id: "mt_003", user: "0x5678...9abc", status: "success", cost: "0.05", timestamp: "2026-03-14T08:15:00Z" },
  { id: "mt_004", user: "0xdef0...1234", status: "failed", cost: "0.05", timestamp: "2026-03-13T22:00:00Z" },
  { id: "mt_005", user: "0x3456...7890", status: "success", cost: "0.05", timestamp: "2026-03-13T18:45:00Z" },
  { id: "mt_006", user: "0x7890...abcd", status: "success", cost: "0.05", timestamp: "2026-03-13T15:30:00Z" },
  { id: "mt_007", user: "0xabcd...ef01", status: "success", cost: "0.05", timestamp: "2026-03-13T12:00:00Z" },
  { id: "mt_008", user: "0xef01...2345", status: "success", cost: "0.05", timestamp: "2026-03-12T20:00:00Z" },
  { id: "mt_009", user: "0x2345...6789", status: "disputed", cost: "0.05", timestamp: "2026-03-12T16:30:00Z" },
  { id: "mt_010", user: "0x6789...abcd", status: "success", cost: "0.05", timestamp: "2026-03-12T10:00:00Z" },
];

export const mockProtocolStats = {
  totalTasks: 24891,
  totalVolume: 8432,
  treasuryEarned: 843.2,
  agentsListed: 25,
  slashEvents: 12,
  usersProtected: 12,
};

export const mockRevenueBreakdown = [
  { name: "Protocol fees 10%", value: 748.2, fill: "#00D4FF" },
  { name: "Arena commission 2%", value: 62.4, fill: "#8B5CF6" },
  { name: "Listing fees", value: 25.0, fill: "#10B981" },
  { name: "Slash revenue 20%", value: 7.6, fill: "#F59E0B" },
];

export const mockTreasuryAllocation = [
  { name: "Development", value: 40, fill: "#00D4FF" },
  { name: "Stability", value: 30, fill: "#00FF94" },
  { name: "Community", value: 30, fill: "#8B5CF6" },
];

export const mockFeeStructure = [
  { type: "Protocol fee", rate: "10% per call" },
  { type: "Arena premium", rate: "+2% per arena" },
  { type: "Agent listing", rate: "5 HLUSD one-time" },
  { type: "Slash share", rate: "20% of slash" },
];
