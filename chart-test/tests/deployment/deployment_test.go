package deployment

import (
	"errors"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeploymentBasic(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	deployment, releaseName, _ := givenADeploymentTemplateWithHelm(t, assertions, map[string]string{})

	assertions.Equal(releaseName+"-chart-test", deployment.Name)

	assertions.Equal(int32(1), *deployment.Spec.Replicas)
	assertions.Equal(appsv1.DeploymentStrategyType("RollingUpdate"), deployment.Spec.Strategy.Type)
	assertions.Equal("25%", deployment.Spec.Strategy.RollingUpdate.MaxSurge.StrVal)
	assertions.Equal("25%", deployment.Spec.Strategy.RollingUpdate.MaxUnavailable.StrVal)
	assertions.Equal(int32(600), *deployment.Spec.ProgressDeadlineSeconds)
	assertions.Equal(int32(0), deployment.Spec.MinReadySeconds)
	assertions.Equal(int32(3), *deployment.Spec.RevisionHistoryLimit)
	assertions.Equal(false, deployment.Spec.Paused)

	annotations := map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   "9000",
		"prometheus.io/path":   "/metrics",
	}
	labels := map[string]string{
		"app.kubernetes.io/name":     "chart-test",
		"app.kubernetes.io/instance": releaseName,
	}
	assertions.Equal(labels, deployment.Spec.Selector.MatchLabels)
	assertions.Equal(labels, deployment.Spec.Template.Labels)
	assertions.Equal(annotations, deployment.Spec.Template.Annotations)

	assertions.Equal(1, len(deployment.Spec.Template.Spec.ImagePullSecrets))
	assertions.Equal("myregistrykey", deployment.Spec.Template.Spec.ImagePullSecrets[0].Name)
	assertions.Equal("default", deployment.Spec.Template.Spec.ServiceAccountName)

	assertions.Empty(deployment.Spec.Template.Spec.Volumes)
	assertions.Empty(deployment.Spec.Template.Spec.InitContainers)

	deploymentContainers := deployment.Spec.Template.Spec.Containers
	assertions.Equal(len(deploymentContainers), 1)
	container := deploymentContainers[0]
	expectedContainerImage := "nginx:latest"
	assertions.Equal(expectedContainerImage, container.Image)
	assertions.Equal("chart-test", container.Name)
	assertions.Equal(v1.PullIfNotPresent, container.ImagePullPolicy)
	assertions.Nil(container.Command)
	assertions.Nil(container.Args)

	envVars := container.Env
	assertions.Equal(3, len(envVars))
	assertions.Contains(envVars, v1.EnvVar{Name: "LOG_LEVEL_APP", Value: "INFO"})
	assertions.Contains(envVars, v1.EnvVar{Name: "MANAGEMENT_PORT", Value: "9000"})
	assertions.Contains(envVars, v1.EnvVar{Name: "SERVER_PORT", Value: "8000"})

	assertions.Empty(container.VolumeMounts)

	ports := container.Ports
	assertions.Equal(2, len(ports))
	assertions.Contains(ports, v1.ContainerPort{Name: "http", ContainerPort: 8000, Protocol: "TCP"})
	assertions.Contains(ports, v1.ContainerPort{Name: "health-check", ContainerPort: 9000, Protocol: "TCP"})

	assertions.Equal("/health", container.StartupProbe.HTTPGet.Path)
	assertions.Equal(int32(9000), container.StartupProbe.HTTPGet.Port.IntVal)
	assertions.Equal(int32(10), container.StartupProbe.PeriodSeconds)
	assertions.Equal(int32(1), container.StartupProbe.TimeoutSeconds)
	assertions.Equal(int32(30), container.StartupProbe.FailureThreshold)

	assertions.Equal("/health", container.LivenessProbe.HTTPGet.Path)
	assertions.Equal(int32(9000), container.LivenessProbe.HTTPGet.Port.IntVal)
	assertions.Equal(int32(20), container.LivenessProbe.PeriodSeconds)
	assertions.Equal(int32(1), container.LivenessProbe.TimeoutSeconds)
	assertions.Equal(int32(3), container.LivenessProbe.FailureThreshold)

	assertions.Equal("/health", container.ReadinessProbe.HTTPGet.Path)
	assertions.Equal(int32(9000), container.ReadinessProbe.HTTPGet.Port.IntVal)
	assertions.Equal(int32(10), container.ReadinessProbe.PeriodSeconds)
	assertions.Equal(int32(1), container.ReadinessProbe.TimeoutSeconds)
	assertions.Equal(int32(3), container.ReadinessProbe.FailureThreshold)

	assertions.Empty(container.Resources)
	assertions.Empty(container.Lifecycle)

	assertions.Nil(deployment.Spec.Template.Spec.TerminationGracePeriodSeconds)

}

