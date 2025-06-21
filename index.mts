// Dynamic import for CommonJS module
const { Daytona } = await import('@daytonaio/sdk');
  
// Initialize the Daytona client
const daytona = new Daytona({ apiKey: process.env.DAYTONA_API_KEY });

// Create the Sandbox instance
const sandbox = await daytona.create({
  language: 'typescript',
});

// Run the code securely inside the Sandbox
const response = await sandbox.process.codeRun('console.log("Hello World from code!")')
console.log(response.result);
  