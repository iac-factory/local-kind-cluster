resource "kubernetes_manifest" "serviceaccount_istio_system_kiali" {
  depends_on = [helm_release.istio-base, helm_release.istiod, helm_release.istio-gateway]

  manifest = {
    "apiVersion" = "v1"
    "kind" = "ServiceAccount"
    "metadata" = {
      "labels" = {
        "app" = "kiali"
        "app.kubernetes.io/instance" = "kiali"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "kiali"
        "app.kubernetes.io/part-of" = "kiali"
        "app.kubernetes.io/version" = "v1.82.0"
        "helm.sh/chart" = "kiali-server-1.82.0"
        "version" = "v1.82.0"
      }
      "name" = "kiali"
      "namespace" = "istio-system"
    }
  }
}

resource "kubernetes_manifest" "configmap_istio_system_kiali" {
  depends_on = [helm_release.istio-base, helm_release.istiod, helm_release.istio-gateway]

  manifest = {
    "apiVersion" = "v1"
    "data" = {
      "config.yaml" = <<-EOT
      additional_display_details:
      - annotation: kiali.io/api-spec
        icon_annotation: kiali.io/api-type
        title: API Documentation
      auth:
        openid: {}
        openshift:
          client_id_prefix: kiali
        strategy: anonymous
      clustering:
        autodetect_secrets:
          enabled: true
          label: kiali.io/multiCluster=true
        clusters: []
      deployment:
        accessible_namespaces:
        - '**'
        additional_service_yaml: {}
        affinity:
          node: {}
          pod: {}
          pod_anti: {}
        configmap_annotations: {}
        custom_secrets: []
        host_aliases: []
        hpa:
          api_version: autoscaling/v2
          spec: {}
        image_digest: ""
        image_name: quay.io/kiali/kiali
        image_pull_policy: Always
        image_pull_secrets: []
        image_version: v1.82
        ingress:
          additional_labels: {}
          class_name: nginx
          override_yaml:
            metadata: {}
        ingress_enabled: false
        instance_name: kiali
        logger:
          log_format: text
          log_level: info
          sampler_rate: "1"
          time_field_format: 2006-01-02T15:04:05Z07:00
        namespace: istio-system
        node_selector: {}
        pod_annotations: {}
        pod_labels:
          sidecar.istio.io/inject: "false"
        priority_class_name: ""
        replicas: 1
        resources:
          limits:
            memory: 1Gi
          requests:
            cpu: 10m
            memory: 64Mi
        secret_name: kiali
        security_context: {}
        service_annotations: {}
        service_type: ""
        tolerations: []
        version_label: v1.82.0
        view_only_mode: false
      external_services:
        custom_dashboards:
          enabled: true
        istio:
          root_namespace: istio-system
        tracing:
          enabled: false
      identity:
        cert_file: ""
        private_key_file: ""
      istio_namespace: istio-system
      kiali_feature_flags:
        certificates_information_indicators:
          enabled: true
          secrets:
          - cacerts
          - istio-ca-secret
        disabled_features: []
        validations:
          ignore:
          - KIA1301
      login_token:
        signing_key: CHANGEME00000000
      server:
        observability:
          metrics:
            enabled: true
            port: 9090
        port: 20001
        web_root: /kiali

      EOT
    }
    "kind" = "ConfigMap"
    "metadata" = {
      "labels" = {
        "app" = "kiali"
        "app.kubernetes.io/instance" = "kiali"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "kiali"
        "app.kubernetes.io/part-of" = "kiali"
        "app.kubernetes.io/version" = "v1.82.0"
        "helm.sh/chart" = "kiali-server-1.82.0"
        "version" = "v1.82.0"
      }
      "name" = "kiali"
      "namespace" = "istio-system"
    }
  }
}

