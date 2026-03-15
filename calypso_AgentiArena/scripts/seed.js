const hre = require("hardhat");
const { addresses } = require("../src/contracts/addresses"); // We need to write a simple reader or just require the json
const fs = require("fs");
const path = require("path");

// Read agents from frontend data
// We'll run this from the project root
// Since it's an ES module in frontend, we might need a workaround for commonjs script
// So we'll fetch the agents array using a dynamic import or just read and regex it.
// Another option is to define the exact agents again, but the spec says "matching src/data/agents.js exactly"

async function main() {
  const [deployer] = await hre.ethers.getSigners();
  console.log("Seeding with account:", deployer.address);

  // Address lookup
  const envContent = fs.readFileSync(".env", "utf8");
  const getEnv = (key) => {
    const match = envContent.match(new RegExp(`^${key}=(.*)$`, "m"));
    return match ? match[1] : null;
  };

  const hlusdAddress = getEnv("VITE_CONTRACT_HLUSD");
  const registryAddress = getEnv("VITE_CONTRACT_REGISTRY");
  const reputationAddress = getEnv("VITE_CONTRACT_REPUTATION");

  if (!hlusdAddress || !registryAddress || !reputationAddress) {
    throw new Error("Missing contract addresses in .env");
  }

  const MockHLUSD = await hre.ethers.getContractAt("MockHLUSD", hlusdAddress);
  const AgentRegistry = await hre.ethers.getContractAt("AgentRegistry", registryAddress);
  const ReputationEngine = await hre.ethers.getContractAt("ReputationEngine", reputationAddress);

  const amountToMint = hre.ethers.parseUnits("100000", 18);
  console.log(`Minting 100,000 HLUSD to deployer for stakes...`);
  await MockHLUSD.mint(deployer.address, amountToMint);
  
  // Approve for all agents
  console.log(`Approving HLUSD spending for AgentRegistry...`);
  await MockHLUSD.approve(registryAddress, hre.ethers.MaxUint256);

  // Define 5 mock agents for seeding (since 25 takes a long time on testnet script, we'll do 25 if possible, but let's do 5 first)
  // The spec says "all 25 agents", so let's import the file
  const agentsDataPath = path.join(__dirname, "../src/data/agents.js");
  const agentsDataStr = fs.readFileSync(agentsDataPath, "utf8");
  
  // Extract the agents array string using a hacky but functional approach
  const match = agentsDataStr.match(/export const agents = (\[[\s\S]*?\]);/);
  if (!match) throw new Error("Could not parse agents.js");
  
  const agents = eval(`(${match[1]})`);
  console.log(`Found ${agents.length} agents to deploy.`);

  for (let i = 0; i < agents.length; i++) {
    const agent = agents[i];
    console.log(`Registering [${i + 1}/${agents.length}]: ${agent.name}...`);
    
    // Some agents might not have pricePerCall as float, so parse cleanly
    const priceStr = agent.pricePerCall ? agent.pricePerCall.toString() : "0.01";
    const priceWei = hre.ethers.parseUnits(priceStr, 18);
    const stakeWei = hre.ethers.parseUnits((agent.stakedAmount || 100).toString(), 18);
    
    const tx = await AgentRegistry.registerAgent(
      agent.name,
      agent.category,
      agent.framework,
      priceWei,
      "https://api.agentarena.network/agent/" + agent.id,
      agent.description || "Top tier AI agent",
      stakeWei
    );
    await tx.wait(); // Wait for 1 confirmation
    
    // Set mock stats
    const totalTasks = agent.totalTasksCompleted || 0;
    const successCount = agent.successCount || 0;
    
    if (totalTasks > 0) {
      await AgentRegistry.setMockStats(i + 1, totalTasks, successCount);
      
      // Update reputation score manually (testnet backdoor)
      const repScore = Math.floor((successCount * 100) / totalTasks);
      await ReputationEngine.setMockScore(i + 1, repScore);
    }
  }

  console.log("All agents seeded successfully!");
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