func givenADeploymentTemplateWithHelm(t *testing.T, assertions *require.Assertions, values map[string]string) (deployment appsv1.Deployment, releaseName string, namespaceName string) {
	helmChartPath, err := filepath.Abs("../../")
	releaseName = "helm-basic"
	assertions.NoError(err)

	namespaceName = "medieval-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues:      values,
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/deployment.yaml"})

	helm.UnmarshalK8SYaml(t, output, &deployment)
	return deployment, releaseName, namespaceName
}

func TestDeploymentCustomServiceAccount(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"global.serviceAccountName": "customSA",
	}
	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	assertions.Equal("customSA", deployment.Spec.Template.Spec.ServiceAccountName)
}

func TestDeploymentMetricsDisabled(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"metrics.enabled": "false",
	}
	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	annotations := map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   "9000",
		"prometheus.io/path":   "/metrics",
	}
	for annotationKey := range annotations {
		assertions.Empty(deployment.Spec.Template.Annotations[annotationKey], annotationKey+" should be not defined")
	}
}

func TestDeploymentMetricsEnabled(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"metrics.enabled": "true",
		"metrics.port":    "7777",
		"metrics.path":    "/met",
	}
	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	annotations := map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   "7777",
		"prometheus.io/path":   "/met",
	}
	for annotationKey, annotationVal := range annotations {
		assertions.Equal(annotationVal, deployment.Spec.Template.Annotations[annotationKey])
	}
}

func TestDeploymentDefaultIpPoolDisabled(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"defaultIpPool": "false",
	}
	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	annotations := map[string]string{
		"cni.projectcalico.org/ipv4pools": "[\"default-pool\"]",
	}
	for annotationKey := range annotations {
		assertions.Empty(deployment.Spec.Template.Annotations[annotationKey])
	}
}

func TestDeploymentDefaultIpPoolEnabled(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"defaultIpPool": "true",
	}
	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	annotations := map[string]string{
		"cni.projectcalico.org/ipv4pools": "[\"default-pool\"]",
	}
	for annotationKey, annotationVal := range annotations {
		assertions.Equal(annotationVal, deployment.Spec.Template.Annotations[annotationKey])
	}
}

func TestDeploymentCustomPodAnnotation(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"podAnnotations.hello":                     "hello",
		"podAnnotations.\"cust\\.annotation/key\"": "custValue",
	}
	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	annotations := map[string]string{
		"cust.annotation/key": "custValue",
		"hello":               "hello",
	}
	for annotationKey, annotationVal := range annotations {
		assertions.Equal(annotationVal, deployment.Spec.Template.Annotations[annotationKey])
	}
}

func TestDeploymentCustomAnnotation(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"annotations.hello":                     "hello",
		"annotations.\"cust\\.annotation/key\"": "custValue",
	}
	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	annotations := map[string]string{
		"cust.annotation/key": "custValue",
		"hello":               "hello",
	}
	for annotationKey, annotationVal := range annotations {
		assertions.Equal(annotationVal, deployment.Annotations[annotationKey])
	}
}