resource "kubernetes_manifest" "clusterrole_kiali_viewer" {
  depends_on = [helm_release.istio-base, helm_release.istiod, helm_release.istio-gateway]

  manifest = {
    "apiVersion" = "rbac.authorization.k8s.io/v1"
    "kind" = "ClusterRole"
    "metadata" = {
      "labels" = {
        "app" = "kiali"
        "app.kubernetes.io/instance" = "kiali"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "kiali"
        "app.kubernetes.io/part-of" = "kiali"
        "app.kubernetes.io/version" = "v1.82.0"
        "helm.sh/chart" = "kiali-server-1.82.0"
        "version" = "v1.82.0"
      }
      "name" = "kiali-viewer"
    }
    "rules" = [
      {
        "apiGroups" = [
          "",
        ]
        "resources" = [
          "configmaps",
          "endpoints",
          "pods/log",
        ]
        "verbs" = [
          "get",
          "list",
          "watch",
        ]
      },
      {
        "apiGroups" = [
          "",
        ]
        "resources" = [
          "namespaces",
          "pods",
          "replicationcontrollers",
          "services",
        ]
        "verbs" = [
          "get",
          "list",
          "watch",
        ]
      },
      {
        "apiGroups" = [
          "",
        ]
        "resources" = [
          "pods/portforward",
        ]
        "verbs" = [
          "create",
          "post",
        ]
      },
      {
        "apiGroups" = [
          "extensions",
          "apps",
        ]
        "resources" = [
          "daemonsets",
          "deployments",
          "replicasets",
          "statefulsets",
        ]
        "verbs" = [
          "get",
          "list",
          "watch",
        ]
      },
      {
        "apiGroups" = [
          "batch",
        ]
        "resources" = [
          "cronjobs",
          "jobs",
        ]
        "verbs" = [
          "get",
          "list",
          "watch",
        ]
      },
      {
        "apiGroups" = [
          "networking.istio.io",
          "security.istio.io",
          "extensions.istio.io",
          "telemetry.istio.io",
          "gateway.networking.k8s.io",
        ]
        "resources" = [
          "*",
        ]
        "verbs" = [
          "get",
          "list",
          "watch",
        ]
      },
      {
        "apiGroups" = [
          "apps.openshift.io",
        ]
        "resources" = [
          "deploymentconfigs",
        ]
        "verbs" = [
          "get",
          "list",
          "watch",
        ]
      },
      {
        "apiGroups" = [
          "project.openshift.io",
        ]
        "resources" = [
          "projects",
        ]
        "verbs" = [
          "get",
        ]
      },
      {
        "apiGroups" = [
          "route.openshift.io",
        ]
        "resources" = [
          "routes",
        ]
        "verbs" = [
          "get",
        ]
      },
      {
        "apiGroups" = [
          "authentication.k8s.io",
        ]
        "resources" = [
          "tokenreviews",
        ]
        "verbs" = [
          "create",
        ]
      },
    ]
  }
}

