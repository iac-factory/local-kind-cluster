resource "kubernetes_manifest" "deployment_istio_system_jaeger" {
  manifest = {
    "apiVersion" = "apps/v1"
    "kind" = "Deployment"
    "metadata" = {
      "labels" = {
        "app" = "jaeger"
      }
      "name" = "jaeger"
      "namespace" = "istio-system"
    }
    "spec" = {
      "selector" = {
        "matchLabels" = {
          "app" = "jaeger"
        }
      }
      "template" = {
        "metadata" = {
          "annotations" = {
            "prometheus.io/port" = "14269"
            "prometheus.io/scrape" = "true"
          }
          "labels" = {
            "app" = "jaeger"
            "sidecar.istio.io/inject" = "false"
          }
        }
        "spec" = {
          "containers" = [
            {
              "env" = [
                {
                  "name" = "BADGER_EPHEMERAL"
                  "value" = "false"
                },
                {
                  "name" = "SPAN_STORAGE_TYPE"
                  "value" = "badger"
                },
                {
                  "name" = "BADGER_DIRECTORY_VALUE"
                  "value" = "/badger/data"
                },
                {
                  "name" = "BADGER_DIRECTORY_KEY"
                  "value" = "/badger/key"
                },
                {
                  "name" = "COLLECTOR_ZIPKIN_HOST_PORT"
                  "value" = ":9411"
                },
                {
                  "name" = "MEMORY_MAX_TRACES"
                  "value" = "50000"
                },
                {
                  "name" = "QUERY_BASE_PATH"
                  "value" = "/jaeger"
                },
              ]
              "image" = "docker.io/jaegertracing/all-in-one:1.56"
              "livenessProbe" = {
                "httpGet" = {
                  "path" = "/"
                  "port" = 14269
                }
              }
              "name" = "jaeger"
              "readinessProbe" = {
                "httpGet" = {
                  "path" = "/"
                  "port" = 14269
                }
              }
              "resources" = {
                "requests" = {
                  "cpu" = "10m"
                }
              }
              "volumeMounts" = [
                {
                  "mountPath" = "/badger"
                  "name" = "data"
                },
              ]
            },
          ]
          "volumes" = [
            {
              "emptyDir" = {}
              "name" = "data"
            },
          ]
        }
      }
    }
  }
}

resource "kubernetes_manifest" "service_istio_system_tracing" {
  manifest = {
    "apiVersion" = "v1"
    "kind" = "Service"
    "metadata" = {
      "labels" = {
        "app" = "jaeger"
      }
      "name" = "tracing"
      "namespace" = "istio-system"
    }
    "spec" = {
      "ports" = [
        {
          "name" = "http-query"
          "port" = 80
          "protocol" = "TCP"
          "targetPort" = 16686
        },
        {
          "name" = "grpc-query"
          "port" = 16685
          "protocol" = "TCP"
          "targetPort" = 16685
        },
      ]
      "selector" = {
        "app" = "jaeger"
      }
      "type" = "ClusterIP"
    }
  }
}

resource "kubernetes_manifest" "service_istio_system_zipkin" {
  manifest = {
    "apiVersion" = "v1"
    "kind" = "Service"
    "metadata" = {
      "labels" = {
        "name" = "zipkin"
      }
      "name" = "zipkin"
      "namespace" = "istio-system"
    }
    "spec" = {
      "ports" = [
        {
          "name" = "http-query"
          "port" = 9411
          "targetPort" = 9411
        },
      ]
      "selector" = {
        "app" = "jaeger"
      }
    }
  }
}

resource "kubernetes_manifest" "service_istio_system_jaeger_collector" {
  manifest = {
    "apiVersion" = "v1"
    "kind" = "Service"
    "metadata" = {
      "labels" = {
        "app" = "jaeger"
      }
      "name" = "jaeger-collector"
      "namespace" = "istio-system"
    }
    "spec" = {
      "ports" = [
        {
          "name" = "jaeger-collector-http"
          "port" = 14268
          "protocol" = "TCP"
          "targetPort" = 14268
        },
        {
          "name" = "jaeger-collector-grpc"
          "port" = 14250
          "protocol" = "TCP"
          "targetPort" = 14250
        },
        {
          "name" = "http-zipkin"
          "port" = 9411
          "targetPort" = 9411
        },
        {
          "name" = "grpc-otel"
          "port" = 4317
        },
        {
          "name" = "http-otel"
          "port" = 4318
        },
      ]
      "selector" = {
        "app" = "jaeger"
      }
      "type" = "ClusterIP"
    }
  }
}
