// Simple script to verify our changes
const fs = require('fs');
const path = require('path');

// Check the table-utils.ts file
const tableUtilsPath = path.join('src', 'lib', 'table-utils.ts');
const tableUtilsContent = fs.readFileSync(tableUtilsPath, 'utf8');

console.log('Checking table-utils.ts...');
if (tableUtilsContent.includes('DEFAULT_PAGE_SIZE = 25')) {
  console.log('✅ DEFAULT_PAGE_SIZE is correctly set to 25');
} else {
  console.log('❌ DEFAULT_PAGE_SIZE is not set to 25');
}

// Check the Pagination.tsx file
const paginationPath = path.join('src', 'components', 'Pagination.tsx');
const paginationContent = fs.readFileSync(paginationPath, 'utf8');

console.log('\nChecking Pagination.tsx...');
if (paginationContent.includes('[10, 25, 50, 100].map')) {
  console.log('✅ Pagination options are correctly set to [10, 25, 50, 100]');
} else {
  console.log('❌ Pagination options are not set to [10, 25, 50, 100]');
}

// Check a few table components to ensure they use the DEFAULT_PAGE_SIZE
const tableComponents = [
  'src/components/WorkspaceTable.tsx',
  'src/components/ImageTable.tsx',
  'src/components/RegistryTable.tsx',
  'src/components/ApiKeyTable.tsx'
];

console.log('\nChecking table components...');
tableComponents.forEach(tablePath => {
  try {
    const tableContent = fs.readFileSync(tablePath, 'utf8');
    const fileName = path.basename(tablePath);
    
    if (tableContent.includes('import { DEFAULT_PAGE_SIZE } from') && 
        tableContent.includes('pagination: {') && 
        tableContent.includes('pageSize: DEFAULT_PAGE_SIZE')) {
      console.log(`✅ ${fileName} correctly imports and uses DEFAULT_PAGE_SIZE`);
    } else {
      console.log(`❌ ${fileName} does not correctly use DEFAULT_PAGE_SIZE`);
    }
  } catch (error) {
    console.log(`❌ Error reading ${tablePath}: ${error.message}`);
  }
});
