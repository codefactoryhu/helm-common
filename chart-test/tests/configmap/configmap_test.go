package configmap

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

func givenAConfigMapTemplateWithHelm(t *testing.T, require *require.Assertions, values map[string]string) (string, v1.ConfigMap) {
	helmChartPath, err := filepath.Abs("../../")
	releaseName := "helm-basic"
	require.NoError(err)

	namespaceName := "medieval-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues:      values,
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/configmap.yaml"})

	var configMap v1.ConfigMap
	helm.UnmarshalK8SYaml(t, output, &configMap)
	return releaseName, configMap
}

func TestConfigMapBasic(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	values := map[string]string{
		"env.configMap.KEY_1": "aaa",
		"env.configMap.KEY_2": "bbb",
	}
	//releaseName, configMap := givenASecretTemplateWithHelm(t, require, defaultValues)
	_, configMap := givenAConfigMapTemplateWithHelm(t, require, values)

	require.Equal("app-env-config-map", configMap.Name)
	require.Empty(configMap.Annotations)
	require.Empty(configMap.Labels)
	//require.Equal(v1.SecretType("Opaque"), configMap.Type)

	require.Len(configMap.Data, 2)

	require.Equal("aaa", configMap.Data["KEY_1"])
	require.Equal("bbb", configMap.Data["KEY_2"])

}