resource "kubernetes_manifest" "clusterrole_kiali" {
  depends_on = [helm_release.istio-base, helm_release.istiod, helm_release.istio-gateway]

  manifest = {
    "apiVersion" = "rbac.authorization.k8s.io/v1"
    "kind" = "ClusterRole"
    "metadata" = {
      "labels" = {
        "app" = "kiali"
        "app.kubernetes.io/instance" = "kiali"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "kiali"
        "app.kubernetes.io/part-of" = "kiali"
        "app.kubernetes.io/version" = "v1.82.0"
        "helm.sh/chart" = "kiali-server-1.82.0"
        "version" = "v1.82.0"
      }
      "name" = "kiali"
    }
    "rules" = [
      {
        "apiGroups" = [
          "",
        ]
        "resources" = [
          "configmaps",
          "endpoints",
          "pods/log",
        ]
        "verbs" = [
          "get",
          "list",
          "watch",
        ]
      },
      {
        "apiGroups" = [
          "",
        ]
        "resources" = [
          "namespaces",
          "pods",
          "replicationcontrollers",
          "services",
        ]
        "verbs" = [
          "get",
          "list",
          "watch",
          "patch",
        ]
      },
      {
        "apiGroups" = [
          "",
        ]
        "resources" = [
          "pods/portforward",
        ]
        "verbs" = [
          "create",
          "post",
        ]
      },
      {
        "apiGroups" = [
          "extensions",
          "apps",
        ]
        "resources" = [
          "daemonsets",
          "deployments",
          "replicasets",
          "statefulsets",
        ]
        "verbs" = [
          "get",
          "list",
          "watch",
          "patch",
        ]
      },
      {
        "apiGroups" = [
          "batch",
        ]
        "resources" = [
          "cronjobs",
          "jobs",
        ]
        "verbs" = [
          "get",
          "list",
          "watch",
          "patch",
        ]
      },
      {
        "apiGroups" = [
          "networking.istio.io",
          "security.istio.io",
          "extensions.istio.io",
          "telemetry.istio.io",
          "gateway.networking.k8s.io",
        ]
        "resources" = [
          "*",
        ]
        "verbs" = [
          "get",
          "list",
          "watch",
          "create",
          "delete",
          "patch",
        ]
      },
      {
        "apiGroups" = [
          "apps.openshift.io",
        ]
        "resources" = [
          "deploymentconfigs",
        ]
        "verbs" = [
          "get",
          "list",
          "watch",
          "patch",
        ]
      },
      {
        "apiGroups" = [
          "project.openshift.io",
        ]
        "resources" = [
          "projects",
        ]
        "verbs" = [
          "get",
        ]
      },
      {
        "apiGroups" = [
          "route.openshift.io",
        ]
        "resources" = [
          "routes",
        ]
        "verbs" = [
          "get",
        ]
      },
      {
        "apiGroups" = [
          "authentication.k8s.io",
        ]
        "resources" = [
          "tokenreviews",
        ]
        "verbs" = [
          "create",
        ]
      },
    ]
  }
}

resource "kubernetes_manifest" "clusterrolebinding_kiali" {
  depends_on = [helm_release.istio-base, helm_release.istiod, helm_release.istio-gateway]

  manifest = {
    "apiVersion" = "rbac.authorization.k8s.io/v1"
    "kind" = "ClusterRoleBinding"
    "metadata" = {
      "labels" = {
        "app" = "kiali"
        "app.kubernetes.io/instance" = "kiali"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "kiali"
        "app.kubernetes.io/part-of" = "kiali"
        "app.kubernetes.io/version" = "v1.82.0"
        "helm.sh/chart" = "kiali-server-1.82.0"
        "version" = "v1.82.0"
      }
      "name" = "kiali"
    }
    "roleRef" = {
      "apiGroup" = "rbac.authorization.k8s.io"
      "kind" = "ClusterRole"
      "name" = "kiali"
    }
    "subjects" = [
      {
        "kind" = "ServiceAccount"
        "name" = "kiali"
        "namespace" = "istio-system"
      },
    ]
  }
}

resource "kubernetes_manifest" "role_istio_system_kiali_controlplane" {
  depends_on = [helm_release.istio-base, helm_release.istiod, helm_release.istio-gateway]

  manifest = {
    "apiVersion" = "rbac.authorization.k8s.io/v1"
    "kind" = "Role"
    "metadata" = {
      "labels" = {
        "app" = "kiali"
        "app.kubernetes.io/instance" = "kiali"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "kiali"
        "app.kubernetes.io/part-of" = "kiali"
        "app.kubernetes.io/version" = "v1.82.0"
        "helm.sh/chart" = "kiali-server-1.82.0"
        "version" = "v1.82.0"
      }
      "name" = "kiali-controlplane"
      "namespace" = "istio-system"
    }
    "rules" = [
      {
        "apiGroups" = [
          "",
        ]
        "resourceNames" = [
          "cacerts",
          "istio-ca-secret",
        ]
        "resources" = [
          "secrets",
        ]
        "verbs" = [
          "get",
          "list",
          "watch",
        ]
      },
    ]
  }
}

