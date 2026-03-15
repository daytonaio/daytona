const hre = require("hardhat");
const fs = require("fs");
const path = require("path");

async function main() {
  const [deployer] = await hre.ethers.getSigners();
  console.log("Deploying contracts with the account:", deployer.address);

  const treasury = process.env.TREASURY_ADDRESS || deployer.address;
  console.log("Using treasury address:", treasury);

  // 1. Deploy MockHLUSD
  const MockHLUSD = await hre.ethers.getContractFactory("MockHLUSD");
  const hlusd = await MockHLUSD.deploy();
  await hlusd.waitForDeployment();
  const hlusdAddress = await hlusd.getAddress();
  console.log("MockHLUSD deployed to:", hlusdAddress);

  // 2. Deploy AgentVault
  const AgentVault = await hre.ethers.getContractFactory("AgentVault");
  const vault = await AgentVault.deploy(hlusdAddress, treasury);
  await vault.waitForDeployment();
  const vaultAddress = await vault.getAddress();
  console.log("AgentVault deployed to:", vaultAddress);

  // 3. Deploy AgentRegistry
  const AgentRegistry = await hre.ethers.getContractFactory("AgentRegistry");
  const registry = await AgentRegistry.deploy(hlusdAddress, treasury);
  await registry.waitForDeployment();
  const registryAddress = await registry.getAddress();
  console.log("AgentRegistry deployed to:", registryAddress);

  // Link Registry and Vault
  await registry.setVault(vaultAddress);
  console.log("Vault linked to Registry");

  // 4. Deploy TaskLedger
  const TaskLedger = await hre.ethers.getContractFactory("TaskLedger");
  const taskLedger = await TaskLedger.deploy(registryAddress, vaultAddress);
  await taskLedger.waitForDeployment();
  const taskLedgerAddress = await taskLedger.getAddress();
  console.log("TaskLedger deployed to:", taskLedgerAddress);

  // 5. Deploy ReputationEngine
  const ReputationEngine = await hre.ethers.getContractFactory("ReputationEngine");
  const reputation = await ReputationEngine.deploy(registryAddress);
  await reputation.waitForDeployment();
  const reputationAddress = await reputation.getAddress();
  console.log("ReputationEngine deployed to:", reputationAddress);
  await reputation.setVault(vaultAddress);

  // 6. Deploy ArenaEngine
  const ArenaEngine = await hre.ethers.getContractFactory("ArenaEngine");
  const arena = await ArenaEngine.deploy(hlusdAddress, registryAddress, vaultAddress, taskLedgerAddress);
  await arena.waitForDeployment();
  const arenaAddress = await arena.getAddress();
  console.log("ArenaEngine deployed to:", arenaAddress);

  // 7. Deploy DisputeResolver
  const DisputeResolver = await hre.ethers.getContractFactory("DisputeResolver");
  const dispute = await DisputeResolver.deploy(taskLedgerAddress, vaultAddress, registryAddress);
  await dispute.waitForDeployment();
  const disputeAddress = await dispute.getAddress();
  console.log("DisputeResolver deployed to:", disputeAddress);

  // Authorize contracts
  console.log("Authorizing contracts cross-communication...");
  await vault.setAuthorized(registryAddress, true);
  await vault.setAuthorized(taskLedgerAddress, true);
  await vault.setAuthorized(arenaAddress, true);
  await vault.setAuthorized(disputeAddress, true);
  
  await registry.transferOwnership(reputationAddress); // Since rep engine auto-suspends

  await taskLedger.setAuthorized(arenaAddress, true);
  await taskLedger.setAuthorized(disputeAddress, true);
  await taskLedger.setAuthorized(deployer.address, true); // For seeding tasks

  await reputation.setAuthorized(taskLedgerAddress, true);
  await reputation.setAuthorized(deployer.address, true); // For seeding scores

  console.log("\n--- Deployment Complete ---\n");
  console.log(`VITE_CONTRACT_HLUSD=${hlusdAddress}`);
  console.log(`VITE_CONTRACT_VAULT=${vaultAddress}`);
  console.log(`VITE_CONTRACT_REGISTRY=${registryAddress}`);
  console.log(`VITE_CONTRACT_TASK_LEDGER=${taskLedgerAddress}`);
  console.log(`VITE_CONTRACT_REPUTATION=${reputationAddress}`);
  console.log(`VITE_CONTRACT_ARENA=${arenaAddress}`);
  console.log(`VITE_CONTRACT_DISPUTE=${disputeAddress}`);

  // Create .env content
  let envContent = fs.existsSync(".env") ? fs.readFileSync(".env", "utf8") : "";
  const updateEnv = (key, value) => {
    const regex = new RegExp(`^${key}=.*`, "m");
    if (envContent.match(regex)) {
      envContent = envContent.replace(regex, `${key}=${value}`);
    } else {
      envContent += `\n${key}=${value}`;
    }
  };

  updateEnv("VITE_CONTRACT_HLUSD", hlusdAddress);
  updateEnv("VITE_CONTRACT_VAULT", vaultAddress);
  updateEnv("VITE_CONTRACT_REGISTRY", registryAddress);
  updateEnv("VITE_CONTRACT_TASK_LEDGER", taskLedgerAddress);
  updateEnv("VITE_CONTRACT_REPUTATION", reputationAddress);
  updateEnv("VITE_CONTRACT_ARENA", arenaAddress);
  updateEnv("VITE_CONTRACT_DISPUTE", disputeAddress);

  fs.writeFileSync(".env", envContent);
  console.log("Addresses saved to .env");

  // Read ABIs
  const getAbi = (name) => {
    return JSON.parse(fs.readFileSync(path.join(__dirname, "../artifacts/contracts", `${name}.sol`, `${name}.json`), "utf8")).abi;
  };

  const hlusdAbi = getAbi("MockHLUSD");
  const registryAbi = getAbi("AgentRegistry");
  const vaultAbi = getAbi("AgentVault");
  const ledgerAbi = getAbi("TaskLedger");
  const repAbi = getAbi("ReputationEngine");
  const arenaAbi = getAbi("ArenaEngine");
  const disputeAbi = getAbi("DisputeResolver");

  // Save for frontend
  const addressesContent = `
export const CONTRACT_ADDRESSES = {
  hlusd: "${hlusdAddress}",
  registry: "${registryAddress}",
  vault: "${vaultAddress}",
  taskLedger: "${taskLedgerAddress}",
  reputation: "${reputationAddress}",
  arena: "${arenaAddress}",
  dispute: "${disputeAddress}"
};

export const ABIS = {
  MockHLUSD: ${JSON.stringify(hlusdAbi, null, 2)},
  AgentRegistry: ${JSON.stringify(registryAbi, null, 2)},
  AgentVault: ${JSON.stringify(vaultAbi, null, 2)},
  TaskLedger: ${JSON.stringify(ledgerAbi, null, 2)},
  ReputationEngine: ${JSON.stringify(repAbi, null, 2)},
  ArenaEngine: ${JSON.stringify(arenaAbi, null, 2)},
  DisputeResolver: ${JSON.stringify(disputeAbi, null, 2)}
};
  `.trim();

  fs.mkdirSync(path.join(__dirname, "../src/contracts"), { recursive: true });
  fs.writeFileSync(path.join(__dirname, "../src/contracts/addresses.js"), addressesContent);
  console.log("Addresses saved to src/contracts/addresses.js");
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
