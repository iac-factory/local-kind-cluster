resource "kubernetes_manifest" "serviceaccount_istio_system_loki" {
  manifest = {
    "apiVersion" = "v1"
    "automountServiceAccountToken" = true
    "kind" = "ServiceAccount"
    "metadata" = {
      "labels" = {
        "app.kubernetes.io/instance" = "loki"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "loki"
        "app.kubernetes.io/version" = "3.0.0"
        "helm.sh/chart" = "loki-6.0.0"
      }
      "name" = "loki"
      "namespace" = "istio-system"
    }
  }
}

resource "kubernetes_manifest" "configmap_istio_system_loki" {
  manifest = {
    "apiVersion" = "v1"
    "data" = {
      "config.yaml" = <<-EOT
      
      auth_enabled: false
      common:
        compactor_address: 'http://loki:3100'
        path_prefix: /var/loki
        replication_factor: 1
        storage:
          filesystem:
            chunks_directory: /var/loki/chunks
            rules_directory: /var/loki/rules
      frontend:
        scheduler_address: ""
        tail_proxy_url: http://loki-querier.istio-system.svc.cluster.local:3100
      frontend_worker:
        scheduler_address: ""
      index_gateway:
        mode: simple
      limits_config:
        max_cache_freshness_per_query: 10m
        query_timeout: 300s
        reject_old_samples: true
        reject_old_samples_max_age: 168h
        split_queries_by_interval: 15m
      memberlist:
        join_members:
        - loki-memberlist
      query_range:
        align_queries_with_step: true
      ruler:
        storage:
          type: local
      runtime_config:
        file: /etc/loki/runtime-config/runtime-config.yaml
      schema_config:
        configs:
        - from: "2024-04-01"
          index:
            period: 24h
            prefix: index_
          object_store: filesystem
          schema: v13
          store: tsdb
      server:
        grpc_listen_port: 9095
        http_listen_port: 3100
        http_server_read_timeout: 600s
        http_server_write_timeout: 600s
      storage_config:
        boltdb_shipper:
          index_gateway_client:
            server_address: ""
        hedging:
          at: 250ms
          max_per_second: 20
          up_to: 3
        tsdb_shipper:
          index_gateway_client:
            server_address: ""
      tracing:
        enabled: false
      EOT
    }
    "kind" = "ConfigMap"
    "metadata" = {
      "labels" = {
        "app.kubernetes.io/instance" = "loki"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "loki"
        "app.kubernetes.io/version" = "3.0.0"
        "helm.sh/chart" = "loki-6.0.0"
      }
      "name" = "loki"
      "namespace" = "istio-system"
    }
  }
}

resource "kubernetes_manifest" "configmap_istio_system_loki_runtime" {
  manifest = {
    "apiVersion" = "v1"
    "data" = {
      "runtime-config.yaml" = "{}"
    }
    "kind" = "ConfigMap"
    "metadata" = {
      "labels" = {
        "app.kubernetes.io/instance" = "loki"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "loki"
        "app.kubernetes.io/version" = "3.0.0"
        "helm.sh/chart" = "loki-6.0.0"
      }
      "name" = "loki-runtime"
      "namespace" = "istio-system"
    }
  }
}

resource "kubernetes_manifest" "service_istio_system_loki_memberlist" {
  manifest = {
    "apiVersion" = "v1"
    "kind" = "Service"
    "metadata" = {
      "labels" = {
        "app.kubernetes.io/instance" = "loki"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "loki"
        "app.kubernetes.io/version" = "3.0.0"
        "helm.sh/chart" = "loki-6.0.0"
      }
      "name" = "loki-memberlist"
      "namespace" = "istio-system"
    }
    "spec" = {
      "clusterIP" = "None"
      "ports" = [
        {
          "name" = "tcp"
          "port" = 7946
          "protocol" = "TCP"
          "targetPort" = "http-memberlist"
        },
      ]
      "selector" = {
        "app.kubernetes.io/instance" = "loki"
        "app.kubernetes.io/name" = "loki"
        "app.kubernetes.io/part-of" = "memberlist"
      }
      "type" = "ClusterIP"
    }
  }
}

resource "kubernetes_manifest" "service_istio_system_loki_headless" {
  manifest = {
    "apiVersion" = "v1"
    "kind" = "Service"
    "metadata" = {
      "labels" = {
        "app.kubernetes.io/instance" = "loki"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "loki"
        "app.kubernetes.io/version" = "3.0.0"
        "helm.sh/chart" = "loki-6.0.0"
        "prometheus.io/service-monitor" = "false"
        "variant" = "headless"
      }
      "name" = "loki-headless"
      "namespace" = "istio-system"
    }
    "spec" = {
      "clusterIP" = "None"
      "ports" = [
        {
          "name" = "http-metrics"
          "port" = 3100
          "protocol" = "TCP"
          "targetPort" = "http-metrics"
        },
      ]
      "selector" = {
        "app.kubernetes.io/instance" = "loki"
        "app.kubernetes.io/name" = "loki"
      }
    }
  }
}

