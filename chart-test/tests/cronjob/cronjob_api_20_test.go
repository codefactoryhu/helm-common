package cronjob

import (
	"k8s.io/api/batch/v1beta1"
	v1 "k8s.io/api/core/v1"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
)

func TestCronJobBasicApi20(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	releaseName, cronJob := givenACronJobTemplateWithHelmApi20(t, assertions, map[string]string{})

	assertions.Equal(releaseName+"-chart-test", cronJob.Name)

	assertions.Equal(v1beta1.AllowConcurrent, cronJob.Spec.ConcurrencyPolicy)
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

func givenACronJobTemplateWithHelmApi20(t *testing.T, require *require.Assertions, values map[string]string) (string, v1beta1.CronJob) {
	helmChartPath, err := filepath.Abs("../../")
	releaseName := "helm-basic"
	require.NoError(err)

	namespaceName := "medieval-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues:      values,
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/cronjob.yaml"}, "--kube-version=v1.20.0")

	var cronJob v1beta1.CronJob
	helm.UnmarshalK8SYaml(t, output, &cronJob)
	return releaseName, cronJob
}

func TestCronJobCustomValuesApi20(t *testing.T) {
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
	releaseName, cronJob := givenACronJobTemplateWithHelmApi20(t, assertions, values)

	assertions.Equal(releaseName+"-chart-test", cronJob.Name)

	assertions.Equal(v1beta1.ReplaceConcurrent, cronJob.Spec.ConcurrencyPolicy)
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
