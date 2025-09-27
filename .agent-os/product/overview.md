# Baralga - Multi-User Time Tracking Application

## Product Vision

Baralga is a comprehensive multi-user time tracking application designed for teams and organizations to track time spent on projects and activities. It provides both a modern web interface and a REST API for programmatic access, with robust authentication, reporting, and export capabilities.

## Target Users

- **Small to Medium Teams**: Organizations needing structured time tracking
- **Project Managers**: Users who need detailed reporting and project oversight
- **Individual Contributors**: Team members tracking their work activities
- **Administrators**: Users managing organizations, projects, and user access

## Core Value Proposition

- **Multi-tenant Architecture**: Organization-based isolation ensures data security
- **Dual Interface**: Both web UI and REST API for maximum flexibility
- **Comprehensive Reporting**: Multiple report types with export capabilities
- **Modern Tech Stack**: Built with Go, PostgreSQL, and modern web technologies
- **Security-First**: JWT authentication, CSRF protection, and role-based access control

## Key Differentiators

1. **Organization-based Multi-tenancy**: Complete data isolation between organizations
2. **Role-based Access Control**: Granular permissions (User/Admin roles)
3. **OAuth Integration**: Seamless login with GitHub and Google
4. **Advanced Reporting**: Time-based, project-based, and tag-based analytics
5. **Export Capabilities**: CSV and Excel export for data portability
6. **Modern UI**: Server-side rendered HTML with HTMX for responsive interactions