resource "kubernetes_manifest" "service_istio_system_loki" {
  manifest = {
    "apiVersion" = "v1"
    "kind" = "Service"
    "metadata" = {
      "labels" = {
        "app.kubernetes.io/instance" = "loki"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "loki"
        "app.kubernetes.io/version" = "3.0.0"
        "helm.sh/chart" = "loki-6.0.0"
      }
      "name" = "loki"
      "namespace" = "istio-system"
    }
    "spec" = {
      "ports" = [
        {
          "name" = "http-metrics"
          "port" = 3100
          "protocol" = "TCP"
          "targetPort" = "http-metrics"
        },
        {
          "name" = "grpc"
          "port" = 9095
          "protocol" = "TCP"
          "targetPort" = "grpc"
        },
      ]
      "selector" = {
        "app.kubernetes.io/component" = "single-binary"
        "app.kubernetes.io/instance" = "loki"
        "app.kubernetes.io/name" = "loki"
      }
      "type" = "ClusterIP"
    }
  }
}

resource "kubernetes_manifest" "statefulset_istio_system_loki" {
  manifest = {
    "apiVersion" = "apps/v1"
    "kind" = "StatefulSet"
    "metadata" = {
      "labels" = {
        "app.kubernetes.io/component" = "single-binary"
        "app.kubernetes.io/instance" = "loki"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "loki"
        "app.kubernetes.io/part-of" = "memberlist"
        "app.kubernetes.io/version" = "3.0.0"
        "helm.sh/chart" = "loki-6.0.0"
      }
      "name" = "loki"
      "namespace" = "istio-system"
    }
    "spec" = {
      "persistentVolumeClaimRetentionPolicy" = {
        "whenDeleted" = "Delete"
        "whenScaled" = "Delete"
      }
      "podManagementPolicy" = "Parallel"
      "replicas" = 1
      "revisionHistoryLimit" = 10
      "selector" = {
        "matchLabels" = {
          "app.kubernetes.io/component" = "single-binary"
          "app.kubernetes.io/instance" = "loki"
          "app.kubernetes.io/name" = "loki"
        }
      }
      "serviceName" = "loki-headless"
      "template" = {
        "metadata" = {
          "annotations" = {
            "checksum/config" = "d378803f8e80ff751fb2c3d0262d651fc521f91169c30b550b3003701ee357cb"
          }
          "labels" = {
            "app.kubernetes.io/component" = "single-binary"
            "app.kubernetes.io/instance" = "loki"
            "app.kubernetes.io/name" = "loki"
            "app.kubernetes.io/part-of" = "memberlist"
          }
        }
        "spec" = {
          "affinity" = {
            "podAntiAffinity" = {
              "requiredDuringSchedulingIgnoredDuringExecution" = [
                {
                  "labelSelector" = {
                    "matchLabels" = {
                      "app.kubernetes.io/component" = "single-binary"
                    }
                  }
                  "topologyKey" = "kubernetes.io/hostname"
                },
              ]
            }
          }
          "automountServiceAccountToken" = true
          "containers" = [
            {
              "args" = [
                "-config.file=/etc/loki/config/config.yaml",
                "-target=all",
              ]
              "image" = "docker.io/grafana/loki:3.0.0"
              "imagePullPolicy" = "IfNotPresent"
              "name" = "loki"
              "ports" = [
                {
                  "containerPort" = 3100
                  "name" = "http-metrics"
                  "protocol" = "TCP"
                },
                {
                  "containerPort" = 9095
                  "name" = "grpc"
                  "protocol" = "TCP"
                },
                {
                  "containerPort" = 7946
                  "name" = "http-memberlist"
                  "protocol" = "TCP"
                },
              ]
              "readinessProbe" = {
                "httpGet" = {
                  "path" = "/ready"
                  "port" = "http-metrics"
                }
                "initialDelaySeconds" = 30
                "timeoutSeconds" = 1
              }
              "resources" = {}
              "securityContext" = {
                "allowPrivilegeEscalation" = false
                "capabilities" = {
                  "drop" = [
                    "ALL",
                  ]
                }
                "readOnlyRootFilesystem" = true
              }
              "volumeMounts" = [
                {
                  "mountPath" = "/tmp"
                  "name" = "tmp"
                },
                {
                  "mountPath" = "/etc/loki/config"
                  "name" = "config"
                },
                {
                  "mountPath" = "/etc/loki/runtime-config"
                  "name" = "runtime-config"
                },
                {
                  "mountPath" = "/var/loki"
                  "name" = "storage"
                },
              ]
            },
          ]
          "enableServiceLinks" = true
          "securityContext" = {
            "fsGroup" = 10001
            "runAsGroup" = 10001
            "runAsNonRoot" = true
            "runAsUser" = 10001
          }
          "serviceAccountName" = "loki"
          "terminationGracePeriodSeconds" = 30
          "volumes" = [
            {
              "emptyDir" = {}
              "name" = "tmp"
            },
            {
              "configMap" = {
                "items" = [
                  {
                    "key" = "config.yaml"
                    "path" = "config.yaml"
                  },
                ]
                "name" = "loki"
              }
              "name" = "config"
            },
            {
              "configMap" = {
                "name" = "loki-runtime"
              }
              "name" = "runtime-config"
            },
          ]
        }
      }
      "updateStrategy" = {
        "rollingUpdate" = {
          "partition" = 0
        }
      }
      "volumeClaimTemplates" = [
        {
          "apiVersion" = "v1"
          "kind" = "PersistentVolumeClaim"
          "metadata" = {
            "name" = "storage"
          }
          "spec" = {
            "accessModes" = [
              "ReadWriteOnce",
            ]
            "resources" = {
              "requests" = {
                "storage" = "10Gi"
              }
            }
          }
        },
      ]
    }
  }
}
