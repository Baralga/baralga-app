# Baralga 

Multi user time tracking application with web frontend and API.

## Administration

### Accessing the Web User Interface

The web user interface is available at `http://localhost:8080/`. You can log in as administrator with `admin/adm1n` or as user with `user1/us3r`.

### Configuration

The backend is configured using the following environment variables:

| Environment Variable  | Default Value                        | Description  |
| --------------------- |:------------------------------------| :--------|
| `BARALGA_DB`      | `postgres://postgres:postgres@localhost:5432/baralga`| PostgreSQL Connection string for database |
| `PORT` | `8080`      |    http server port |
| `BARALGA_JWTSECRET` | `secret`      |    Random secret for JWT generation |
| `BARALGA_ENV` | `dev`      |    use `production` for production mode |


### Users and Roles

Baralga supports the following roles:

| Role  | DB Name | Description                        |
| ----- |:------- |:------------------------------------|
| User  | `ROLE_USER` |Full access to his own activities but can only read projects. |
| Admin | `ROLE_ADMIN`  | Full access to activities of all users and projects.          |


#### Administration

### Database

* [PostgreSQL](https://www.postgresql.org/)

#### PostgreSQL Configuration
```bash
BARALGA_DB=postgres://postgres:postgres@localhost:5432/baralga
```
                         
### Health Check

A health check is available at `http://localhost:8080/health`.