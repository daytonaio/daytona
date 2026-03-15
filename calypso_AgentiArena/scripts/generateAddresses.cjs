const fs = require('fs');
const path = require('path');

const getAbi = (name) => {
  const filePath = path.join(__dirname, '..', 'artifacts', 'contracts', `${name}.sol`, `${name}.json`);
  return JSON.parse(fs.readFileSync(filePath, 'utf8')).abi;
};

const abis = {
  MockHLUSD: getAbi('MockHLUSD'),
  AgentRegistry: getAbi('AgentRegistry'),
  AgentVault: getAbi('AgentVault'),
  TaskLedger: getAbi('TaskLedger'),
  ReputationEngine: getAbi('ReputationEngine'),
  ArenaEngine: getAbi('ArenaEngine'),
  DisputeResolver: getAbi('DisputeResolver')
};

const ZERO = '0x0000000000000000000000000000000000000000';

const content = `// Auto-generated contract addresses and ABIs
// Replace placeholder addresses after deploying to HeLa Testnet

export const CONTRACT_ADDRESSES = {
  hlusd: "${ZERO}",
  registry: "${ZERO}",
  vault: "${ZERO}",
  taskLedger: "${ZERO}",
  reputation: "${ZERO}",
  arena: "${ZERO}",
  dispute: "${ZERO}"
};

export const ABIS = ${JSON.stringify(abis, null, 2)};
`;

const outDir = path.join(__dirname, '..', 'src', 'contracts');
fs.mkdirSync(outDir, { recursive: true });
fs.writeFileSync(path.join(outDir, 'addresses.js'), content);
console.log('Created src/contracts/addresses.js with', Object.keys(abis).length, 'ABIs');
