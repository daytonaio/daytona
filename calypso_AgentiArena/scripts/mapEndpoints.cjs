const fs = require('fs');
const path = require('path');

const agentsDir = path.join(__dirname, '..', 'server', 'agents');
const categories = fs.readdirSync(agentsDir).filter(f => fs.statSync(path.join(agentsDir, f)).isDirectory());

const map = {};

categories.forEach(category => {
  const filePath = path.join(agentsDir, category, 'index.js');
  if (fs.existsSync(filePath)) {
    const content = fs.readFileSync(filePath, 'utf8');
    
    // Match router.post('/route', payPerCall(ID,
    const regex = /router\.post\('([^']+)',\s*payPerCall\((\d+)/g;
    let match;
    while ((match = regex.exec(content)) !== null) {
      const route = match[1];
      const id = parseInt(match[2]);
      map[id] = `/${category}${route}`;
    }
  }
});

const outPath = path.join(__dirname, 'agentRoutes.json');
fs.writeFileSync(outPath, JSON.stringify(map, null, 2));
console.log('Wrote to', outPath);
