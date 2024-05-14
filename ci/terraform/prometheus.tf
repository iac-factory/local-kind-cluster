resource "kubernetes_manifest" "serviceaccount_istio_system_prometheus" {
  depends_on = [helm_release.istio-base, helm_release.istiod, helm_release.istio-gateway]

  manifest = {
    "apiVersion" = "v1"
    "kind" = "ServiceAccount"
    "metadata" = {
      "labels" = {
        "app.kubernetes.io/component" = "server"
        "app.kubernetes.io/instance" = "prometheus"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "prometheus"
        "app.kubernetes.io/part-of" = "prometheus"
        "app.kubernetes.io/version" = "v2.51.1"
        "helm.sh/chart" = "prometheus-25.19.1"
      }
      "name" = "prometheus"
      "namespace" = "istio-system"
    }
  }
}

resource "kubernetes_manifest" "configmap_istio_system_prometheus" {
  depends_on = [helm_release.istio-base, helm_release.istiod, helm_release.istio-gateway]

  manifest = {
    "apiVersion" = "v1"
    "data" = {
      "alerting_rules.yml" = <<-EOT
      {}

      EOT
      "alerts" = <<-EOT
      {}

      EOT
      "allow-snippet-annotations" = "false"
      "prometheus.yml" = <<-EOT
      global:
        evaluation_interval: 1m
        scrape_interval: 15s
        scrape_timeout: 10s
      rule_files:
      - /etc/config/recording_rules.yml
      - /etc/config/alerting_rules.yml
      - /etc/config/rules
      - /etc/config/alerts
      scrape_configs:
      - job_name: prometheus
        static_configs:
        - targets:
          - localhost:9090
      - bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
        job_name: kubernetes-apiservers
        kubernetes_sd_configs:
        - role: endpoints
        relabel_configs:
        - action: keep
          regex: default;kubernetes;https
          source_labels:
          - __meta_kubernetes_namespace
          - __meta_kubernetes_service_name
          - __meta_kubernetes_endpoint_port_name
        scheme: https
        tls_config:
          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
          insecure_skip_verify: true
      - bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
        job_name: kubernetes-nodes
        kubernetes_sd_configs:
        - role: node
        relabel_configs:
        - action: labelmap
          regex: __meta_kubernetes_node_label_(.+)
        - replacement: kubernetes.default.svc:443
          target_label: __address__
        - regex: (.+)
          replacement: /api/v1/nodes/$1/proxy/metrics
          source_labels:
          - __meta_kubernetes_node_name
          target_label: __metrics_path__
        scheme: https
        tls_config:
          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
          insecure_skip_verify: true
      - bearer_token_file: /var/run/secrets/kubernetes.io/serviceaccount/token
        job_name: kubernetes-nodes-cadvisor
        kubernetes_sd_configs:
        - role: node
        relabel_configs:
        - action: labelmap
          regex: __meta_kubernetes_node_label_(.+)
        - replacement: kubernetes.default.svc:443
          target_label: __address__
        - regex: (.+)
          replacement: /api/v1/nodes/$1/proxy/metrics/cadvisor
          source_labels:
          - __meta_kubernetes_node_name
          target_label: __metrics_path__
        scheme: https
        tls_config:
          ca_file: /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
          insecure_skip_verify: true
      - honor_labels: true
        job_name: kubernetes-service-endpoints
        kubernetes_sd_configs:
        - role: endpoints
        relabel_configs:
        - action: keep
          regex: true
          source_labels:
          - __meta_kubernetes_service_annotation_prometheus_io_scrape
        - action: drop
          regex: true
          source_labels:
          - __meta_kubernetes_service_annotation_prometheus_io_scrape_slow
        - action: replace
          regex: (https?)
          source_labels:
          - __meta_kubernetes_service_annotation_prometheus_io_scheme
          target_label: __scheme__
        - action: replace
          regex: (.+)
          source_labels:
          - __meta_kubernetes_service_annotation_prometheus_io_path
          target_label: __metrics_path__
        - action: replace
          regex: (.+?)(?::\d+)?;(\d+)
          replacement: $1:$2
          source_labels:
          - __address__
          - __meta_kubernetes_service_annotation_prometheus_io_port
          target_label: __address__
        - action: labelmap
          regex: __meta_kubernetes_service_annotation_prometheus_io_param_(.+)
          replacement: __param_$1
        - action: labelmap
          regex: __meta_kubernetes_service_label_(.+)
        - action: replace
          source_labels:
          - __meta_kubernetes_namespace
          target_label: namespace
        - action: replace
          source_labels:
          - __meta_kubernetes_service_name
          target_label: service
        - action: replace
          source_labels:
          - __meta_kubernetes_pod_node_name
          target_label: node
      - honor_labels: true
        job_name: kubernetes-service-endpoints-slow
        kubernetes_sd_configs:
        - role: endpoints
        relabel_configs:
        - action: keep
          regex: true
          source_labels:
          - __meta_kubernetes_service_annotation_prometheus_io_scrape_slow
        - action: replace
          regex: (https?)
          source_labels:
          - __meta_kubernetes_service_annotation_prometheus_io_scheme
          target_label: __scheme__
        - action: replace
          regex: (.+)
          source_labels:
          - __meta_kubernetes_service_annotation_prometheus_io_path
          target_label: __metrics_path__
        - action: replace
          regex: (.+?)(?::\d+)?;(\d+)
          replacement: $1:$2
          source_labels:
          - __address__
          - __meta_kubernetes_service_annotation_prometheus_io_port
          target_label: __address__
        - action: labelmap
          regex: __meta_kubernetes_service_annotation_prometheus_io_param_(.+)
          replacement: __param_$1
        - action: labelmap
          regex: __meta_kubernetes_service_label_(.+)
        - action: replace
          source_labels:
          - __meta_kubernetes_namespace
          target_label: namespace
        - action: replace
          source_labels:
          - __meta_kubernetes_service_name
          target_label: service
        - action: replace
          source_labels:
          - __meta_kubernetes_pod_node_name
          target_label: node
        scrape_interval: 5m
        scrape_timeout: 30s
      - honor_labels: true
        job_name: prometheus-pushgateway
        kubernetes_sd_configs:
        - role: service
        relabel_configs:
        - action: keep
          regex: pushgateway
          source_labels:
          - __meta_kubernetes_service_annotation_prometheus_io_probe
      - honor_labels: true
        job_name: kubernetes-services
        kubernetes_sd_configs:
        - role: service
        metrics_path: /probe
        params:
          module:
          - http_2xx
        relabel_configs:
        - action: keep
          regex: true
          source_labels:
          - __meta_kubernetes_service_annotation_prometheus_io_probe
        - source_labels:
          - __address__
          target_label: __param_target
        - replacement: blackbox
          target_label: __address__
        - source_labels:
          - __param_target
          target_label: instance
        - action: labelmap
          regex: __meta_kubernetes_service_label_(.+)
        - source_labels:
          - __meta_kubernetes_namespace
          target_label: namespace
        - source_labels:
          - __meta_kubernetes_service_name
          target_label: service
      - honor_labels: true
        job_name: kubernetes-pods
        kubernetes_sd_configs:
        - role: pod
        relabel_configs:
        - action: keep
          regex: true
          source_labels:
          - __meta_kubernetes_pod_annotation_prometheus_io_scrape
        - action: drop
          regex: true
          source_labels:
          - __meta_kubernetes_pod_annotation_prometheus_io_scrape_slow
        - action: replace
          regex: (https?)
          source_labels:
          - __meta_kubernetes_pod_annotation_prometheus_io_scheme
          target_label: __scheme__
        - action: replace
          regex: (.+)
          source_labels:
          - __meta_kubernetes_pod_annotation_prometheus_io_path
          target_label: __metrics_path__
        - action: replace
          regex: (\d+);(([A-Fa-f0-9]{1,4}::?){1,7}[A-Fa-f0-9]{1,4})
          replacement: '[$2]:$1'
          source_labels:
          - __meta_kubernetes_pod_annotation_prometheus_io_port
          - __meta_kubernetes_pod_ip
          target_label: __address__
        - action: replace
          regex: (\d+);((([0-9]+?)(\.|$)){4})
          replacement: $2:$1
          source_labels:
          - __meta_kubernetes_pod_annotation_prometheus_io_port
          - __meta_kubernetes_pod_ip
          target_label: __address__
        - action: labelmap
          regex: __meta_kubernetes_pod_annotation_prometheus_io_param_(.+)
          replacement: __param_$1
        - action: labelmap
          regex: __meta_kubernetes_pod_label_(.+)
        - action: replace
          source_labels:
          - __meta_kubernetes_namespace
          target_label: namespace
        - action: replace
          source_labels:
          - __meta_kubernetes_pod_name
          target_label: pod
        - action: drop
          regex: Pending|Succeeded|Failed|Completed
          source_labels:
          - __meta_kubernetes_pod_phase
        - action: replace
          source_labels:
          - __meta_kubernetes_pod_node_name
          target_label: node
      - honor_labels: true
        job_name: kubernetes-pods-slow
        kubernetes_sd_configs:
        - role: pod
        relabel_configs:
        - action: keep
          regex: true
          source_labels:
          - __meta_kubernetes_pod_annotation_prometheus_io_scrape_slow
        - action: replace
          regex: (https?)
          source_labels:
          - __meta_kubernetes_pod_annotation_prometheus_io_scheme
          target_label: __scheme__
        - action: replace
          regex: (.+)
          source_labels:
          - __meta_kubernetes_pod_annotation_prometheus_io_path
          target_label: __metrics_path__
        - action: replace
          regex: (\d+);(([A-Fa-f0-9]{1,4}::?){1,7}[A-Fa-f0-9]{1,4})
          replacement: '[$2]:$1'
          source_labels:
          - __meta_kubernetes_pod_annotation_prometheus_io_port
          - __meta_kubernetes_pod_ip
          target_label: __address__
        - action: replace
          regex: (\d+);((([0-9]+?)(\.|$)){4})
          replacement: $2:$1
          source_labels:
          - __meta_kubernetes_pod_annotation_prometheus_io_port
          - __meta_kubernetes_pod_ip
          target_label: __address__
        - action: labelmap
          regex: __meta_kubernetes_pod_annotation_prometheus_io_param_(.+)
          replacement: __param_$1
        - action: labelmap
          regex: __meta_kubernetes_pod_label_(.+)
        - action: replace
          source_labels:
          - __meta_kubernetes_namespace
          target_label: namespace
        - action: replace
          source_labels:
          - __meta_kubernetes_pod_name
          target_label: pod
        - action: drop
          regex: Pending|Succeeded|Failed|Completed
          source_labels:
          - __meta_kubernetes_pod_phase
        - action: replace
          source_labels:
          - __meta_kubernetes_pod_node_name
          target_label: node
        scrape_interval: 5m
        scrape_timeout: 30s

      EOT
      "recording_rules.yml" = <<-EOT
      {}

      EOT
      "rules" = "{}"
    }
    "kind" = "ConfigMap"
    "metadata" = {
      "labels" = {
        "app.kubernetes.io/component" = "server"
        "app.kubernetes.io/instance" = "prometheus"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "prometheus"
        "app.kubernetes.io/part-of" = "prometheus"
        "app.kubernetes.io/version" = "v2.51.1"
        "helm.sh/chart" = "prometheus-25.19.1"
      }
      "name" = "prometheus"
      "namespace" = "istio-system"
    }
  }
}

