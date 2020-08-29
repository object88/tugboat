FROM gobuild:local AS BUILD

# FROM debian:buster-slim AS RELEASE
FROM scratch AS RELEASE

USER appuser

COPY --from=BUILD "/etc/ssl/certs/ca-certificates.crt" "/etc/ssl/certs/ca-certificates.crt"
COPY --from=BUILD "/opt/appuser/*" "/etc/"

CMD ["/usr/local/bin/tugboat-watcher", "run", "--verbose"]

# Keep this late to minimize the number of layer changes.
COPY --from=BUILD "/go/src/github.com/object88/tugboat/bin/tugboat-watcher-linux-amd64" "/usr/local/bin/tugboat-watcher"
