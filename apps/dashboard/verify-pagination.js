// Simple script to verify the pagination implementation
const fs = require('fs');
const path = require('path');

// Check the table-utils.ts file
console.log('Checking table-utils.ts...');
const tableUtilsPath = path.join('src', 'lib', 'table-utils.ts');
const tableUtilsContent = fs.readFileSync(tableUtilsPath, 'utf8');

if (tableUtilsContent.includes('export const DEFAULT_PAGE_SIZE = 25')) {
  console.log('✅ DEFAULT_PAGE_SIZE is correctly set to 25');
} else {
  console.log('❌ DEFAULT_PAGE_SIZE is not set to 25');
}

// Check the Pagination.tsx file
console.log('\nChecking Pagination.tsx...');
const paginationPath = path.join('src', 'components', 'Pagination.tsx');
const paginationContent = fs.readFileSync(paginationPath, 'utf8');

if (paginationContent.includes('[10, 25, 50, 100].map')) {
  console.log('✅ Pagination options are correctly set to [10, 25, 50, 100]');
} else {
  console.log('❌ Pagination options are not set to [10, 25, 50, 100]');
}

// Check table components
const tableComponents = [
  { name: 'WorkspaceTable', path: path.join('src', 'components', 'WorkspaceTable.tsx') },
  { name: 'ImageTable', path: path.join('src', 'components', 'ImageTable.tsx') },
  { name: 'RegistryTable', path: path.join('src', 'components', 'RegistryTable.tsx') },
  { name: 'ApiKeyTable', path: path.join('src', 'components', 'ApiKeyTable.tsx') },
  { name: 'OrganizationMemberTable', path: path.join('src', 'components', 'OrganizationMembers', 'OrganizationMemberTable.tsx') },
  { name: 'OrganizationInvitationTable', path: path.join('src', 'components', 'OrganizationMembers', 'OrganizationInvitationTable.tsx') },
  { name: 'UserOrganizationInvitationTable', path: path.join('src', 'components', 'UserOrganizationInvitations', 'UserOrganizationInvitationTable.tsx') },
  { name: 'OrganizationRoleTable', path: path.join('src', 'components', 'OrganizationRoles', 'OrganizationRoleTable.tsx') }
];

console.log('\nChecking table components...');
tableComponents.forEach(component => {
  try {
    const content = fs.readFileSync(component.path, 'utf8');
    
    // Check for import
    const hasImport = content.includes('import { DEFAULT_PAGE_SIZE } from');
    
    // Check for initialState
    const hasInitialState = content.includes('initialState: {') && 
                           content.includes('pagination: {') && 
                           content.includes('pageSize: DEFAULT_PAGE_SIZE');
    
    // Check for Pagination component
    const hasPagination = content.includes('<Pagination table={table}');
    
    if (hasImport && hasInitialState && hasPagination) {
      console.log(`✅ ${component.name}: All checks passed`);
    } else {
      console.log(`❌ ${component.name}: Some checks failed`);
      if (!hasImport) console.log(`  - Missing import for DEFAULT_PAGE_SIZE`);
      if (!hasInitialState) console.log(`  - Missing initialState configuration with DEFAULT_PAGE_SIZE`);
      if (!hasPagination) console.log(`  - Missing Pagination component`);
    }
  } catch (error) {
    console.log(`❌ Error reading ${component.name}: ${error.message}`);
  }
});

console.log('\nVerification complete!');
