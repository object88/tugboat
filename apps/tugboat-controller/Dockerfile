FROM gobuild:local AS BUILD

# FROM debian:buster-slim AS RELEASE
FROM scratch AS RELEASE

USER appuser

COPY --from=BUILD "/etc/ssl/certs/ca-certificates.crt" "/etc/ssl/certs/ca-certificates.crt"
COPY --from=BUILD "/opt/appuser/passwd" "/etc/passwd"
COPY --from=BUILD "/opt/appuser/group" "/etc/group"

CMD ["/usr/local/bin/tugboat-controller", "run", "--verbose"]

# Keep this late to minimize the number of layer changes.
COPY --from=BUILD "/go/src/github.com/object88/tugboat/bin/tugboat-controller-linux-amd64" "/usr/local/bin/tugboat-controller"
