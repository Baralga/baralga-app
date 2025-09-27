# Spec Requirements Document

> Spec: Organization Invite Links
> Created: 2025-09-27

## Overview

Implement a secure organization invite link system that allows administrators to generate time-limited invitation links for their organization. New users can register using these links and automatically join the organization with ROLE_USER permissions, streamlining the onboarding process for team members.

## User Stories

### Admin Invite Link Generation

As an organization administrator, I want to generate secure invite links for my organization, so that I can easily invite new team members without manually creating accounts for them.

**Detailed Workflow:**
1. Admin navigates to organization settings
2. Admin clicks "Generate Invite Link" button
3. System generates a unique, time-limited invite link (valid for 24 hours)
4. Admin can copy the link and share it via email, chat, or other means
5. Admin can view active invite links and their expiration times
6. Admin can revoke active invite links if needed

### User Registration via Invite Link

As a new user, I want to register for an organization using an invite link, so that I can quickly join a team and start tracking time without complex setup.

**Detailed Workflow:**
1. User clicks on invite link received from admin
2. System validates the invite link (checks expiration and validity)
3. User fills out registration form (name, username, email, password)
4. System creates user account with ROLE_USER in the specified organization
5. User is automatically logged in and redirected to the main dashboard
6. User can immediately start tracking time for organization projects

## Spec Scope

1. **Invite Link Generation** - Admin interface to create time-limited organization invite links
2. **Invite Link Validation** - Secure validation of invite links with expiration checking
3. **User Registration Flow** - Streamlined registration process for invited users
4. **Organization Assignment** - Automatic assignment of new users to the correct organization
5. **Role Management** - Automatic assignment of ROLE_USER role to invited users
6. **Invite Link Management** - Admin interface to view and revoke active invite links

## Out of Scope

- Bulk invite functionality (multiple users at once)
- Custom invite link expiration times (fixed at 24 hours)
- Email notifications for invite link generation
- Invite link usage analytics or tracking
- Custom role assignment during invite (all users get ROLE_USER)

## Expected Deliverable

1. Admin can generate organization invite links from the organization settings page
2. Generated invite links are valid for exactly 24 hours and automatically expire
3. New users can register using valid invite links and are automatically assigned to the correct organization with ROLE_USER role
4. Admin can view active invite links and their expiration times
5. Admin can revoke active invite links before they expire
6. Invalid or expired invite links show appropriate error messages to users