func TestDeploymentWithNodeSelector(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"nodeSelector.disktype": "ssd",
	}
	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	annotations := map[string]string{
		"disktype": "ssd",
	}
	for annotationKey, annotationVal := range annotations {
		assertions.Equal(annotationVal, deployment.Spec.Template.Spec.NodeSelector[annotationKey])
	}
}

func TestDeploymentWithToleration(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	tolerationOne := v1.Toleration{
		Key:      "key1",
		Operator: "Equal",
		Value:    "value1",
		Effect:   "NoSchedule",
	}
	tolerationTwo := v1.Toleration{
		Key:      "key2",
		Operator: "Exists",
		Effect:   "NoSchedule",
	}
	values := map[string]string{
		"tolerations[0].key":      "key1",
		"tolerations[0].operator": "Equal",
		"tolerations[0].value":    "value1",
		"tolerations[0].effect":   "NoSchedule",
		"tolerations[1].key":      "key2",
		"tolerations[1].operator": "Exists",
		"tolerations[1].effect":   "NoSchedule",
	}
	logger.Log(t, values)
	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	assertions.Equal(2, len(deployment.Spec.Template.Spec.Tolerations))
	assertions.Contains(deployment.Spec.Template.Spec.Tolerations, tolerationOne)
	assertions.Contains(deployment.Spec.Template.Spec.Tolerations, tolerationTwo)
}

func TestDeploymentWithAffinity(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	affinity := v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      "kubernetes.io/e2e-az-name",
								Operator: "In",
								Values:   []string{"e2e-az1", "e2e-az2"},
							},
						},
					},
				},
			},
		},
	}

	values := map[string]string{
		"affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].key":       "kubernetes.io/e2e-az-name",
		"affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].operator":  "In",
		"affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values[0]": "e2e-az1",
		"affinity.nodeAffinity.requiredDuringSchedulingIgnoredDuringExecution.nodeSelectorTerms[0].matchExpressions[0].values[1]": "e2e-az2",
	}
	logger.Log(t, values)
	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	assertions.Equal(deployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key, affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key)
	assertions.Equal(deployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Operator, affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Operator)
	assertions.Equal(deployment.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values, affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values)

}

func TestDeploymentWithResourceLimits(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"resources.limits.cpu":      "100m",
		"resources.limits.memory":   "256Mi",
		"resources.requests.cpu":    "100m",
		"resources.requests.memory": "256Mi",
	}
	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	cpu := resource.MustParse("100m")
	mem := resource.MustParse("256Mi")

	res := v1.ResourceRequirements{
		Limits:   v1.ResourceList{"cpu": cpu, "memory": mem},
		Requests: v1.ResourceList{"cpu": cpu, "memory": mem},
	}
	assertions.Equal(res, deployment.Spec.Template.Spec.Containers[0].Resources)
}

func TestDeploymentWithSimpleEnvVars(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{}

	extraEnvVarCount := random.Random(5, 20)
	for i := 0; i < extraEnvVarCount; i++ {
		values["env.normal.ENV_"+strings.ToUpper(random.UniqueId())] = random.UniqueId()
	}

	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	envVars := deployment.Spec.Template.Spec.Containers[0].Env
	assertions.Equal(extraEnvVarCount+3, len(envVars))

	for key, val := range values {
		envVar := v1.EnvVar{Name: strings.TrimPrefix(key, "env.normal."), Value: val}
		assertions.Contains(envVars, envVar)
	}
}

func TestDeploymentWithSecretEnvVars(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{}

	extraEnvVarCount := random.Random(2, 10)
	for i := 0; i < extraEnvVarCount; i++ {
		values["env.secret.ENV_"+strings.ToUpper(random.UniqueId())] = random.UniqueId()
	}

	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	envVars := deployment.Spec.Template.Spec.Containers[0].Env
	assertions.Equal(extraEnvVarCount+3, len(envVars))

	for key := range values {
		envKey := strings.TrimPrefix(key, "env.secret.")
		envVar, err := findByKey(envKey, envVars)
		assertions.NoError(err)
		assertions.Equal("app-env-secret", envVar.ValueFrom.SecretKeyRef.Name)
		assertions.Equal(envKey, envVar.ValueFrom.SecretKeyRef.Key)
	}
}

