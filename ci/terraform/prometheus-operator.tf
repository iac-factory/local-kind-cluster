resource "kubernetes_manifest" "podmonitor_istio_system_envoy_stats_monitor" {
  depends_on = [
    helm_release.istiod, helm_release.istio-base, helm_release.istio-gateway
  ]

  manifest = {
    "apiVersion" = "monitoring.coreos.com/v1"
    "kind" = "PodMonitor"
    "metadata" = {
      "labels" = {
        "monitoring" = "istio-proxies"
        "release" = "istio"
      }
      "name" = "envoy-stats-monitor"
      "namespace" = "istio-system"
    }
    "spec" = {
      "jobLabel" = "envoy-stats"
      "namespaceSelector" = {
        "any" = true
      }
      "podMetricsEndpoints" = [
        {
          "interval" = "15s"
          "path" = "/stats/prometheus"
          "relabelings" = [
            {
              "action" = "keep"
              "regex" = "istio-proxy"
              "sourceLabels" = [
                "__meta_kubernetes_pod_container_name",
              ]
            },
            {
              "action" = "keep"
              "sourceLabels" = [
                "__meta_kubernetes_pod_annotationpresent_prometheus_io_scrape",
              ]
            },
            {
              "action" = "replace"
              "regex" = "(\\d+);(([A-Fa-f0-9]{1,4}::?){1,7}[A-Fa-f0-9]{1,4})"
              "replacement" = "[$2]:$1"
              "sourceLabels" = [
                "__meta_kubernetes_pod_annotation_prometheus_io_port",
                "__meta_kubernetes_pod_ip",
              ]
              "targetLabel" = "__address__"
            },
            {
              "action" = "replace"
              "regex" = "(\\d+);((([0-9]+?)(\\.|$)){4})"
              "replacement" = "$2:$1"
              "sourceLabels" = [
                "__meta_kubernetes_pod_annotation_prometheus_io_port",
                "__meta_kubernetes_pod_ip",
              ]
              "targetLabel" = "__address__"
            },
            {
              "action" = "labeldrop"
              "regex" = "__meta_kubernetes_pod_label_(.+)"
            },
            {
              "action" = "replace"
              "sourceLabels" = [
                "__meta_kubernetes_namespace",
              ]
              "targetLabel" = "namespace"
            },
            {
              "action" = "replace"
              "sourceLabels" = [
                "__meta_kubernetes_pod_name",
              ]
              "targetLabel" = "pod_name"
            },
          ]
        },
      ]
      "selector" = {
        "matchExpressions" = [
          {
            "key" = "istio-prometheus-ignore"
            "operator" = "DoesNotExist"
          },
        ]
      }
    }
  }

}

resource "kubernetes_manifest" "servicemonitor_istio_system_istio_component_monitor" {
  depends_on = [
    helm_release.istiod, helm_release.istio-base, helm_release.istio-gateway
  ]

  manifest = {
    "apiVersion" = "monitoring.coreos.com/v1"
    "kind" = "ServiceMonitor"
    "metadata" = {
      "labels" = {
        "monitoring" = "istio-components"
        "release" = "istio"
      }
      "name" = "istio-component-monitor"
      "namespace" = "istio-system"
    }
    "spec" = {
      "endpoints" = [
        {
          "interval" = "15s"
          "port" = "http-monitoring"
        },
      ]
      "jobLabel" = "istio"
      "namespaceSelector" = {
        "any" = true
      }
      "selector" = {
        "matchExpressions" = [
          {
            "key" = "istio"
            "operator" = "In"
            "values" = [
              "pilot",
            ]
          },
        ]
      }
      "targetLabels" = [
        "app",
      ]
    }
  }

}
