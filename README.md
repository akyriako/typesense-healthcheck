# typesense-healthcheck

A lightweight Go service that reports the aggregated health status of a Typesense cluster and exposes both a JSON-based API 
and a built-in web UI for visualizing node status in real time.

![image](https://github.com/user-attachments/assets/42db73c4-f175-4fb5-a862-ec83ace6ada7)

## Features

### Cluster readiness & liveness:

* `/livez` returns a simple `200 OK` for basic liveness probes.
* `/readyz` returns a detailed JSON report of cluster's **and** each node’s health, e.g. :

```json
{
    "cluster_status": "OK",
    "cluster_health": true,
    "nodes_health_check": {
        "c-kind-2-sts-0.c-kind-2-sts-svc": {
            "node_status": {
                "committed_index": 6813,
                "queued_writes": 0,
                "state": "FOLLOWER"
            },
            "node_health": {
                "ok": true
            }
        },
        "c-kind-2-sts-1.c-kind-2-sts-svc": {
            "node_status": {
                "committed_index": 6813,
                "queued_writes": 0,
                "state": "FOLLOWER"
            },
            "node_health": {
                "ok": true
            }
        },
        "c-kind-2-sts-2.c-kind-2-sts-svc": {
            "node_status": {
                "committed_index": 6813,
                "queued_writes": 0,
                "state": "FOLLOWER"
            },
            "node_health": {
                "ok": true
            }
        },
        "c-kind-2-sts-3.c-kind-2-sts-svc": {
            "node_status": {
                "committed_index": 6813,
                "queued_writes": 0,
                "state": "LEADER"
            },
            "node_health": {
                "ok": true
            }
        },
        "c-kind-2-sts-4.c-kind-2-sts-svc": {
            "node_status": {
                "committed_index": 6813,
                "queued_writes": 0,
                "state": "FOLLOWER"
            },
            "node_health": {
                "ok": true
            }
        },
        "c-kind-2-sts-5.c-kind-2-sts-svc": {
            "node_status": {
                "committed_index": 6813,
                "queued_writes": 0,
                "state": "FOLLOWER"
            },
            "node_health": {
                "ok": true
            }
        },
        "c-kind-2-sts-6.c-kind-2-sts-svc": {
            "node_status": {
                "committed_index": 6813,
                "queued_writes": 0,
                "state": "FOLLOWER"
            },
            "node_health": {
                "ok": true
            }
        }
    }
}
```

### Web UI

Interactive single-page Vue.js + Vuetify application inspired (sic!) by [podinfo](https://github.com/stefanprodan/podinfo) landing page.

* Auto-refresh every 3 seconds.
* Color-coded status indicators per node.

## Configuration 

| Env Var                  | Type   | Default                                                                                      | Required |
|--------------------------|--------|----------------------------------------------------------------------------------------------|----------|
| `LOG_LEVEL`              | `int`  | `0`                                                                                          | No       |
| `CLUSTER_NAMESPACE`      | `string` | `default`                                                                                   | No       |
| `TYPESENSE_API_KEY`      | `string` | —                                                                                            | Yes      |
| `TYPESENSE_PROTOCOL`     | `string` | `http`                                                                                       | No       |
| `TYPESENSE_API_PORT`     | `uint`   | `8108`                                                                                       | No       |
| `TYPESENSE_PEERING_PORT` | `uint`   | `8107`                                                                                       | No       |
| `HEALTHCHECK_PORT`       | `uint`   | `8808`                                                                                       | No       |
| `TYPESENSE_NODES`        | `string` | `/usr/share/typesense/nodes`                                                                 | No       |

> [!IMPORTANT]
> Although this lightweight server is designed to run as a sidecar in the same Pod as your Typesense nodes deployed by [TyKO](https://github.com/akyriako/typesense-operator),
> you can monitor any Typesense Cluster as long as you provide a reachable path using `TYPESENSE_NODES` variable (e.g. by running this service as a docker container and
> mounting your nodes list file in it) 



