// Dynamic import for CommonJS module
const { Daytona } = await import('@daytonaio/sdk');
  
// Initialize the Daytona client
const daytona = new Daytona({ apiKey: 'dtn_292ca9aa6c96cd08230b6cfa050da7610482c6cfdec7f282472fa5db48589c6b' });

// Create the Sandbox instance
const sandbox = await daytona.create({
  language: 'typescript',
});

// Run the code securely inside the Sandbox
const response = await sandbox.process.codeRun('console.log("Hello World from code!")')
console.log(response.result);
  