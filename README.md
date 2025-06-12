# typesense-healthcheck

A lightweight Go service that continuously checks the health of a Typesense cluster, provisioned by https://github.com/akyriako/typesense-operator, and exposes both a JSON-based API 
and a built-in web UI for visualizing node status in real time.

## Features

### Cluster readiness & liveness:

* `/readyz` returns a detailed JSON report of cluster's **and** each node’s health.
* `/livez` returns a simple `200 OK` for basic liveness probes.

### Web UI

Interactive single-page Vue.js + Vuetify application inspired (sic!) by [podinfo](https://github.com/stefanprodan/podinfo) landing page.

* Auto-refresh every 3 seconds.
* Color-coded status indicators per node.


> [!IMPORTANT]
> This lightweight server is designed to run as a sidecar in the same Pod as your Typesense nodes deployed by [TyKO](https://github.com/akyriako/typesense-operator).



