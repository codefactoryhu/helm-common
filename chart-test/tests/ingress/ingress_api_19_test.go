package ingress

import (
	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/networking/v1"
	"path/filepath"
	"strings"
	"testing"
)
func givenAnIngressTemplateWithHelmApi19(t *testing.T, require *require.Assertions, values map[string]string) (string, string, v1.Ingress) {
	helmChartPath, err := filepath.Abs("../../")
	releaseName := "helm-basic"
	require.NoError(err)

	namespaceName := "medieval-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues:      values,
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/ingress.yaml"}, "--kube-version=v1.19.0")

	var ingress v1.Ingress
	helm.UnmarshalK8SYaml(t, output, &ingress)
	return namespaceName, releaseName, ingress
}

func TestIngressBasicApi19(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	defaultValues := map[string]string{
		"ingress.enabled": "true",
	}
	namespaceName, releaseName, ingress := givenAnIngressTemplateWithHelmApi19(t, assertions, defaultValues)

	assertions.Equal(releaseName+"-chart-test", ingress.Name)
//	assertions.NotNil(ingress.Annotations)
//	assertions.NotEmpty(ingress.Annotations)
//	assertions.NotEmpty(ingress.Annotations["nginx.ingress.kubernetes.io/configuration-snippet"])

	assertions.Len(ingress.Spec.TLS, 1)
//	assertions.Len(ingress.Spec.TLS[0].Hosts, 2)

//	assertions.Len(ingress.Spec.Rules, 2)
	for _, ingressRule := range ingress.Spec.Rules {
		assertions.Len(ingressRule.HTTP.Paths, 1)
		assertions.Equal("/", ingressRule.HTTP.Paths[0].Path)
		assertions.Equal(namespaceName+"-service-name", ingressRule.HTTP.Paths[0].Backend.Service.Name)
		assertions.Equal(int32(8000), ingressRule.HTTP.Paths[0].Backend.Service.Port.Number)
	}
}


func TestIngressMoreServicePathApi19(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	defaultValues := map[string]string{
		"ingress.enabled":                      "true",
		"ingress.paths[0].path":                "/first",
		"ingress.paths[0].backend.serviceName": "first-service",
		"ingress.paths[1].path":                "/second",
		"ingress.paths[1].backend.serviceName": "second-service",
	}
	_, _, ingress := givenAnIngressTemplateWithHelmApi19(t, assertions, defaultValues)

//	assertions.Len(ingress.Spec.Rules, 2)
	for _, ingressRule := range ingress.Spec.Rules {
		assertions.Len(ingressRule.HTTP.Paths, 2)
		assertions.Equal("/first", ingressRule.HTTP.Paths[0].Path)
		assertions.Equal("first-service", ingressRule.HTTP.Paths[0].Backend.Service.Name)
		assertions.Equal(int32(8000), ingressRule.HTTP.Paths[0].Backend.Service.Port.Number)
		assertions.Equal("/second", ingressRule.HTTP.Paths[1].Path)
		assertions.Equal("second-service", ingressRule.HTTP.Paths[1].Backend.Service.Name)
		assertions.Equal(int32(8000), ingressRule.HTTP.Paths[1].Backend.Service.Port.Number)
	}

}

func TestIngressCustomIngressClassApi19(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	defaultValues := map[string]string{
		"ingress.enabled":      "true",
		"ingress.ingressClass": "custom-ingress-class",
	}
	_, releaseName, ingress := givenAnIngressTemplateWithHelmApi19(t, assertions, defaultValues)

	assertions.Equal(releaseName+"-chart-test", ingress.Name)
	assertions.Equal("custom-ingress-class", *ingress.Spec.IngressClassName)

}
