package deployment

import (
	"errors"
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"path/filepath"
	"strings"
	"testing"
)

func TestDeploymentProbesHttpGet(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"application.liveness.type":                 "httpGet",
		"application.liveness.path":                 "/liveness",
		"application.liveness.port":                 "9001",
		"application.liveness.host":                 "livenessHost",
		"application.liveness.httpHeaders[0].name":  "Accept",
		"application.liveness.httpHeaders[0].value": "application/json",
		"application.liveness.scheme":               "HTTPS",

		"application.readiness.type":   "httpGet",
		"application.readiness.path":   "/readiness",
		"application.readiness.port":   "9002",
		"application.readiness.host":   "readinessHost",
		"application.readiness.scheme": "HTTP",

		"application.startupProbe.type":                 "httpGet",
		"application.startupProbe.path":                 "/startupProbe",
		"application.startupProbe.port":                 "9003",
		"application.startupProbe.host":                 "startupProbeHost",
		"application.startupProbe.httpHeaders[0].name":  "Accept",
		"application.startupProbe.httpHeaders[0].value": "*/*",
		"application.startupProbe.httpHeaders[1].name":  "X-custom-header",
		"application.startupProbe.httpHeaders[1].value": "hello",
	}

	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)
	deploymentContainers := deployment.Spec.Template.Spec.Containers
	assertions.Equal(len(deploymentContainers), 1)
	container := deploymentContainers[0]

	assertions.Equal("/startupProbe", container.StartupProbe.HTTPGet.Path)
	assertions.Equal("startupProbeHost", container.StartupProbe.HTTPGet.Host)
	assertions.Equal(2, len(container.StartupProbe.HTTPGet.HTTPHeaders))
	assertions.Equal("Accept", container.StartupProbe.HTTPGet.HTTPHeaders[0].Name)
	assertions.Equal("*/*", container.StartupProbe.HTTPGet.HTTPHeaders[0].Value)
	assertions.Equal("X-custom-header", container.StartupProbe.HTTPGet.HTTPHeaders[1].Name)
	assertions.Equal("hello", container.StartupProbe.HTTPGet.HTTPHeaders[1].Value)
	assertions.Equal(v1.URISchemeHTTP, container.StartupProbe.HTTPGet.Scheme)
	assertions.Equal(int32(9003), container.StartupProbe.HTTPGet.Port.IntVal)
	assertions.Equal(int32(10), container.StartupProbe.PeriodSeconds)
	assertions.Equal(int32(1), container.StartupProbe.TimeoutSeconds)
	assertions.Equal(int32(30), container.StartupProbe.FailureThreshold)
	assertions.Empty(container.StartupProbe.TCPSocket)
	assertions.Empty(container.StartupProbe.Exec)

	assertions.Equal("/liveness", container.LivenessProbe.HTTPGet.Path)
	assertions.Equal("livenessHost", container.LivenessProbe.HTTPGet.Host)
	assertions.Equal(1, len(container.LivenessProbe.HTTPGet.HTTPHeaders))
	assertions.Equal("Accept", container.LivenessProbe.HTTPGet.HTTPHeaders[0].Name)
	assertions.Equal("application/json", container.LivenessProbe.HTTPGet.HTTPHeaders[0].Value)
	assertions.Equal(v1.URISchemeHTTPS, container.LivenessProbe.HTTPGet.Scheme)
	assertions.Equal(int32(9001), container.LivenessProbe.HTTPGet.Port.IntVal)
	assertions.Equal(int32(20), container.LivenessProbe.PeriodSeconds)
	assertions.Equal(int32(1), container.LivenessProbe.TimeoutSeconds)
	assertions.Equal(int32(3), container.LivenessProbe.FailureThreshold)
	assertions.Empty(container.LivenessProbe.TCPSocket)
	assertions.Empty(container.LivenessProbe.Exec)

	assertions.Equal("/readiness", container.ReadinessProbe.HTTPGet.Path)
	assertions.Equal("readinessHost", container.ReadinessProbe.HTTPGet.Host)
	assertions.Equal(0, len(container.ReadinessProbe.HTTPGet.HTTPHeaders))
	assertions.Equal(v1.URISchemeHTTP, container.ReadinessProbe.HTTPGet.Scheme)
	assertions.Equal(int32(9002), container.ReadinessProbe.HTTPGet.Port.IntVal)
	assertions.Equal(int32(10), container.ReadinessProbe.PeriodSeconds)
	assertions.Equal(int32(1), container.ReadinessProbe.TimeoutSeconds)
	assertions.Equal(int32(3), container.ReadinessProbe.FailureThreshold)
	assertions.Empty(container.ReadinessProbe.TCPSocket)
	assertions.Empty(container.ReadinessProbe.Exec)
}

