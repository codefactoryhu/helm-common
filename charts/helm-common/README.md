# helm-common

![Version: 0.0.1](https://img.shields.io/badge/Version-0.0.1-informational?style=flat-square) ![Type: library](https://img.shields.io/badge/Type-library-informational?style=flat-square)
[![Contributor Covenant](https://img.shields.io/badge/Contributor%20Covenant-2.0-4baaaa.svg)](./CODE_OF_CONDUCT.md)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/cf-common)](https://artifacthub.io/packages/search?repo=cf-common)

A Helm chart template for creating helm charts

**Homepage:** <https://github.com/codefactoryhu/helm-common>

Code Factory's `helm-common` is a **developer friendly DevOps solution** to **deploy** any application or service into **Kubernetes**. It's based on inidustrial standards and contains many years of experience developing and shipping real life services. It highly supports microservice architecture and semantic versioning. The template system of `helm-common` is a really powerfool tool, which allows you to ship your products even faster.

Code Factory's `helm-common` is a very good place to start your jurney with Kubernetes and Helm, or if you want to have a standardized deployment solution.

If you have any question about this tool, or if you have any suggestion how to improve it, open a new issue or PR, and let's discuss it!

Happy helming!

## Usage
**Create a new project**

To use helm-common you need a separate project for managing your micro services with helm. Let's call it `helm`.

In order to maximize the power of `helm-common` we recommend to use a similar folder structure like this.
```
project_name/
  charts/
    ms1/
      templates/
        deployment.yaml
        ...
      Chart.yaml
      values.yaml
    ms2/
    ...
  templates/
    ingress.yaml
    secret.yaml
    ...
  Chart.yaml
  values.yaml
.gitignore
.releaserc.yaml
.gitlab-ci.yaml
values-dev.yaml
values-test.yaml
values-prod.yaml
```

Set as dependency for a new or an existing chart:
```
# project_name/Chart.yaml
...
dependencies:
- name: "helm-common"
  version: 0.0.1
  repository: "@arti_internal" # or "https://domain.tld/artifactory/helm-virtual"
...
```

@arti_internal should be added as a helm repo:

`helm repo add @arti_internal https://domain.tld/artifactory/helm-virtual`

Now you can include the following templates in your template yaml files and configure their behaviour with the `values.yaml`:
#### configmap.yaml
```
{{- template "common.app-env-configmap" . -}}
# configmap for environment variables in your deployment
```

#### deployment.yaml
```
{{- template "common.deployment" . -}}
# deployment for your microservice
```

#### ingress.yaml
```
{{- template "common.ingress" . -}}
# ingress for your services
```

#### secret.yaml
```
{{- template "common.app-env-secret" . --}}
# secret for environment variables in your deployment
```

#### service.yaml
```
{{- template "common.service" . -}}
# service for your deployment
```

#### cronjob.yaml
```
{{- template "common.cronjob" . -}}
# cronjob for your microservice
```

## Roles of the files
- root level values-*.yaml files contain the envionment specific values, like evironment variables, ingress config, etc.
- project_name/Chart.yaml contains the ubrella chart info about the project.
```yaml
apiVersion: v2
appVersion: 1.0.1
description: Some description
name: project_name
type: application
version: 1.0.1
dependencies:
  - name: helm-common
    version: 1.0.0
    repository: "@arti_internal"
  - name: ms1
    version: ">=0.0.0-0"
  - name: ms2
    version: ">=0.0.0-0"
```
- `project_name/values.yaml` contains the project related values, like ingress config, resource allocations, imagePullSecrets, stb.
- `project_name/templates` folder contains yaml files, which uses `helm-common` templates. See above.
- `project_name/charts` folder contains the micro service configurations
- `project_name/charts/templates` similar to `project_name/templates`
- `project_name/chart/values.yaml` contains the values to a specific microservice, eg. resource allocation, secret name, environment variables, images name, version, liveness/readiness settings, etc.
- `project_name/chart/Chart.yaml` contains the chart data of a specific microservice.
```yaml
apiVersion: v2
appVersion: 1.0.0
description: Some description
name: ms1
type: application
version: 1.0.0
dependencies:
  - name: "helm-common"
    version: 1.0.0
    repository: "@arti_internal"
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Configure affinity |
| annotations | object | `{}` | Configure annotations for the deployment and service |
| appEnvConfigMap | object | `{"annotations":{},"name":"app-env-config-map"}` | Configure configmap for env vars See `env.configMap` for more |
| appEnvSecret.name | string | `"app-env-secret"` | Name of the secret for sensitive env vars (It will be removed in future versions.) See `env.secret` for more |
| application.args | string | `nil` | Set args for the application container <br> https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#notes |
| application.command | string | `nil` | Set command for the application container <br> https://kubernetes.io/docs/tasks/inject-data-application/define-command-argument-container/#notes |
| application.lifecycle | string | `nil` | Set postStart and preStop hook for the application container <br> https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks <br> https://kubernetes.io/docs/tasks/configure-pod-container/attach-handler-lifecycle-event/#define-poststart-and-prestop-handlers |
| application.liveness | object | `{"command":null,"enabled":true,"failureThreshold":3,"host":null,"httpHeaders":null,"initialDelaySeconds":0,"path":"/health","periodSeconds":20,"port":9000,"scheme":"HTTP","timeoutSeconds":1,"type":"httpGet"}` | Configure the health check for the application |
| application.liveness.command | string | `nil` | Liveness check http headers (used only probe type exec) |
| application.liveness.enabled | bool | `true` | Set false to disable liveness probe |
| application.liveness.failureThreshold | int | `3` | Liveness check failureThreshold |
| application.liveness.host | string | `nil` | Liveness check host (used only probe type httpGet and tcpSocket) |
| application.liveness.httpHeaders | string | `nil` | Liveness check http headers (used only probe type httpGet) |
| application.liveness.initialDelaySeconds | int | `0` | Liveness check initialDelaySeconds |
| application.liveness.path | string | `"/health"` | Liveness check endpoint (used only probe type httpGet) |
| application.liveness.periodSeconds | int | `20` | Liveness check periodSeconds |
| application.liveness.port | int | `9000` | Liveness check port (used only probe type httpGet and tcpSocket) |
| application.liveness.scheme | string | `"HTTP"` | Liveness check Scheme to use for connecting to the host. Defaults to HTTP. (used only probe type httpGet) |
| application.liveness.timeoutSeconds | int | `1` | Liveness check timeoutSeconds |
| application.liveness.type | string | `"httpGet"` | Valid probe types are: [httpGet](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-liveness-http-request), [tcpSocket](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-tcp-liveness-probe), [exec](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-liveness-command) |
| application.managementPort | int | `9000` | The mangement port the application, where the metrics, liveness and readiness is reachable |
| application.readiness | object | `{"command":null,"enabled":true,"failureThreshold":3,"host":null,"httpHeaders":null,"initialDelaySeconds":0,"path":"/health","periodSeconds":10,"port":9000,"scheme":"HTTP","successThreshold":1,"timeoutSeconds":1,"type":"httpGet"}` | Configure the ready check for the application |
| application.readiness.command | string | `nil` | Readiness check http headers (used only probe type exec) |
| application.readiness.enabled | bool | `true` | Set false to disable readiness probe |
| application.readiness.failureThreshold | int | `3` | Readiness check failureThreshold |
| application.readiness.host | string | `nil` | Readiness check host (used only probe type httpGet and tcpSocket) |
| application.readiness.httpHeaders | string | `nil` | Readiness check http headers (used only probe type httpGet) |
| application.readiness.initialDelaySeconds | int | `0` | Readiness check initialDelaySeconds |
| application.readiness.path | string | `"/health"` | Readiness check endpoint (used only probe type httpGet) |
| application.readiness.periodSeconds | int | `10` | Readiness check periodSeconds |
| application.readiness.port | int | `9000` | Readiness check port (used only probe type httpGet and tcpSocket) |
| application.readiness.scheme | string | `"HTTP"` | Readiness check Scheme to use for connecting to the host. Defaults to HTTP. (used only probe type httpGet) |
| application.readiness.successThreshold | int | `1` | Readiness check successThreshold |
| application.readiness.timeoutSeconds | int | `1` | Readiness check timeoutSeconds |
| application.readiness.type | string | `"httpGet"` | Valid probe types are: [httpGet](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-liveness-http-request), [tcpSocket](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-tcp-liveness-probe), [exec](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-liveness-command) |
| application.serverPort | int | `8000` | The port where the application listens |
| application.startupProbe | object | `{"command":null,"enabled":true,"failureThreshold":30,"host":null,"httpHeaders":null,"initialDelaySeconds":0,"path":"/health","periodSeconds":10,"port":9000,"scheme":"HTTP","timeoutSeconds":1,"type":"httpGet"}` | Configure the startup check for the application |
| application.startupProbe.command | string | `nil` | Startup check http headers (used only probe type exec) |
| application.startupProbe.enabled | bool | `true` | Set false to disable startup probe |
| application.startupProbe.failureThreshold | int | `30` | Startup check failureThreshold |
| application.startupProbe.host | string | `nil` | Startup check host (used only probe type httpGet and tcpSocket) |
| application.startupProbe.httpHeaders | string | `nil` | Startup check http headers (used only probe type httpGet) |
| application.startupProbe.initialDelaySeconds | int | `0` | Startup check initialDelaySeconds |
| application.startupProbe.path | string | `"/health"` | Startup check endpoint (used only probe type httpGet) |
| application.startupProbe.periodSeconds | int | `10` | Startup check periodSeconds |
| application.startupProbe.port | int | `9000` | Startup check port (used only probe type httpGet and tcpSocket) |
| application.startupProbe.scheme | string | `"HTTP"` | Startup check Scheme to use for connecting to the host. Defaults to HTTP. (used only probe type httpGet) |
| application.startupProbe.timeoutSeconds | int | `1` | Startup check timeoutSeconds |
| application.startupProbe.type | string | `"httpGet"` | Valid probe types are: [httpGet](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-liveness-http-request), [tcpSocket](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-tcp-liveness-probe), [exec](https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes/#define-a-liveness-command) |
| application.terminationGracePeriodSeconds | string | `nil` | Configure time to wait until the pod is killed [more](https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#hook-handler-execution) |
| cronJob.concurrencyPolicy | string | `"Allow"` | [concurrency-policy](https://kubernetes.io/docs/tasks/job/automated-tasks-with-cron-jobs/#concurrency-policy) |
| cronJob.failedJobsHistoryLimit | int | `1` | [jobs-history-limits](https://kubernetes.io/docs/tasks/job/automated-tasks-with-cron-jobs/#jobs-history-limits) |
| cronJob.job.activeDeadlineSeconds | string | `nil` | [job-termination-and-cleanup](https://kubernetes.io/docs/concepts/workloads/controllers/job/#job-termination-and-cleanup) |
| cronJob.job.backoffLimit | int | `6` | [pod-backoff-failure-policy](https://kubernetes.io/docs/concepts/workloads/controllers/job/#pod-backoff-failure-policy) |
| cronJob.job.completions | int | `1` | [parallel-jobs](https://kubernetes.io/docs/concepts/workloads/controllers/job/#parallel-jobs) |
| cronJob.job.parallelism | int | `1` | [parallel-jobs](https://kubernetes.io/docs/concepts/workloads/controllers/job/#parallel-jobs) |
| cronJob.job.podRestartPolicy | string | `"OnFailure"` | Supported values: "OnFailure", "Never" |
| cronJob.job.ttlSecondsAfterFinished | string | `nil` | [ttl-mechanism-for-finished-jobs](https://kubernetes.io/docs/concepts/workloads/controllers/job/#ttl-mechanism-for-finished-jobs) |
| cronJob.schedule | string | `"@daily"` | [schedule](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/#cron-schedule-syntax) |
| cronJob.startingDeadlineSeconds | string | `nil` | [starting-deadline](https://kubernetes.io/docs/tasks/job/automated-tasks-with-cron-jobs/#starting-deadline) |
| cronJob.successfulJobsHistoryLimit | int | `3` | [jobs-history-limits](https://kubernetes.io/docs/tasks/job/automated-tasks-with-cron-jobs/#jobs-history-limits) |
| cronJob.suspend | bool | `false` | [suspend](https://kubernetes.io/docs/tasks/job/automated-tasks-with-cron-jobs/#suspend) |
| defaultIpPool | bool | `false` | Use 192.168.x.x IP for the pod instead of reserved IPs for the application. It will be removed after moving to the NSXT clusters. |
| deployment.minReadySeconds | int | `0` | [min-ready-seconds](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#min-ready-seconds) |
| deployment.paused | bool | `false` | [paused](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#paused) |
| deployment.progressDeadlineSeconds | int | `600` | [progress-deadline-seconds](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#progress-deadline-seconds) |
| deployment.revisionHistoryLimit | int | `3` | [revision-history-limit](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#revision-history-limit) |
| deployment.strategy.rollingUpdate.maxSurge | string | `"25%"` | [max-surge](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#max-surge) |
| deployment.strategy.rollingUpdate.maxUnavailable | string | `"25%"` | [max-unavailable](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#max-unavailable) |
| deployment.strategy.type | string | `"RollingUpdate"` | [strategy](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#strategy) |
| env.configMap | object | `{}` | environment variables stored in configmap See 'appEnvConfigMap' for configuring the ConfigMap object |
| env.normal | object | `{"LOG_LEVEL_APP":"INFO","MANAGEMENT_PORT":9000,"SERVER_PORT":8000}` | Environment variable variables |
| env.secret | object | `{}` | sensitive environment variables, if they should be. (It will be removed in future versions.) See 'appEnvSecret' for configuring the Secret object |
| env.vault | object | `{}` | environment variables stored in vault See https://banzaicloud.com/products/bank-vaults/ |
| extraInitContainers | string | `nil` | Configure extra volume mounts for the init containers <br> [Example](chart-test/tests/deployment/values-extra-init-containers.yaml) |
| extraVolumeMounts | string | `nil` | Configure extra volume mounts for the application container <br> [Example](chart-test/tests/deployment/values-extra-init-containers.yaml) |
| extraVolumes | string | `nil` | Configure extra volumes for (init)containers <br> [Example](chart-test/tests/deployment/values-extra-init-containers.yaml) |
| fullnameOverride | string | `""` |  |
| global.serviceAccountName | string | `"default"` | The name of the service account who runs the pod(s) |
| global.vaultAddress | string | `"https://vault-dev.domain.tld"` | The address of HashiCorp Vault server |
| image | object | `{"pullPolicy":"IfNotPresent","repository":"nginx","tag":"latest"}` | Set the image properties of the application-container |
| imagePullSecrets | list | `[{"name":"myregistrykey"}]` | Pull secret for K8S to get the image |
| ingress.enabled | bool | `false` | Set ingerss object enabled |
| ingress.hosts | list | `["{{ .Release.Namespace }}"]` | List of ingress hosts |
| ingress.ingressClass | string | `"{{ .Release.Namespace }}-ingress"` | Name of the ingressClass. Override only if your ingress controller uses different ingress class than the default by convention (NAMESPACE-ingress) |
| ingress.paths | list | `[{"backend":{"serviceName":"{{ .Release.Namespace }}-service-name","servicePort":8000},"path":"/"}]` | List of ingress paths |
| metrics | object | `{"enabled":true,"path":"/metrics","port":9000}` | Configure metrics for Prometheus |
| nameOverride | string | `""` |  |
| nodeSelector | object | `{}` | Configure node selectors |
| podAnnotations | object | `{}` | Configure annotations for the pod |
| replicaCount | int | `1` | The number of desired replicas of the deployment |
| resources | object | `{}` | Configure resources for the container and init-containers. Example: `{"limits":{"cpu":"100m","memory":"128Mi"},"requests":{"cpu":"100m","memory":"128Mi"}}` |
| service | object | `{"port":8000,"type":"ClusterIP"}` | Configure service |
| tolerations | list | `[]` | Configure tolerations |

## Requirements

Kubernetes: `>= 1.21.6`

## Source Code

* <https://github.com/codefactoryhu/helm-common>
#### [Contributing](./CONTRIBUTING.md)

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Zoltan Kadlecsik | <kadlecsik.zoltan@codefactory.hu> |  |
| Zoltan Vigh | <vigh.zoltan@codefactory.hu> |  |
| Abraham Szilagyi | <szilagyi.abraham@codefactory.hu> |  |
| Zoltan Magyar | <magyar.zoltan@codefactory.hu> |  |

----------------------------------------------
Autogenerated from chart metadata using [helm-docs v1.10.0](https://github.com/norwoodj/helm-docs/releases/v1.10.0)

