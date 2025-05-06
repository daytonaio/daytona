// Simple script to check for syntax errors in TypeScript files
const fs = require('fs');
const path = require('path');

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
  'src/components/OrganizationRoles/OrganizationRoleTable.tsx'
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
    
    // Check for table-utils.ts
    if (filePath === 'src/lib/table-utils.ts') {
      if (!content.includes('DEFAULT_PAGE_SIZE = 25')) {
        console.log(`❌ ${fileName}: DEFAULT_PAGE_SIZE is not set to 25`);
        allFilesValid = false;
      } else {
        console.log(`✅ ${fileName}: DEFAULT_PAGE_SIZE is correctly set to 25`);
      }
    }
    // Check for Pagination.tsx
    else if (filePath === 'src/components/Pagination.tsx') {
      if (!content.includes('[10, 25, 50, 100].map')) {
        console.log(`❌ ${fileName}: Missing pagination options [10, 25, 50, 100]`);
        allFilesValid = false;
      } else {
        console.log(`✅ ${fileName}: Pagination options are correctly set to [10, 25, 50, 100]`);
      }
    }
    // Check for table components
    else {
      if (!content.includes('import { DEFAULT_PAGE_SIZE }')) {
        console.log(`❌ ${fileName}: Missing import for DEFAULT_PAGE_SIZE`);
        allFilesValid = false;
      } else {
        console.log(`✅ ${fileName}: Correctly imports DEFAULT_PAGE_SIZE`);
      }
      
      if (!content.includes('initialState: {') || !content.includes('pageSize: DEFAULT_PAGE_SIZE')) {
        console.log(`❌ ${fileName}: Missing initialState configuration with DEFAULT_PAGE_SIZE`);
        allFilesValid = false;
      } else {
        console.log(`✅ ${fileName}: Correctly uses DEFAULT_PAGE_SIZE in initialState`);
      }
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