func TestDeploymentProbesTcpSocket(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"application.liveness.type": "tcpSocket",
		"application.liveness.port": "9001",
		"application.liveness.host": "livenessHost",

		"application.readiness.type": "tcpSocket",
		"application.readiness.port": "9002",

		"application.startupProbe.type": "tcpSocket",
		"application.startupProbe.port": "9003",
	}

	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)
	deploymentContainers := deployment.Spec.Template.Spec.Containers
	assertions.Equal(len(deploymentContainers), 1)
	container := deploymentContainers[0]

	assertions.Empty(container.StartupProbe.TCPSocket.Host)
	assertions.Equal(int32(9003), container.StartupProbe.TCPSocket.Port.IntVal)
	assertions.Empty(container.StartupProbe.HTTPGet)
	assertions.Empty(container.StartupProbe.Exec)

	assertions.Equal("livenessHost", container.LivenessProbe.TCPSocket.Host)
	assertions.Equal(int32(9001), container.LivenessProbe.TCPSocket.Port.IntVal)
	assertions.Empty(container.LivenessProbe.HTTPGet)
	assertions.Empty(container.LivenessProbe.Exec)

	assertions.Empty(container.ReadinessProbe.TCPSocket.Host)
	assertions.Equal(int32(9002), container.ReadinessProbe.TCPSocket.Port.IntVal)
	assertions.Empty(container.ReadinessProbe.HTTPGet)
	assertions.Empty(container.ReadinessProbe.Exec)
}

func TestDeploymentProbesExec(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"application.liveness.type":       "exec",
		"application.liveness.command[0]": "sh",
		"application.liveness.command[1]": "-c",
		"application.liveness.command[2]": "echo hello",

		"application.readiness.type":       "exec",
		"application.readiness.command[0]": "cat",
		"application.readiness.command[1]": "/tmp/ready",

		"application.startupProbe.type":       "exec",
		"application.startupProbe.command[0]": "ls",
	}

	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)
	deploymentContainers := deployment.Spec.Template.Spec.Containers
	assertions.Equal(len(deploymentContainers), 1)
	container := deploymentContainers[0]

	commandStart := container.StartupProbe.Exec.Command
	assertions.Equal(1, len(commandStart))
	assertions.Equal("ls", commandStart[0])
	assertions.Empty(container.StartupProbe.HTTPGet)
	assertions.Empty(container.StartupProbe.TCPSocket)

	commandLive := container.LivenessProbe.Exec.Command
	assertions.Equal(3, len(commandLive))
	assertions.Equal("sh", commandLive[0])
	assertions.Equal("-c", commandLive[1])
	assertions.Equal("echo hello", commandLive[2])
	assertions.Empty(container.LivenessProbe.HTTPGet)
	assertions.Empty(container.LivenessProbe.TCPSocket)

	commandReady := container.ReadinessProbe.Exec.Command
	assertions.Equal(2, len(commandReady))
	assertions.Equal("cat", commandReady[0])
	assertions.Equal("/tmp/ready", commandReady[1])
	assertions.Empty(container.ReadinessProbe.HTTPGet)
	assertions.Empty(container.ReadinessProbe.TCPSocket)
}

func TestDeploymentProbesDisabled(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"application.liveness.enabled":     "false",
		"application.readiness.enabled":    "false",
		"application.startupProbe.enabled": "false",
	}

	deployment, _, _ := givenADeploymentTemplateWithHelm(t, assertions, values)
	deploymentContainers := deployment.Spec.Template.Spec.Containers
	assertions.Equal(len(deploymentContainers), 1)
	container := deploymentContainers[0]

	assertions.Empty(container.StartupProbe)
	assertions.Empty(container.LivenessProbe)
	assertions.Empty(container.ReadinessProbe)
}

func TestDeploymentInvalidProbeType(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	values := map[string]string{
		"application.liveness.type": "Invalid",
	}

	helmChartPath, err := filepath.Abs("../../")
	assertions.NoError(err)

	options := &helm.Options{
		SetValues:      values,
		KubectlOptions: k8s.NewKubectlOptions("", "", "medieval-"+strings.ToLower(random.UniqueId())),
	}

	_, err = helm.RenderTemplateE(t, options, helmChartPath, "helm-basic", []string{"templates/deployment.yaml"})

	errors.New("Invalid probe type, must be one of (httpGet,tcpSocket,exec)")

	assertions.Error(err)
	assertions.Contains(err.Error(), "Invalid probe type, must be one of (httpGet,tcpSocket,exec)")
}