resource "kubernetes_manifest" "clusterrole_prometheus" {
  depends_on = [helm_release.istio-base, helm_release.istiod, helm_release.istio-gateway]

  manifest = {
    "apiVersion" = "rbac.authorization.k8s.io/v1"
    "kind" = "ClusterRole"
    "metadata" = {
      "labels" = {
        "app.kubernetes.io/component" = "server"
        "app.kubernetes.io/instance" = "prometheus"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "prometheus"
        "app.kubernetes.io/part-of" = "prometheus"
        "app.kubernetes.io/version" = "v2.51.1"
        "helm.sh/chart" = "prometheus-25.19.1"
      }
      "name" = "prometheus"
    }
    "rules" = [
      {
        "apiGroups" = [
          "",
        ]
        "resources" = [
          "nodes",
          "nodes/proxy",
          "nodes/metrics",
          "services",
          "endpoints",
          "pods",
          "ingresses",
          "configmaps",
        ]
        "verbs" = [
          "get",
          "list",
          "watch",
        ]
      },
      {
        "apiGroups" = [
          "extensions",
          "networking.k8s.io",
        ]
        "resources" = [
          "ingresses/status",
          "ingresses",
        ]
        "verbs" = [
          "get",
          "list",
          "watch",
        ]
      },
      {
        "apiGroups" = [
          "discovery.k8s.io",
        ]
        "resources" = [
          "endpointslices",
        ]
        "verbs" = [
          "get",
          "list",
          "watch",
        ]
      },
      {
        "nonResourceURLs" = [
          "/metrics",
        ]
        "verbs" = [
          "get",
        ]
      },
    ]
  }
}

