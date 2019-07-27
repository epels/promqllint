# promqllint

A web-based linter for the PromQL (Prometheus Query Language). Hosted on [promqllint.com](https://promqllint.com).

## Requirements

* Go 1.12

## Deployment
Although promqllint can be deployed anywhere, it comes with a basic Google App Engine configuration out of the box. After authorizing with a service account or user account account, it's easily deployed using the gcloud SDK:

```bash
gcloud app deploy
```
