# Running Tugboat in a development environment

## 3rd party tools

Local development will require using some 3rd party tools.

* [minikube](https://kubernetes.io/docs/tasks/tools/install-minikube/): used to host
* [Slack](https://api.slack.com/apps/): 
* [ngrok](https://ngrok.com): used to proxy requests from slack to your locally running applications


## Configure ngrok

See [online instructions](https://dashboard.ngrok.com/get-started/setup).  Note the console output:

``` sh
ngrok by @inconshreveable                                                                                                                                                                                                     (Ctrl+C to quit)

Session Status                online
Account                       Paul Brousseau (Plan: Free)
Version                       2.3.35
Region                        United States (us)
Web Interface                 http://127.0.0.1:4040
Forwarding                    http://abcdef123456.ngrok.io -> http://localhost:80
Forwarding                    https://abcdef123456.ngrok.io -> http://localhost:80

Connections                   ttl     opn     rt1     rt5     p50     p90
                              0       0       0.00    0.00    0.00    0.00
```

## Slack

Once the `tugboat-slack` app is running and `ngrok` is proxying, the Slack application can be set up.