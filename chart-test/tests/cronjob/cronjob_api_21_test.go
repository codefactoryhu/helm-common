package cronjob

import (
	"errors"
	"github.com/gruntwork-io/terratest/modules/logger"
	batchV1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
)

func TestCronJobBasicApi21(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	releaseName, cronJob := givenACronJobTemplateWithHelmApi21(t, assertions, map[string]string{})

	assertions.Equal(releaseName+"-chart-test", cronJob.Name)

	assertions.Equal(batchV1.AllowConcurrent, cronJob.Spec.ConcurrencyPolicy)
	assertions.Equal(int32(1), *cronJob.Spec.FailedJobsHistoryLimit)
	assertions.Equal("@daily", cronJob.Spec.Schedule)
	assertions.Nil(cronJob.Spec.StartingDeadlineSeconds)
	assertions.Equal(int32(3), *cronJob.Spec.SuccessfulJobsHistoryLimit)
	assertions.False(*cronJob.Spec.Suspend)

	assertions.Nil(cronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds)
	assertions.Equal(int32(6), *cronJob.Spec.JobTemplate.Spec.BackoffLimit)
	assertions.Equal(int32(1), *cronJob.Spec.JobTemplate.Spec.Completions)
	assertions.Equal(int32(1), *cronJob.Spec.JobTemplate.Spec.Parallelism)
	assertions.Nil(cronJob.Spec.JobTemplate.Spec.TTLSecondsAfterFinished)

	annotations := map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   "9000",
		"prometheus.io/path":   "/metrics",
	}
	assertions.Equal(annotations, cronJob.Spec.JobTemplate.Spec.Template.Annotations)

	assertions.Equal(1, len(cronJob.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets))
	assertions.Equal("myregistrykey", cronJob.Spec.JobTemplate.Spec.Template.Spec.ImagePullSecrets[0].Name)
	assertions.Equal("default", cronJob.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName)

	assertions.Empty(cronJob.Spec.JobTemplate.Spec.Template.Spec.Volumes)
	assertions.Empty(cronJob.Spec.JobTemplate.Spec.Template.Spec.InitContainers)

	containers := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers
	assertions.Equal(len(containers), 1)
	container := containers[0]
	expectedContainerImage := "nginx:latest"
	assertions.Equal(expectedContainerImage, container.Image)
	assertions.Equal("chart-test", container.Name)
	assertions.Equal(v1.PullIfNotPresent, container.ImagePullPolicy)

	envVars := container.Env
	assertions.Equal(3, len(envVars))
	assertions.Contains(envVars, v1.EnvVar{Name: "LOG_LEVEL_APP", Value: "INFO"})
	assertions.Contains(envVars, v1.EnvVar{Name: "MANAGEMENT_PORT", Value: "9000"})
	assertions.Contains(envVars, v1.EnvVar{Name: "SERVER_PORT", Value: "8000"})

	assertions.Empty(container.VolumeMounts)

	ports := container.Ports
	assertions.Equal(0, len(ports))

	assertions.Empty(container.LivenessProbe)
	assertions.Empty(container.ReadinessProbe)

	assertions.Empty(container.Resources)

}

func givenACronJobTemplateWithHelmApi21(t *testing.T, require *require.Assertions, values map[string]string) (string, batchV1.CronJob) {
	helmChartPath, err := filepath.Abs("../../")
	releaseName := "helm-basic"
	require.NoError(err)

	namespaceName := "medieval-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues:      values,
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/cronjob.yaml"}, "--kube-version=v1.20.0")

	var cronJob batchV1.CronJob
	helm.UnmarshalK8SYaml(t, output, &cronJob)
	return releaseName, cronJob
}

func TestCronJobCustomServiceAccountApi21(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"global.serviceAccountName": "customSA",
	}
	_, cronJob := givenACronJobTemplateWithHelmApi21(t, assertions, values)

	assertions.Equal("customSA", cronJob.Spec.JobTemplate.Spec.Template.Spec.ServiceAccountName)
}

func TestCronJobMetricsDisabledApi21(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"metrics.enabled": "false",
	}
	_, cronJob := givenACronJobTemplateWithHelmApi21(t, assertions, values)

	annotations := map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   "9000",
		"prometheus.io/path":   "/metrics",
	}
	for annotationKey := range annotations {
		assertions.Empty(cronJob.Spec.JobTemplate.Spec.Template.Annotations[annotationKey], annotationKey+" should be not defined")
	}
}

