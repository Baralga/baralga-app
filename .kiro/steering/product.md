# Product Overview

Baralga is a multi-user time tracking application with both web frontend and REST API capabilities. It allows users to track time spent on projects and activities, generate reports, and manage user authentication with role-based access control.

## Key Features
- Time tracking for activities with project association
- Multi-user support with organization-based isolation
- Role-based access control (User/Admin roles)
- Web interface with keyboard shortcuts for productivity
- REST API for programmatic access
- Reporting capabilities (daily, weekly, monthly, quarterly)
- OAuth integration (GitHub, Google)
- Tag support for activities
- Export functionality (CSV, Excel)

## User Roles
- **User** (`ROLE_USER`): Full access to own activities, read-only access to projects
- **Admin** (`ROLE_ADMIN`): Full access to all users' activities and projects

## Authentication
- JWT-based authentication with configurable expiry
- CSRF protection for web interface
- OAuth providers: GitHub and Google
- BCrypt password hashing (version $2a, strength 10)