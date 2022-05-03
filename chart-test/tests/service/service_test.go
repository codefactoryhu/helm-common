package service

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"path/filepath"
	"strings"
	"testing"
)

func givenAServiceTemplateWithHelm(t *testing.T, require *require.Assertions, values map[string]string) (string, v1.Service) {
	helmChartPath, err := filepath.Abs("../../")
	releaseName := "helm-basic"
	require.NoError(err)

	namespaceName := "medieval-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues:      values,
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/service.yaml"})

	var service v1.Service
	helm.UnmarshalK8SYaml(t, output, &service)
	return releaseName, service
}

func TestServiceBasic(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	defaultValues := map[string]string{}
	releaseName, service := givenAServiceTemplateWithHelm(t, require, defaultValues)

	require.Equal(releaseName+"-chart-test", service.Name)
	require.Equal(v1.ServiceType("ClusterIP"), service.Spec.Type)

	servicePorts := service.Spec.Ports
	require.Len(servicePorts, 1)
	require.Equal("http", servicePorts[0].Name)
	require.Equal(v1.Protocol("TCP"), servicePorts[0].Protocol)
	require.Equal(int32(8000), servicePorts[0].Port)
	require.Nil(servicePorts[0].AppProtocol)
	require.Equal(int32(0), servicePorts[0].NodePort)
	require.Equal("http", servicePorts[0].TargetPort.String())

	selector := service.Spec.Selector
	require.Equal("chart-test", selector["app.kubernetes.io/name"])
	require.Equal("helm-basic", selector["app.kubernetes.io/instance"])

}

func TestServiceDifferentPort(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	defaultValues := map[string]string{
		"service.port": "9000",
	}
	_, service := givenAServiceTemplateWithHelm(t, require, defaultValues)

	servicePorts := service.Spec.Ports
	require.Len(servicePorts, 1)
	require.Equal(int32(9000), servicePorts[0].Port)
}

func TestServiceNodePort(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	defaultValues := map[string]string{
		"service.type": "NodePort",
	}
	_, service := givenAServiceTemplateWithHelm(t, require, defaultValues)

	require.Equal(v1.ServiceType("NodePort"), service.Spec.Type)
}

func TestServiceHeadless(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	defaultValues := map[string]string{
		"service.type": "None",
	}
	_, service := givenAServiceTemplateWithHelm(t, require, defaultValues)

	require.Equal(v1.ServiceType("None"), service.Spec.Type)
}
