FROM gobuild:local AS build

# FROM debian:buster AS RELEASE
FROM scratch AS RELEASE

USER appuser

COPY --from=build "/etc/ssl/certs/ca-certificates.crt" "/etc/ssl/certs/ca-certificates.crt"
COPY --from=build "/opt/appuser/*" "/etc/"
COPY --chown=appuser:appuser --from=build "/home/appuser" "/home/appuser"

CMD ["/usr/local/bin/tugboat-notifier-slack", "run", "--verbose"]

# Keep this late to minimize the number of layer changes.
COPY --from=build "/go/src/github.com/object88/tugboat/bin/tugboat-notifier-slack-linux-amd64" "/usr/local/bin/tugboat-notifier-slack"
