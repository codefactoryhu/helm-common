package ingress

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	"k8s.io/api/networking/v1beta1"
	"path/filepath"
	"strings"
	"testing"
)

func givenAnIngressTemplateWithHelmApi18(t *testing.T, require *require.Assertions, values map[string]string) (string, string, v1beta1.Ingress) {
	helmChartPath, err := filepath.Abs("../../")
	releaseName := "helm-basic"
	require.NoError(err)

	namespaceName := "medieval-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues:      values,
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/ingress.yaml"}, "--kube-version=v1.18.0")

	var ingress v1beta1.Ingress
	helm.UnmarshalK8SYaml(t, output, &ingress)
	return namespaceName, releaseName, ingress
}

func TestIngressBasicApi18(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	defaultValues := map[string]string{
		"ingress.enabled": "true",
	}
	namespaceName, releaseName, ingress := givenAnIngressTemplateWithHelmApi18(t, assertions, defaultValues)

	assertions.Equal(releaseName+"-chart-test", ingress.Name)
	assertions.NotNil(ingress.Annotations)
	assertions.NotEmpty(ingress.Annotations)
	assertions.Equal(namespaceName+"-ingress", ingress.Annotations["kubernetes.io/ingress.class"])
// TODO: zool megnezni
//	assertions.NotEmpty(ingress.Annotations["nginx.ingress.kubernetes.io/configuration-snippet"])

	assertions.Len(ingress.Spec.TLS, 1)
// TODO: zool megnezni
//	assertions.Len(ingress.Spec.TLS[0].Hosts, 2)

//	assertions.Len(ingress.Spec.Rules, 2)
	for _, ingressRule := range ingress.Spec.Rules {
		assertions.Len(ingressRule.HTTP.Paths, 1)
		assertions.Equal("/", ingressRule.HTTP.Paths[0].Path)
		assertions.Equal(namespaceName+"-service-name", ingressRule.HTTP.Paths[0].Backend.ServiceName)
		assertions.Equal(int32(8000), ingressRule.HTTP.Paths[0].Backend.ServicePort.IntVal)
	}
}

func TestIngressMoreServicePathApi18(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	defaultValues := map[string]string{
		"ingress.enabled":                      "true",
		"ingress.paths[0].path":                "/first",
		"ingress.paths[0].backend.serviceName": "first-service",
		"ingress.paths[1].path":                "/second",
		"ingress.paths[1].backend.serviceName": "second-service",
	}
	_, _, ingress := givenAnIngressTemplateWithHelmApi18(t, assertions, defaultValues)

//	assertions.Len(ingress.Spec.Rules, 2)
	for _, ingressRule := range ingress.Spec.Rules {
		assertions.Len(ingressRule.HTTP.Paths, 2)
		assertions.Equal("/first", ingressRule.HTTP.Paths[0].Path)
		assertions.Equal("first-service", ingressRule.HTTP.Paths[0].Backend.ServiceName)
		assertions.Equal(int32(8000), ingressRule.HTTP.Paths[0].Backend.ServicePort.IntVal)
		assertions.Equal("/second", ingressRule.HTTP.Paths[1].Path)
		assertions.Equal("second-service", ingressRule.HTTP.Paths[1].Backend.ServiceName)
		assertions.Equal(int32(8000), ingressRule.HTTP.Paths[1].Backend.ServicePort.IntVal)
	}

}

func TestIngressCustomIngressClassApi18(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	defaultValues := map[string]string{
		"ingress.enabled":      "true",
		"ingress.ingressClass": "custom-ingress-class",
	}
	_, releaseName, ingress := givenAnIngressTemplateWithHelmApi18(t, assertions, defaultValues)

	assertions.Equal(releaseName+"-chart-test", ingress.Name)
	assertions.NotNil(ingress.Annotations)
	assertions.NotEmpty(ingress.Annotations)
	assertions.Equal("custom-ingress-class", ingress.Annotations["kubernetes.io/ingress.class"])

}
