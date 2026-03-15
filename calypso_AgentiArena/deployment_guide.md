# AgentArena Ultimate Deployment Guide

Since you have already deployed your **AI Agents (Backend)** to Render, you are 1/3 of the way done! Here are the exact steps to deploy the **Smart Contracts** and the **Frontend Web App**.

---

## Part 1: Deploy Smart Contracts to HeLa Testnet

You don't need Remix for this! Your codebase is already configured to deploy directly to the HeLa Testnet using **Hardhat**.

### Step 1: Set up your Wallet
1. Open the `.env` file located in `d:\hackjklu\.env`.
2. Add your MetaMask Private Key to this file. **Make sure your wallet has testnet HL tokens to pay for gas!**
   \`\`\`env
   PRIVATE_KEY="your_private_key_here"
   HELA_RPC_URL="https://testnet-rpc.helachain.com"
   \`\`\`

### Step 2: Deploy from Terminal
1. Open a new terminal window inside the `d:\hackjklu` folder.
2. Run this command to deploy your contracts directly to the blockchain:
   \`\`\`bash
   npx hardhat run scripts/deploy.js --network hela
   \`\`\`
3. Wait for the command to finish. The terminal will output a list of **contract addresses** (e.g., `TaskLedger deployed to: 0xABCD...`). Save these addresses!

### Step 3: Tell Your Frontend about the New Contracts
1. Open `d:\hackjklu\src\contracts\addresses.js`.
2. Replace the placeholder contract addresses in that file with the real ones you just generated in the terminal.

---

## Part 2: Connect the Frontend to Your Render Agents

The frontend needs to know where your Live Render agents are located!

1. Open `d:\hackjklu\src\data\agents.js`.
2. Look through the list of agents in that file.
3. For each agent you want to showcase, change its `endpointUrl` to the live URL Render gave you.
   *Example:*
   \`\`\`javascript
   {
       name: "Crypto Tax Reporter",
       category: "finance",
       endpointUrl: "https://your-tax-render-url.onrender.com/api/v1/execute",
       // ...
   }
   \`\`\`

---

## Part 3: Deploy the Frontend to Vercel (The Cloud Server)

Vercel is the easiest, most professional way to host a React/Vite app for free.

### Step 1: Push Your Final Code to GitHub
Now that your contract addresses and Render URLs are saved in the code, push the changes to GitHub.
1. Open your terminal in `d:\hackjklu`.
2. Run:
   \`\`\`bash
   git add .
   git commit -m "update addresses and agent URLs for production"
   git push origin main
   \`\`\`

### Step 2: Deploy on Vercel
1. Go to [Vercel.com](https://vercel.com/) and click **Sign Up** (use your GitHub account).
2. Click **Add New...** -> **Project**.
3. Choose the `hackjklu` GitHub repository from the list and click **Import**.
4. Important step: Click on the **Environment Variables** tab to expand it. Add this specific variable:
   * **Name:** `VITE_DEMO_MODE`
   * **Value:** `false`
   *(This ensures the live site uses real blockchain transactions instead of mocks!)*
5. Click the big **Deploy** button.

### Conclusion
Wait about 2 minutes. Vercel will process your files and grant you a live, public URL (like `https://agent-arena-tau.vercel.app`).

You can now send that link to **anyone**. When they click around, it will talk to your real contracts on HeLa and your real agents on Render!
