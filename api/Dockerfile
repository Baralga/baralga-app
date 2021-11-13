FROM alpine
COPY baralga-api /usr/bin/baralga-api

ENV BARALGA_DB "postgres://postgres:postgres@localhost:5432/baralga"
ENV BARALGA_ENV "production"
ENV BARALGA_JWTSECRET "-my-secret-"

EXPOSE 8080

# Command to run when starting the container.
ENTRYPOINT ["/usr/bin/baralga-api"]
