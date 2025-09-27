# Organization Invite Configuration

## Local Development

For local development, the application uses the default configuration:

```bash
# Default values in shared/config.go
Webroot: "http://localhost:8080"
```

This means invite links will be generated as:
```
http://localhost:8080/signup/invite/{token}
```

The invite page also supports social logins:
```
http://localhost:8080/github/login/invite/{token}
http://localhost:8080/google/login/invite/{token}
```

## Production Configuration

For production deployment, set the `WEBROOT` environment variable to your production domain:

```bash
export WEBROOT="https://your-domain.com"
```

Or set it in your deployment configuration:

```yaml
# docker-compose.yml example
environment:
  - WEBROOT=https://your-domain.com
```

This will generate invite links as:
```
https://your-domain.com/signup/invite/{token}
```

And social login invite URLs as:
```
https://your-domain.com/github/login/invite/{token}
https://your-domain.com/google/login/invite/{token}
```

## Environment Variables

The following environment variables control the invite system:

- `WEBROOT`: Base URL for the application (default: `http://localhost:8080`)
- `ENV`: Environment setting (default: `dev`, set to `production` for production)

## Example Production Setup

```bash
# Production environment variables
export WEBROOT="https://baralga.yourcompany.com"
export ENV="production"
export DB="postgres://user:password@db-host:5432/baralga"
export JWT_SECRET="your-secure-jwt-secret"
export CSRF_SECRET="your-secure-csrf-secret"
```

## Testing Invite Links

To test invite links locally:

1. Start the application: `go run main.go`
2. Navigate to `http://localhost:8080`
3. Login as an admin user
4. Go to Organization Settings
5. Generate an invite link
6. Copy the generated link (e.g., `http://localhost:8080/signup/invite/abc123`)
7. Open the link in a new browser/incognito window
8. **Option A**: Complete the registration form manually
9. **Option B**: Use social login buttons (GitHub/Google) for quick registration
10. The new user will be added to your organization with `ROLE_USER` permissions

### Social Login Testing

For social login testing, you can also use these direct URLs:
- `http://localhost:8080/github/login/invite/{token}`
- `http://localhost:8080/google/login/invite/{token}`

These will redirect users directly to the OAuth provider with the invite token preserved.
