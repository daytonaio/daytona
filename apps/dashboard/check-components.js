// Script to check components
const fs = require('fs');
const path = require('path');

// List of files to check
const filesToCheck = [
  'src/components/OrganizationMembers/OrganizationMemberTable.tsx',
  'src/components/OrganizationMembers/OrganizationInvitationTable.tsx',
  'src/components/OrganizationRoles/OrganizationRoleTable.tsx',
  'src/components/UserOrganizationInvitations/UserOrganizationInvitationTable.tsx'
];

// Check each file
console.log('Checking components...');
let allFilesValid = true;

filesToCheck.forEach(filePath => {
  try {
    // Read the file
    const content = fs.readFileSync(filePath, 'utf8');
    
    // Check for basic syntax issues
    const fileName = path.basename(filePath);
    
    // Check for imports
    if (!content.includes('import { DEFAULT_PAGE_SIZE }')) {
      console.log(`❌ ${fileName}: Missing import for DEFAULT_PAGE_SIZE`);
      allFilesValid = false;
    } else {
      console.log(`✅ ${fileName}: Correctly imports DEFAULT_PAGE_SIZE`);
    }
    
    // Check for initialState configuration
    if (!content.includes('initialState: {') || !content.includes('pageSize: DEFAULT_PAGE_SIZE')) {
      console.log(`❌ ${fileName}: Missing initialState configuration with DEFAULT_PAGE_SIZE`);
      allFilesValid = false;
    } else {
      console.log(`✅ ${fileName}: Correctly uses DEFAULT_PAGE_SIZE in initialState`);
    }
    
    // Check for Pagination component
    if (!content.includes('<Pagination table={table}')) {
      console.log(`❌ ${fileName}: Missing Pagination component`);
      allFilesValid = false;
    } else {
      console.log(`✅ ${fileName}: Correctly uses Pagination component`);
    }
  } catch (error) {
    console.log(`❌ Error reading ${filePath}: ${error.message}`);
    allFilesValid = false;
  }
});

if (allFilesValid) {
  console.log('\n✅ All components passed the check!');
} else {
  console.log('\n❌ Some components have issues. Please fix them before proceeding.');
}