func TestCronJobMetricsEnabledApi21(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"metrics.enabled": "true",
		"metrics.port":    "7777",
		"metrics.path":    "/met",
	}
	_, cronJob := givenACronJobTemplateWithHelmApi21(t, assertions, values)

	annotations := map[string]string{
		"prometheus.io/scrape": "true",
		"prometheus.io/port":   "7777",
		"prometheus.io/path":   "/met",
	}
	for annotationKey, annotationVal := range annotations {
		assertions.Equal(annotationVal, cronJob.Spec.JobTemplate.Spec.Template.Annotations[annotationKey])
	}
}

func TestCronJobDefaultIpPoolDisabledApi21(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"defaultIpPool": "false",
	}
	_, cronJob := givenACronJobTemplateWithHelmApi21(t, assertions, values)

	annotations := map[string]string{
		"cni.projectcalico.org/ipv4pools": "[\"default-pool\"]",
	}
	for annotationKey := range annotations {
		assertions.Empty(cronJob.Spec.JobTemplate.Spec.Template.Annotations[annotationKey])
	}
}

func TestCronJobDefaultIpPoolEnabledApi21(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"defaultIpPool": "true",
	}
	_, cronJob := givenACronJobTemplateWithHelmApi21(t, assertions, values)

	annotations := map[string]string{
		"cni.projectcalico.org/ipv4pools": "[\"default-pool\"]",
	}
	for annotationKey, annotationVal := range annotations {
		assertions.Equal(annotationVal, cronJob.Spec.JobTemplate.Spec.Template.Annotations[annotationKey])
	}
}

func TestCronJobCustomPodAnnotationApi21(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"podAnnotations.hello":                     "hello",
		"podAnnotations.\"cust\\.annotation/key\"": "custValue",
	}
	_, cronJob := givenACronJobTemplateWithHelmApi21(t, assertions, values)

	annotations := map[string]string{
		"cust.annotation/key": "custValue",
		"hello":               "hello",
	}
	for annotationKey, annotationVal := range annotations {
		assertions.Equal(annotationVal, cronJob.Spec.JobTemplate.Spec.Template.Annotations[annotationKey])
	}
}

func TestCronJobCustomAnnotationApi21(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"annotations.hello":                     "hello",
		"annotations.\"cust\\.annotation/key\"": "custValue",
	}
	_, cronJob := givenACronJobTemplateWithHelmApi21(t, assertions, values)

	annotations := map[string]string{
		"cust.annotation/key": "custValue",
		"hello":               "hello",
	}
	for annotationKey, annotationVal := range annotations {
		assertions.Equal(annotationVal, cronJob.Annotations[annotationKey])
	}
}

func TestCronJobWithNodeSelectorApi21(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"nodeSelector.disktype": "ssd",
	}
	_, cronJob := givenACronJobTemplateWithHelmApi21(t, assertions, values)

	annotations := map[string]string{
		"disktype": "ssd",
	}
	for annotationKey, annotationVal := range annotations {
		assertions.Equal(annotationVal, cronJob.Spec.JobTemplate.Spec.Template.Spec.NodeSelector[annotationKey])
	}
}

func TestCronJobWithTolerationApi21(t *testing.T) {
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
	_, cronJob := givenACronJobTemplateWithHelmApi21(t, assertions, values)

	assertions.Equal(2, len(cronJob.Spec.JobTemplate.Spec.Template.Spec.Tolerations))
	assertions.Contains(cronJob.Spec.JobTemplate.Spec.Template.Spec.Tolerations, tolerationOne)
	assertions.Contains(cronJob.Spec.JobTemplate.Spec.Template.Spec.Tolerations, tolerationTwo)
}

func TestCronJobWithAffinityApi21(t *testing.T) {
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
	_, cronJob := givenACronJobTemplateWithHelmApi21(t, assertions, values)

	assertions.Equal(cronJob.Spec.JobTemplate.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key, affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Key)
	assertions.Equal(cronJob.Spec.JobTemplate.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Operator, affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Operator)
	assertions.Equal(cronJob.Spec.JobTemplate.Spec.Template.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values, affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0].Values)

}

func TestCronJobWithResourceLimitsApi21(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"resources.limits.cpu":      "100m",
		"resources.limits.memory":   "256Mi",
		"resources.requests.cpu":    "100m",
		"resources.requests.memory": "256Mi",
	}
	_, cronJob := givenACronJobTemplateWithHelmApi21(t, assertions, values)

	cpu := resource.MustParse("100m")
	mem := resource.MustParse("256Mi")

	res := v1.ResourceRequirements{
		Limits:   v1.ResourceList{"cpu": cpu, "memory": mem},
		Requests: v1.ResourceList{"cpu": cpu, "memory": mem},
	}
	assertions.Equal(res, cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Resources)

}

