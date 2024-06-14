FROM golang:1.22.4 AS build

# Create user
RUN useradd -m --uid 1000 build
USER 1000

WORKDIR /app

# Build
COPY --chown=1000:1000 . .
RUN --mount=type=cache,uid=1000,gid=1000,target=/home/build/.cache/go-build CGO_ENABLED=0 go build

# prepare /etc/passwd for scratch image
RUN echo "nobody:*:65534:65534:nobody:/_nonexistent:/bin/false" > /tmp/etc_passwd

# Scratch image
FROM scratch

# Switch user
COPY --from=0 /tmp/etc_passwd /etc/passwd
USER nobody

# Copy CA certificates
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary to scratch image
COPY --from=build /app/anythingispwa /

# Copy assets
COPY templates/ /templates

EXPOSE 8080/tcp

CMD ["/anythingispwa"]