resource "kubernetes_manifest" "rolebinding_istio_system_kiali_controlplane" {
  depends_on = [helm_release.istio-base, helm_release.istiod, helm_release.istio-gateway]

  manifest = {
    "apiVersion" = "rbac.authorization.k8s.io/v1"
    "kind" = "RoleBinding"
    "metadata" = {
      "labels" = {
        "app" = "kiali"
        "app.kubernetes.io/instance" = "kiali"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "kiali"
        "app.kubernetes.io/part-of" = "kiali"
        "app.kubernetes.io/version" = "v1.82.0"
        "helm.sh/chart" = "kiali-server-1.82.0"
        "version" = "v1.82.0"
      }
      "name" = "kiali-controlplane"
      "namespace" = "istio-system"
    }
    "roleRef" = {
      "apiGroup" = "rbac.authorization.k8s.io"
      "kind" = "Role"
      "name" = "kiali-controlplane"
    }
    "subjects" = [
      {
        "kind" = "ServiceAccount"
        "name" = "kiali"
        "namespace" = "istio-system"
      },
    ]
  }
}

resource "kubernetes_manifest" "service_istio_system_kiali" {
  depends_on = [helm_release.istio-base, helm_release.istiod, helm_release.istio-gateway]

  manifest = {
    "apiVersion" = "v1"
    "kind" = "Service"
    "metadata" = {
      "labels" = {
        "app" = "kiali"
        "app.kubernetes.io/instance" = "kiali"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "kiali"
        "app.kubernetes.io/part-of" = "kiali"
        "app.kubernetes.io/version" = "v1.82.0"
        "helm.sh/chart" = "kiali-server-1.82.0"
        "version" = "v1.82.0"
      }
      "name" = "kiali"
      "namespace" = "istio-system"
    }
    "spec" = {
      "ports" = [
        {
          "appProtocol" = "http"
          "name" = "http"
          "port" = 20001
          "protocol" = "TCP"
        },
        {
          "appProtocol" = "http"
          "name" = "http-metrics"
          "port" = 9090
          "protocol" = "TCP"
        },
      ]
      "selector" = {
        "app.kubernetes.io/instance" = "kiali"
        "app.kubernetes.io/name" = "kiali"
      }
    }
  }
}