resource "kubernetes_manifest" "clusterrolebinding_prometheus" {
  depends_on = [helm_release.istio-base, helm_release.istiod, helm_release.istio-gateway]

  manifest = {
    "apiVersion" = "rbac.authorization.k8s.io/v1"
    "kind" = "ClusterRoleBinding"
    "metadata" = {
      "labels" = {
        "app.kubernetes.io/component" = "server"
        "app.kubernetes.io/instance" = "prometheus"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "prometheus"
        "app.kubernetes.io/part-of" = "prometheus"
        "app.kubernetes.io/version" = "v2.51.1"
        "helm.sh/chart" = "prometheus-25.19.1"
      }
      "name" = "prometheus"
    }
    "roleRef" = {
      "apiGroup" = "rbac.authorization.k8s.io"
      "kind" = "ClusterRole"
      "name" = "prometheus"
    }
    "subjects" = [
      {
        "kind" = "ServiceAccount"
        "name" = "prometheus"
        "namespace" = "istio-system"
      },
    ]
  }
}

resource "kubernetes_manifest" "service_istio_system_prometheus" {
  depends_on = [helm_release.istio-base, helm_release.istiod, helm_release.istio-gateway]

  manifest = {
    "apiVersion" = "v1"
    "kind" = "Service"
    "metadata" = {
      "labels" = {
        "app.kubernetes.io/component" = "server"
        "app.kubernetes.io/instance" = "prometheus"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "prometheus"
        "app.kubernetes.io/part-of" = "prometheus"
        "app.kubernetes.io/version" = "v2.51.1"
        "helm.sh/chart" = "prometheus-25.19.1"
      }
      "name" = "prometheus"
      "namespace" = "istio-system"
    }
    "spec" = {
      "ports" = [
        {
          "name" = "http"
          "port" = 9090
          "protocol" = "TCP"
          "targetPort" = 9090
        },
      ]
      "selector" = {
        "app.kubernetes.io/component" = "server"
        "app.kubernetes.io/instance" = "prometheus"
        "app.kubernetes.io/name" = "prometheus"
      }
      "sessionAffinity" = "None"
      "type" = "ClusterIP"
    }
  }
}

