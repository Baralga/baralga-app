# Quickstart: Organization Management Dialog

## Overview
This quickstart demonstrates the organization management dialog functionality for administrators to update their organization title.

## Prerequisites
- User must be logged in as an administrator
- Organization must exist in the system
- User must belong to the organization

## User Journey

### 1. Access Organization Management
**Given** I am logged in as an organization administrator
**When** I navigate to my profile page
**Then** I should see an "Organization Settings" option in the profile menu

**Steps**:
1. Log in as an administrator user
2. Navigate to `/profile` page
3. Look for "Organization Settings" or "Manage Organization" link/button
4. Click on the organization management option

**Expected Result**: Organization management dialog opens

### 2. View Current Organization Title
**Given** I am on the organization management dialog
**When** the dialog loads
**Then** I should see the current organization title in the input field

**Steps**:
1. Dialog opens with current organization title pre-filled
2. Organization title field shows existing value
3. Form is ready for editing

**Expected Result**: Current organization title is displayed and editable

### 3. Update Organization Title
**Given** I am on the organization management dialog
**When** I change the organization title and click "Save Changes"
**Then** the organization title should be updated throughout the system

**Steps**:
1. Clear the existing organization title
2. Enter new organization title (e.g., "Acme Corporation Updated")
3. Click "Save Changes" button
4. Wait for success confirmation

**Expected Result**: 
- Success message displayed
- Dialog closes automatically
- Page refreshes to show updated organization title
- Organization title is updated across all relevant pages

### 4. Cancel Changes
**Given** I am on the organization management dialog
**When** I click "Cancel" or the X button
**Then** the dialog should close without saving changes

**Steps**:
1. Make changes to organization title
2. Click "Cancel" button or X button
3. Dialog closes

**Expected Result**: 
- Dialog closes without saving
- Original organization title remains unchanged
- No success/error messages displayed

### 5. Validation Error Handling
**Given** I am on the organization management dialog
**When** I submit an empty organization title
**Then** I should see validation error messages

**Steps**:
1. Clear the organization title field
2. Click "Save Changes" button
3. Observe validation errors

**Expected Result**:
- Validation error messages displayed
- Form remains open for correction
- Organization title is not updated

## Test Scenarios

### Scenario 1: Successful Organization Title Update
```
1. Login as admin user
2. Navigate to profile page
3. Click "Organization Settings"
4. Change organization title to "New Organization Name"
5. Click "Save Changes"
6. Verify success message
7. Verify dialog closes
8. Verify page shows updated title
```

### Scenario 2: Non-Admin User Access Attempt
```
1. Login as regular user (non-admin)
2. Navigate to profile page
3. Verify "Organization Settings" option is not visible
4. Try to access /profile/organization directly
5. Verify access denied message
```

### Scenario 3: Validation Error Handling
```
1. Login as admin user
2. Navigate to organization settings
3. Clear organization title field
4. Click "Save Changes"
5. Verify validation error messages
6. Enter valid title
7. Click "Save Changes"
8. Verify successful update
```

### Scenario 4: Duplicate Organization Title
```
1. Login as admin user
2. Navigate to organization settings
3. Enter organization title that already exists
4. Click "Save Changes"
5. Verify conflict error message
6. Enter unique title
7. Click "Save Changes"
8. Verify successful update
```

## Expected Behavior

### Success Flow
- Organization title updates immediately
- Success message displayed
- Dialog closes automatically
- Page refreshes to show changes
- All organization references updated

### Error Flows
- Validation errors displayed inline
- Form remains open for correction
- Clear error messages
- No partial updates

### Security
- Only administrators can access
- Non-admins see no organization management options
- Direct URL access blocked for non-admins
- CSRF protection on form submissions

## Performance Expectations
- Dialog opens within 200ms
- Form submission completes within 500ms
- No page reload required for dialog interaction
- Smooth user experience with HTMX

## Browser Compatibility
- Modern browsers with HTMX support
- Bootstrap modal functionality
- JavaScript enabled for HTMX interactions

## Implementation Notes

### Existing Infrastructure
- Organizations table already exists with org_id and title fields
- Organization domain model already exists
- Organization repository interface already exists
- Users already have organization_id field

### New Components Needed
- Organization service for business logic
- Organization web interface for user interaction
- Role-based authorization for admin access
- Organization title validation and uniqueness checks