func findByKey(key string, vars []v1.EnvVar) (v1.EnvVar, error) {
	for i := 0; i < len(vars); i++ {
		if vars[i].Name == key {
			return vars[i], nil
		}
	}
	return v1.EnvVar{}, errors.New("key not found: " + key)
}

func TestDeploymentWithConfigMapEnvVars(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{}

	extraEnvVarCount := random.Random(2, 10)
	for i := 0; i < extraEnvVarCount; i++ {
		values["env.configMap.ENV_"+strings.ToUpper(random.UniqueId())] = random.UniqueId()
	}

	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	envVars := deployment.Spec.Template.Spec.Containers[0].Env
	assertions.Equal(extraEnvVarCount+3, len(envVars))

	for key := range values {
		envKey := strings.TrimPrefix(key, "env.configMap.")
		envVar, err := findByKey(envKey, envVars)
		assertions.NoError(err)
		assertions.Equal("app-env-config-map", envVar.ValueFrom.ConfigMapKeyRef.Name)
		assertions.Equal(envKey, envVar.ValueFrom.ConfigMapKeyRef.Key)
	}
}

func TestDeploymentWithTerminationGracePeriod(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"application.terminationGracePeriodSeconds": "60",
	}

	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	assertions.Equal(int64(60), *deployment.Spec.Template.Spec.TerminationGracePeriodSeconds)

}

func TestDeploymentWithVault(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	const (
		vaultAddr             = "https://vault-dev"
		saName                = "ns-sa"
		vaultSecretPathPrefix = "vault:k8s/data/"
		vaultSecretPath       = "internal/aws#AWS_TEST"
		vaultSecretEnvVarName = "AWS_TEST"
	)
	values := map[string]string{
		"global.serviceAccountName":          saName,
		"global.vaultAddress":                vaultAddr,
		"env.vault." + vaultSecretEnvVarName: vaultSecretPath,
	}

	deployment, _, namespaceName := givenADeploymentTemplateWithHelm(t, assertions, values)

	assertions.Equal(saName, deployment.Spec.Template.Spec.ServiceAccountName)

	envVars := deployment.Spec.Template.Spec.Containers[0].Env
	envVar := v1.EnvVar{Name: vaultSecretEnvVarName, Value: vaultSecretPathPrefix + namespaceName + "/" + vaultSecretPath}
	assertions.Contains(envVars, envVar)

	annotations := map[string]string{
		"vault.security.banzaicloud.io/vault-addr": vaultAddr,
		"vault.security.banzaicloud.io/vault-role": namespaceName,
	}
	for annotationKey, annotationVal := range annotations {
		assertions.Equal(annotationVal, deployment.Spec.Template.Annotations[annotationKey])
	}
}

func TestDeploymentWithoutVault(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{}

	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)
	annotations := []string{
		"vault.security.banzaicloud.io/vault-addr",
		"vault.security.banzaicloud.io/vault-role",
	}
	for _, annotation := range annotations {
		assertions.Empty(deployment.Spec.Template.Annotations[annotation])
	}
}

func TestExtraInitContainersAndVolumes(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	helmChartPath, err := filepath.Abs("../../")
	releaseName := "helm-basic"
	assertions.NoError(err)

	namespaceName := "medieval-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		ValuesFiles:    []string{"values-extra-init-containers.yaml"},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(t, output, &deployment)

	containerVolumeMounts := deployment.Spec.Template.Spec.Containers[0].VolumeMounts
	initContainerVolumeMounts := deployment.Spec.Template.Spec.InitContainers[0].VolumeMounts
	volumes := deployment.Spec.Template.Spec.Volumes

	assertions.Equal(1, len(containerVolumeMounts))
	assertions.Equal(1, len(initContainerVolumeMounts))
	assertions.Equal(1, len(volumes))

	volumeMount := v1.VolumeMount{
		Name:             "extra",
		ReadOnly:         false,
		MountPath:        "/extra",
		SubPath:          "",
		MountPropagation: nil,
		SubPathExpr:      "",
	}
	volume := v1.Volume{
		Name: "extra",
		VolumeSource: v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		},
	}
	assertions.Equal(volumeMount, containerVolumeMounts[0])
	assertions.Equal(volumeMount, initContainerVolumeMounts[0])
	assertions.Equal(volume, volumes[0])
}