resource "kubernetes_manifest" "deployment_istio_system_prometheus" {
  depends_on = [helm_release.istio-base, helm_release.istiod, helm_release.istio-gateway]

  manifest = {
    "apiVersion" = "apps/v1"
    "kind" = "Deployment"
    "metadata" = {
      "labels" = {
        "app.kubernetes.io/component" = "server"
        "app.kubernetes.io/instance" = "prometheus"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "prometheus"
        "app.kubernetes.io/part-of" = "prometheus"
        "app.kubernetes.io/version" = "v2.51.1"
        "helm.sh/chart" = "prometheus-25.19.1"
      }
      "name" = "prometheus"
      "namespace" = "istio-system"
    }
    "spec" = {
      "replicas" = 1
      "revisionHistoryLimit" = 10
      "selector" = {
        "matchLabels" = {
          "app.kubernetes.io/component" = "server"
          "app.kubernetes.io/instance" = "prometheus"
          "app.kubernetes.io/name" = "prometheus"
        }
      }
      "strategy" = {
        "type" = "Recreate"
      }
      "template" = {
        "metadata" = {
          "labels" = {
            "app.kubernetes.io/component" = "server"
            "app.kubernetes.io/instance" = "prometheus"
            "app.kubernetes.io/managed-by" = "Helm"
            "app.kubernetes.io/name" = "prometheus"
            "app.kubernetes.io/part-of" = "prometheus"
            "app.kubernetes.io/version" = "v2.51.1"
            "helm.sh/chart" = "prometheus-25.19.1"
            "sidecar.istio.io/inject" = "false"
          }
        }
        "spec" = {
          "containers" = [
            {
              "args" = [
                "--watched-dir=/etc/config",
                "--reload-url=http://127.0.0.1:9090/-/reload",
              ]
              "image" = "ghcr.io/prometheus-operator/prometheus-config-reloader:v0.72.0"
              "imagePullPolicy" = "IfNotPresent"
              "name" = "prometheus-server-configmap-reload"
              "volumeMounts" = [
                {
                  "mountPath" = "/etc/config"
                  "name" = "config-volume"
                  "readOnly" = true
                },
              ]
            },
            {
              "args" = [
                "--storage.tsdb.retention.time=15d",
                "--config.file=/etc/config/prometheus.yml",
                "--storage.tsdb.path=/data",
                "--web.console.libraries=/etc/prometheus/console_libraries",
                "--web.console.templates=/etc/prometheus/consoles",
                "--web.enable-lifecycle",
              ]
              "image" = "prom/prometheus:v2.51.1"
              "imagePullPolicy" = "IfNotPresent"
              "livenessProbe" = {
                "failureThreshold" = 3
                "httpGet" = {
                  "path" = "/-/healthy"
                  "port" = 9090
                  "scheme" = "HTTP"
                }
                "initialDelaySeconds" = 30
                "periodSeconds" = 15
                "successThreshold" = 1
                "timeoutSeconds" = 10
              }
              "name" = "prometheus-server"
              "ports" = [
                {
                  "containerPort" = 9090
                },
              ]
              "readinessProbe" = {
                "failureThreshold" = 3
                "httpGet" = {
                  "path" = "/-/ready"
                  "port" = 9090
                  "scheme" = "HTTP"
                }
                "periodSeconds" = 5
                "successThreshold" = 1
                "timeoutSeconds" = 4
              }
              "volumeMounts" = [
                {
                  "mountPath" = "/etc/config"
                  "name" = "config-volume"
                },
                {
                  "mountPath" = "/data"
                  "name" = "storage-volume"
                },
              ]
            },
          ]
          "dnsPolicy" = "ClusterFirst"
          "enableServiceLinks" = true
          "serviceAccountName" = "prometheus"
          "terminationGracePeriodSeconds" = 300
          "volumes" = [
            {
              "configMap" = {
                "name" = "prometheus"
              }
              "name" = "config-volume"
            },
            {
              "emptyDir" = {}
              "name" = "storage-volume"
            },
          ]
        }
      }
    }
  }
}