func TestCronJobWithSimpleEnvVarsApi21(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{}

	extraEnvVarCount := random.Random(5, 20)
	for i := 0; i < extraEnvVarCount; i++ {
		values["env.normal.ENV_"+strings.ToUpper(random.UniqueId())] = random.UniqueId()
	}

	_, cronJob := givenACronJobTemplateWithHelmApi21(t, assertions, values)

	envVars := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env
	assertions.Equal(extraEnvVarCount+3, len(envVars))

	for key, val := range values {
		envVar := v1.EnvVar{Name: strings.TrimPrefix(key, "env.normal."), Value: val}
		assertions.Contains(envVars, envVar)
	}
}

func TestCronJobWithSecretEnvVarsApi21(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{}

	extraEnvVarCount := random.Random(2, 10)
	for i := 0; i < extraEnvVarCount; i++ {
		values["env.secret.ENV_"+strings.ToUpper(random.UniqueId())] = random.UniqueId()
	}

	_, cronJob := givenACronJobTemplateWithHelmApi21(t, assertions, values)

	envVars := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env
	assertions.Equal(extraEnvVarCount+3, len(envVars))

	for key := range values {
		envKey := strings.TrimPrefix(key, "env.secret.")
		envVar, err := findByKey(envKey, envVars)
		assertions.NoError(err)
		assertions.Equal("app-env-secret", envVar.ValueFrom.SecretKeyRef.Name)
		assertions.Equal(envKey, envVar.ValueFrom.SecretKeyRef.Key)
	}
}

func TestCronJobWithConfigMapEnvVarsApi21(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{}

	extraEnvVarCount := random.Random(2, 10)
	for i := 0; i < extraEnvVarCount; i++ {
		values["env.configMap.ENV_"+strings.ToUpper(random.UniqueId())] = random.UniqueId()
	}

	_, cronJob := givenACronJobTemplateWithHelmApi21(t, assertions, values)

	envVars := cronJob.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Env
	assertions.Equal(extraEnvVarCount+3, len(envVars))

	for key := range values {
		envKey := strings.TrimPrefix(key, "env.configMap.")
		envVar, err := findByKey(envKey, envVars)
		assertions.NoError(err)
		assertions.Equal("app-env-config-map", envVar.ValueFrom.ConfigMapKeyRef.Name)
		assertions.Equal(envKey, envVar.ValueFrom.ConfigMapKeyRef.Key)
	}
}

func TestCronJobCustomValuesApi21(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"cronJob.concurrencyPolicy":           "Replace",
		"cronJob.startingDeadlineSeconds":     "600",
		"cronJob.failedJobsHistoryLimit":      "5",
		"cronJob.successfulJobsHistoryLimit":  "8",
		"cronJob.schedule":                    "* 10 * * *",
		"cronJob.suspend":                     "true",
		"cronJob.job.activeDeadlineSeconds":   "120",
		"cronJob.job.backoffLimit":            "3",
		"cronJob.job.completions":             "2",
		"cronJob.job.parallelism":             "2",
		"cronJob.job.ttlSecondsAfterFinished": "300",
		"cronJob.job.podRestartPolicy":        "Never",
	}
	releaseName, cronJob := givenACronJobTemplateWithHelmApi21(t, assertions, values)

	assertions.Equal(releaseName+"-chart-test", cronJob.Name)

	assertions.Equal(batchV1.ReplaceConcurrent, cronJob.Spec.ConcurrencyPolicy)
	assertions.Equal(int32(5), *cronJob.Spec.FailedJobsHistoryLimit)
	assertions.Equal("* 10 * * *", cronJob.Spec.Schedule)
	assertions.Equal(int64(600), *cronJob.Spec.StartingDeadlineSeconds)
	assertions.Equal(int32(8), *cronJob.Spec.SuccessfulJobsHistoryLimit)
	assertions.True(*cronJob.Spec.Suspend)

	assertions.Equal(int64(120), *cronJob.Spec.JobTemplate.Spec.ActiveDeadlineSeconds)
	assertions.Equal(int32(3), *cronJob.Spec.JobTemplate.Spec.BackoffLimit)
	assertions.Equal(int32(2), *cronJob.Spec.JobTemplate.Spec.Completions)
	assertions.Equal(int32(2), *cronJob.Spec.JobTemplate.Spec.Parallelism)
	assertions.Equal(int32(300), *cronJob.Spec.JobTemplate.Spec.TTLSecondsAfterFinished)

	assertions.Equal(v1.RestartPolicyNever, cronJob.Spec.JobTemplate.Spec.Template.Spec.RestartPolicy)

}

func findByKey(key string, vars []v1.EnvVar) (v1.EnvVar, error) {
	for i := 0; i < len(vars); i++ {
		if vars[i].Name == key {
			return vars[i], nil
		}
	}
	return v1.EnvVar{}, errors.New("key not found: " + key)
}