func TestExtraEnvVarList(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	helmChartPath, err := filepath.Abs("../../")
	releaseName := "helm-basic"
	assertions.NoError(err)

	namespaceName := "medieval-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		ValuesFiles:    []string{"env-var-list.yaml"},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/deployment.yaml"})
	var deployment appsv1.Deployment
	helm.UnmarshalK8SYaml(t, output, &deployment)

	const requiredString = `whitelist:
  - 87426356344D620453F55CF297937679
  - B218691C0C4AE0C2C1C1EB344A57AA93
  - B218691C0C4AE0C2C1C1EB344A57AA77
`

	for i, envVar := range deployment.Spec.Template.Spec.Containers[0].Env {
		if strings.HasPrefix(envVar.Name, "CUSTOM_LIST") {
			logger.Default.Logf(t, "%s", i)
			assertions.Equal(requiredString, envVar.Value)

		}
	}

}

func TestContainerPreStopHook(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"application.lifecycle.preStop.httpGet.path":    "/shutdown",
		"application.lifecycle.preStop.httpGet.port":    "9000",
		"application.lifecycle.preStop.exec.command[0]": "/bin/sh",
		"application.lifecycle.preStop.exec.command[1]": "-c",
		"application.lifecycle.preStop.exec.command[2]": "echo bye",
	}

	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	httpGetAction := deployment.Spec.Template.Spec.Containers[0].Lifecycle.PreStop.HTTPGet
	assertions.Equal("/shutdown", httpGetAction.Path)
	assertions.Equal("9000", httpGetAction.Port.String())
	execAction := deployment.Spec.Template.Spec.Containers[0].Lifecycle.PreStop.Exec.Command
	assertions.Equal("/bin/sh", execAction[0])
	assertions.Equal("-c", execAction[1])
	assertions.Equal("echo bye", execAction[2])

}

func TestDeploymentRecreateStrategy(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"deployment.strategy.type":           "Recreate",
		"deployment.progressDeadlineSeconds": "100",
		"deployment.minReadySeconds":         "1",
		"deployment.revisionHistoryLimit":    "5",
		"deployment.paused":                  "true",
	}

	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)

	assertions.Equal(appsv1.RecreateDeploymentStrategyType, deployment.Spec.Strategy.Type)
	assertions.Nil(deployment.Spec.Strategy.RollingUpdate)
	assertions.Equal(int32(100), *deployment.Spec.ProgressDeadlineSeconds)
	assertions.Equal(int32(1), deployment.Spec.MinReadySeconds)
	assertions.Equal(int32(5), *deployment.Spec.RevisionHistoryLimit)
	assertions.Equal(true, deployment.Spec.Paused)
}

func TestDeploymentInvalidStrategy(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"deployment.strategy.type":           "Invalid",
		"deployment.progressDeadlineSeconds": "100",
		"deployment.minReadySeconds":         "1",
		"deployment.revisionHistoryLimit":    "5",
		"deployment.paused":                  "true",
	}

	helmChartPath, err := filepath.Abs("../../")
	assertions.NoError(err)

	options := &helm.Options{
		SetValues:      values,
		KubectlOptions: k8s.NewKubectlOptions("", "", "medieval-"+strings.ToLower(random.UniqueId())),
	}

	_, err = helm.RenderTemplateE(t, options, helmChartPath, "helm-basic", []string{"templates/deployment.yaml"})

	errors.New("Invalid strategy type, must be one of (RollingUpdate,Recreate)")

	assertions.Error(err)
	assertions.Contains(err.Error(), "Invalid strategy type, must be one of (RollingUpdate,Recreate)")
}