resource "kubernetes_manifest" "deployment_istio_system_kiali" {
  depends_on = [helm_release.istio-base, helm_release.istiod, helm_release.istio-gateway]

  manifest = {
    "apiVersion" = "apps/v1"
    "kind" = "Deployment"
    "metadata" = {
      "labels" = {
        "app" = "kiali"
        "app.kubernetes.io/instance" = "kiali"
        "app.kubernetes.io/managed-by" = "Helm"
        "app.kubernetes.io/name" = "kiali"
        "app.kubernetes.io/part-of" = "kiali"
        "app.kubernetes.io/version" = "v1.82.0"
        "helm.sh/chart" = "kiali-server-1.82.0"
        "version" = "v1.82.0"
      }
      "name" = "kiali"
      "namespace" = "istio-system"
    }
    "spec" = {
      "replicas" = 1
      "selector" = {
        "matchLabels" = {
          "app.kubernetes.io/instance" = "kiali"
          "app.kubernetes.io/name" = "kiali"
        }
      }
      "strategy" = {
        "rollingUpdate" = {
          "maxSurge" = 1
          "maxUnavailable" = 1
        }
        "type" = "RollingUpdate"
      }
      "template" = {
        "metadata" = {
          "annotations" = {
            "checksum/config" = "19405f971010f37081fa485f6c53616dfdf97d1d372e4ac9e09434cf0e71f02d"
            "kiali.io/dashboards" = "go,kiali"
            "prometheus.io/port" = "9090"
            "prometheus.io/scrape" = "true"
          }
          "labels" = {
            "app" = "kiali"
            "app.kubernetes.io/instance" = "kiali"
            "app.kubernetes.io/managed-by" = "Helm"
            "app.kubernetes.io/name" = "kiali"
            "app.kubernetes.io/part-of" = "kiali"
            "app.kubernetes.io/version" = "v1.82.0"
            "helm.sh/chart" = "kiali-server-1.82.0"
            "sidecar.istio.io/inject" = "false"
            "version" = "v1.82.0"
          }
          "name" = "kiali"
        }
        "spec" = {
          "containers" = [
            {
              "command" = [
                "/opt/kiali/kiali",
                "-config",
                "/kiali-configuration/config.yaml",
              ]
              "env" = [
                {
                  "name" = "ACTIVE_NAMESPACE"
                  "valueFrom" = {
                    "fieldRef" = {
                      "fieldPath" = "metadata.namespace"
                    }
                  }
                },
                {
                  "name" = "LOG_LEVEL"
                  "value" = "info"
                },
                {
                  "name" = "LOG_FORMAT"
                  "value" = "text"
                },
                {
                  "name" = "LOG_TIME_FIELD_FORMAT"
                  "value" = "2006-01-02T15:04:05Z07:00"
                },
                {
                  "name" = "LOG_SAMPLER_RATE"
                  "value" = "1"
                },
              ]
              "image" = "quay.io/kiali/kiali:v1.82"
              "imagePullPolicy" = "Always"
              "livenessProbe" = {
                "httpGet" = {
                  "path" = "/kiali/healthz"
                  "port" = "api-port"
                  "scheme" = "HTTP"
                }
                "initialDelaySeconds" = 5
                "periodSeconds" = 30
              }
              "name" = "kiali"
              "ports" = [
                {
                  "containerPort" = 20001
                  "name" = "api-port"
                },
                {
                  "containerPort" = 9090
                  "name" = "http-metrics"
                },
              ]
              "readinessProbe" = {
                "httpGet" = {
                  "path" = "/kiali/healthz"
                  "port" = "api-port"
                  "scheme" = "HTTP"
                }
                "initialDelaySeconds" = 5
                "periodSeconds" = 30
              }
              "resources" = {
                "limits" = {
                  "memory" = "1Gi"
                }
                "requests" = {
                  "cpu" = "10m"
                  "memory" = "64Mi"
                }
              }
              "securityContext" = {
                "allowPrivilegeEscalation" = false
                "capabilities" = {
                  "drop" = [
                    "ALL",
                  ]
                }
                "privileged" = false
                "readOnlyRootFilesystem" = true
                "runAsNonRoot" = true
              }
              "volumeMounts" = [
                {
                  "mountPath" = "/kiali-configuration"
                  "name" = "kiali-configuration"
                },
                {
                  "mountPath" = "/kiali-cert"
                  "name" = "kiali-cert"
                },
                {
                  "mountPath" = "/kiali-secret"
                  "name" = "kiali-secret"
                },
                {
                  "mountPath" = "/kiali-cabundle"
                  "name" = "kiali-cabundle"
                },
              ]
            },
          ]
          "serviceAccountName" = "kiali"
          "volumes" = [
            {
              "configMap" = {
                "name" = "kiali"
              }
              "name" = "kiali-configuration"
            },
            {
              "name" = "kiali-cert"
              "secret" = {
                "optional" = true
                "secretName" = "istio.kiali-service-account"
              }
            },
            {
              "name" = "kiali-secret"
              "secret" = {
                "optional" = true
                "secretName" = "kiali"
              }
            },
            {
              "configMap" = {
                "name" = "kiali-cabundle"
                "optional" = true
              }
              "name" = "kiali-cabundle"
            },
          ]
        }
      }
    }
  }
}
