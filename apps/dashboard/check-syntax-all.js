// Script to check for syntax errors in all modified files
const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

// List of files to check
const filesToCheck = [
  'src/lib/table-utils.ts',
  'src/components/Pagination.tsx',
  'src/components/WorkspaceTable.tsx',
  'src/components/ImageTable.tsx',
  'src/components/RegistryTable.tsx',
  'src/components/ApiKeyTable.tsx',
  'src/components/OrganizationMembers/OrganizationMemberTable.tsx',
  'src/components/OrganizationMembers/OrganizationInvitationTable.tsx',
  'src/components/UserOrganizationInvitations/UserOrganizationInvitationTable.tsx',
  'src/components/OrganizationRoles/OrganizationRoleTable.tsx',
  'src/lib/table-utils.test.ts',
  'src/components/Pagination.test.tsx'
];

// Check each file
console.log('Checking files for syntax errors...');
let allFilesValid = true;

filesToCheck.forEach(filePath => {
  try {
    // Read the file
    const content = fs.readFileSync(filePath, 'utf8');
    
    // Check for basic syntax issues
    const fileName = path.basename(filePath);
    
    // Check for missing imports in table components
    if (filePath !== 'src/lib/table-utils.ts' && 
        filePath !== 'src/components/Pagination.tsx' && 
        filePath !== 'src/lib/table-utils.test.ts' && 
        filePath !== 'src/components/Pagination.test.tsx' && 
        !content.includes('import { DEFAULT_PAGE_SIZE }')) {
      console.log(`❌ ${fileName}: Missing import for DEFAULT_PAGE_SIZE`);
      allFilesValid = false;
    }
    
    // Check for initialState configuration in table components
    if (filePath !== 'src/lib/table-utils.ts' && 
        filePath !== 'src/components/Pagination.tsx' && 
        filePath !== 'src/lib/table-utils.test.ts' && 
        filePath !== 'src/components/Pagination.test.tsx' && 
        (!content.includes('initialState: {') || !content.includes('pageSize: DEFAULT_PAGE_SIZE'))) {
      console.log(`❌ ${fileName}: Missing initialState configuration with DEFAULT_PAGE_SIZE`);
      allFilesValid = false;
    }
    
    // Check for pagination options in Pagination.tsx
    if (filePath === 'src/components/Pagination.tsx' && !content.includes('[10, 25, 50, 100].map')) {
      console.log(`❌ ${fileName}: Missing pagination options [10, 25, 50, 100]`);
      allFilesValid = false;
    }
    
    // Check for DEFAULT_PAGE_SIZE value in table-utils.ts
    if (filePath === 'src/lib/table-utils.ts' && !content.includes('DEFAULT_PAGE_SIZE = 25')) {
      console.log(`❌ ${fileName}: DEFAULT_PAGE_SIZE is not set to 25`);
      allFilesValid = false;
    }
    
    // Check for basic syntax errors
    try {
      // Use Node.js to check for syntax errors
      const syntaxCheck = `
        const content = \`${content.replace(/`/g, '\\`')}\`;
        try {
          new Function(content);
          process.exit(0);
        } catch (error) {
          console.error(error);
          process.exit(1);
        }
      `;
      
      // Write to a temporary file and execute
      const tempFile = `temp-${Math.random().toString(36).substring(7)}.js`;
      fs.writeFileSync(tempFile, syntaxCheck);
      execSync(`node ${tempFile}`);
      fs.unlinkSync(tempFile);
      
      console.log(`✅ ${fileName}: No syntax issues found`);
    } catch (error) {
      console.log(`❌ ${fileName}: Syntax error detected`);
      allFilesValid = false;
    }
  } catch (error) {
    console.log(`❌ Error reading ${filePath}: ${error.message}`);
    allFilesValid = false;
  }
});

if (allFilesValid) {
  console.log('\n✅ All files passed the syntax check!');
} else {
  console.log('\n❌ Some files have syntax issues. Please fix them before proceeding.');
}
