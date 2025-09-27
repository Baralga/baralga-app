# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-09-27-organization-invite-links/spec.md

## Technical Requirements

- **Invite Link Generation**: Secure token-based invite links using UUID with 24-hour expiration
- **Database Schema**: New `organization_invites` table to store invite links with expiration tracking
- **Admin Interface**: HTMX-powered interface for generating and managing invite links in organization settings
- **Registration Flow**: Modified user registration to accept invite tokens and auto-assign organization
- **Security**: Invite links must be cryptographically secure and tamper-proof
- **Validation**: Server-side validation of invite links with proper error handling
- **User Experience**: Seamless registration flow with clear success/error messaging
- **Role Assignment**: Automatic ROLE_USER assignment for invited users
- **Link Management**: Admin interface to view active invites and revoke them
- **Expiration Handling**: Automatic cleanup of expired invite links
- **URL Structure**: Clean invite URLs like `/invite/{token}` for user-friendly sharing

## External Dependencies

No new external dependencies are required. The implementation will use existing Go standard library packages and the current tech stack (Go, PostgreSQL, HTMX, Bootstrap).
