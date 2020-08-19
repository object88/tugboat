# Slack configuration

The Slack integration requires a set of permissions and endpoints.

Note that this documentation should not be considered authoritive over Slack's own documentation. This is guidance in getting Tugboat integrated into 

Create OAuth bot scopes:
* `app_mentions:read`
* `chat:write`
* `commands`

Install to workspace
Note the "Bot User OAuth Access Token"
Note the "Signing Secret" and "Verification Token" from "Basic Information"

* `$HOST/v1/api/commands`
* `$HOST/v1/api/events`
* `$HOST/v1/api/interactive`