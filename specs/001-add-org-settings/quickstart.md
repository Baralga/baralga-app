# Quickstart Guide: Organization Management

## Overview
This guide demonstrates how to test the organization management functionality. The feature allows admin users to update their organization name through a modal dialog accessible from their profile page.

## Prerequisites
- Baralga application running locally
- Admin user account (username: `admin`, password: `adm1n`)
- Regular user account (username: `user1`, password: `us3r`)

## Test Scenarios

### Scenario 1: Admin User Can Access Organization Management
**Given**: Admin user is logged in  
**When**: Admin navigates to profile page  
**Then**: Admin should see "Organization Settings" option

**Steps**:
1. Navigate to `http://localhost:8080`
2. Login with admin credentials (admin/adm1n)
3. Click on user profile/avatar in navigation
4. Verify "Organization Settings" or "Manage Organization" option is visible
5. Click on organization settings option
6. Verify modal dialog opens with current organization name

**Expected Result**: Modal dialog displays with organization name field and save/cancel buttons

### Scenario 2: Admin User Can Update Organization Name
**Given**: Admin user has opened organization management dialog  
**When**: Admin changes organization name and saves  
**Then**: Organization name should be updated successfully

**Steps**:
1. Follow steps from Scenario 1 to open organization dialog
2. Clear current organization name field
3. Enter new organization name: "Updated Organization Name"
4. Click "Save Changes" button
5. Verify success message appears
6. Close modal and reopen to verify name was saved

**Expected Result**: Organization name is updated and success message is displayed

### Scenario 3: Organization Name Validation
**Given**: Admin user has opened organization management dialog  
**When**: Admin enters invalid organization name  
**Then**: Validation error should be displayed

**Steps**:
1. Follow steps from Scenario 1 to open organization dialog
2. Enter invalid name: "AB" (too short)
3. Click "Save Changes" button
4. Verify validation error message appears
5. Enter valid name: "Valid Organization Name"
6. Click "Save Changes" button
7. Verify success message appears

**Expected Result**: Validation errors are displayed for invalid input, success for valid input

### Scenario 4: Non-Admin User Cannot Access Organization Management
**Given**: Regular user is logged in  
**When**: Regular user navigates to profile page  
**Then**: Organization management option should not be visible

**Steps**:
1. Navigate to `http://localhost:8080`
2. Login with regular user credentials (user1/us3r)
3. Click on user profile/avatar in navigation
4. Verify "Organization Settings" option is NOT visible
5. Try to access `/profile/organization` directly
6. Verify 403 Forbidden error or redirect

**Expected Result**: Non-admin users cannot access organization management features

### Scenario 5: CSRF Protection
**Given**: Admin user has opened organization management dialog  
**When**: Admin submits form without CSRF token  
**Then**: Request should be rejected

**Steps**:
1. Follow steps from Scenario 1 to open organization dialog
2. Open browser developer tools
3. Remove CSRF token from form
4. Submit form
5. Verify request is rejected with appropriate error

**Expected Result**: CSRF protection prevents unauthorized form submissions

## Manual Testing Checklist

### Functional Testing
- [ ] Admin can access organization management dialog
- [ ] Admin can update organization name
- [ ] Organization name validation works correctly
- [ ] Success messages are displayed after updates
- [ ] Non-admin users cannot access organization management
- [ ] CSRF protection is working
- [ ] Form preserves state on validation errors

### UI/UX Testing
- [ ] Modal dialog opens and closes correctly
- [ ] Form fields are properly labeled
- [ ] Validation errors are clearly displayed
- [ ] Success messages are user-friendly
- [ ] Cancel button closes modal without saving
- [ ] Save button updates organization name

### Security Testing
- [ ] Authorization checks prevent unauthorized access
- [ ] CSRF tokens are validated
- [ ] Input sanitization prevents XSS
- [ ] SQL injection protection is in place

### Error Handling Testing
- [ ] Network errors are handled gracefully
- [ ] Database errors show user-friendly messages
- [ ] Validation errors are field-specific
- [ ] Authorization errors redirect appropriately

## API Testing

### Test Organization Management Endpoints
```bash
# Test GET endpoint (should return modal HTML)
curl -X GET http://localhost:8080/profile/organization \
  -H "Cookie: session-cookie" \
  -H "X-Requested-With: XMLHttpRequest"

# Test POST endpoint (should update organization)
curl -X POST http://localhost:8080/profile/organization \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -H "Cookie: session-cookie" \
  -d "Name=Test Organization&CSRFToken=valid-token"
```

### Expected Responses
- **GET**: Returns HTML modal content
- **POST**: Returns success message or validation errors
- **403**: Returns forbidden message for non-admin users
- **401**: Returns unauthorized message for non-logged-in users

## Database Verification

### Check Organization Updates
```sql
-- Check current organization name
SELECT org_id, title FROM organizations WHERE org_id = '4ed0c11d-3d6a-41c1-9873-558e86084591';

-- Verify organization name was updated
SELECT org_id, title, updated_at FROM organizations WHERE org_id = '4ed0c11d-3d6a-41c1-9873-558e86084591';
```

### Expected Results
- Organization name should be updated in database
- Update timestamp should reflect recent changes
- No other organization data should be affected

## Performance Testing

### Response Time Requirements
- Modal dialog should load within 200ms
- Form submission should complete within 500ms
- Database updates should complete within 100ms

### Load Testing
- Test with multiple admin users updating organization simultaneously
- Verify no data corruption or race conditions
- Check for proper error handling under load

## Troubleshooting

### Common Issues
1. **Modal not opening**: Check HTMX integration and Bootstrap modal initialization
2. **Validation errors not showing**: Verify form validation and error display logic
3. **Authorization failures**: Check role-based access control implementation
4. **CSRF errors**: Verify token generation and validation

### Debug Steps
1. Check browser console for JavaScript errors
2. Verify network requests in browser developer tools
3. Check server logs for error messages
4. Verify database connection and query execution
5. Test with different user roles and permissions
