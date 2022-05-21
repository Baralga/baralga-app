# Baralga 

Multi user time tracking application with web frontend and API.

## User Guide

### Keyboard Shortcuts

#### Track Activities

| Shortcut                         | Action          |
| -------------------------------- |:----------------|
| <kbd>Alt</kbd> + <kbd>Shift</kbd> + <kbd>n</kbd>  | Add Activity    |
| <kbd>Alt</kbd> + <kbd>Shift</kbd> + <kbd>p</kbd>  | Manage Projects |

#### Report Activities

| Shortcut                                  | Action                 |
| ----------------------------------------- |:-----------------------|
| <kbd>Shift</kbd> + <kbd>Arrow Left</kbd>  | Show previous Timespan |
| <kbd>Shift</kbd> + <kbd>Arrow Down</kbd>  | Show current Timespan  |
| <kbd>Shift</kbd> + <kbd>Arrow Right</kbd> | Show next Timespan     |

## Administration

### Accessing the Web User Interface

The web user interface is available at `http://localhost:8080/`. You can log in as administrator with `admin/adm1n` or as user with `user1/us3r`.

### Configuration

The backend is configured using the following environment variables:

| Environment Variable  | Default Value                        | Description  |
| --------------------- |:------------------------------------| :--------|
| `BARALGA_DB`      | `postgres://postgres:postgres@localhost:5432/baralga`| PostgreSQL Connection string for database |
| `BARALGA_DBMAXCONNS`      | `3`| Maximum number of database connections in pool. |
| `PORT` | `8080`      |    http server port |
| `BARALGA_WEBROOT` | `http://localhost:8080`      |    Web server root |
| `BARALGA_JWTSECRET` | `secret`      |    Random secret for JWT generation |
| `BARALGA_CSRFSECRET` | `CSRFsecret`      |    Random secret for CSRF protection |
| `BARALGA_ENV` | `dev`      |    use `production` for production mode |
| `BARALGA_SMTPSERVERNAME` | `smtp.server:465`      |    Host and port of your SMTP server |
| `BARALGA_SMTPFROM` | `smtp.from@baralga.com`      |    From email for your SMTP server |
| `BARALGA_SMTPUSER` | `smtp.user@baralga.com`      |    User for your SMTP server |
| `BARALGA_SMTPPASSWORD` | `SMTPPassword`      |    Password for your SMTP server |
| `BARALGA_DATAPROTECTIONURL` | `#`      |   URL to data protection rules. |
| `BARALGA_GITHUBCLIENTID` | `GithubClientID`      |    OAuth Client ID for Github. |
| `BARALGA_GITHUBCLIENTSECRET` | `GithubClientSecret`      |    OAuth Client Secret for Github. |
| `BARALGA_GITHUBREDIRECTURL` | `http://localhost:8080/github/callback`      |    OAuth Redirect URL for Github. |

### Users and Roles

Baralga supports the following roles:

| Role  | DB Name | Description                        |
| ----- |:------- |:------------------------------------|
| User  | `ROLE_USER` |Full access to his own activities but can only read projects. |
| Admin | `ROLE_ADMIN`  | Full access to activities of all users and projects.          |

Passwords are encoded in BCrypt with BCrypt version `$2a` and strength 10. The tool https://8gwifi.org/bccrypt.jsp
can be used to create a hashed password to be used in sql.

### Database

* [PostgreSQL](https://www.postgresql.org/)

#### PostgreSQL Configuration
```bash
BARALGA_DB=postgres://postgres:postgres@localhost:5432/baralga
```
                         
### Health Check

A health check is available at `http://localhost:8080/health`.

## Development

You can launch the application in VSCode with your custom environment variables. For that
create a file `.env` in the root of this repository with the following content:

```
PORT=8080
BARALGA_JWTSECRET=my***secret
BARALGA_SMTPSERVERNAME=mysmtp.host.com:465
BARALGA_SMTPUSER=baralga@mydomail.com
BARALGA_SMTPPASSWORD=mysmtp***secret
BARALGA_DATAPROTECTIONURL=https://localhost:8080/dataprotection/
BARALGA_GITHUBCLIENTID=my***clientid
BARALGA_GITHUBCLIENTSECRET=my***secret
BARALGA_GITHUBREDIRECTURL=https://localhost:8080/github/callback
```

## Open Topics
* Add password forgot support