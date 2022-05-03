package secret

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

func givenASecretTemplateWithHelm(t *testing.T, require *require.Assertions, values map[string]string) (string, v1.Secret) {
	helmChartPath, err := filepath.Abs("../../")
	releaseName := "helm-basic"
	require.NoError(err)

	namespaceName := "medieval-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		SetValues:      values,
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/secret.yaml"})

	var secret v1.Secret
	helm.UnmarshalK8SYaml(t, output, &secret)
	return releaseName, secret
}

func TestSecretBasic(t *testing.T) {
	t.Parallel()
	require := require.New(t)

	values := map[string]string{
		"env.secret.SECRET_1": "aaa",
		"env.secret.SECRET_2": "bbb",
	}
	_, secret := givenASecretTemplateWithHelm(t, require, values)

	require.Equal("app-env-secret", secret.Name)
	require.Empty(secret.Annotations)
	require.Empty(secret.Labels)
	require.Equal(v1.SecretType("Opaque"), secret.Type)

	require.Len(secret.Data, 2)
	require.Equal("aaa", string(secret.Data["SECRET_1"]))
	require.Equal("bbb", string(secret.Data["SECRET_2"]))

}

func TestSecretAdvancedWithSubstitution(t *testing.T) {
	t.Parallel()
	assertions := require.New(t)

	helmChartPath, err := filepath.Abs("../../")
	releaseName := "helm-basic"
	assertions.NoError(err)

	namespaceName := "medieval-" + strings.ToLower(random.UniqueId())

	options := &helm.Options{
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
		ValuesFiles:    []string{"secret-env-vars.yaml"},
	}

	output := helm.RenderTemplate(t, options, helmChartPath, releaseName, []string{"templates/secret.yaml"})
	var secret v1.Secret
	helm.UnmarshalK8SYaml(t, output, &secret)

	t.Log(len(secret.Data))
	t.Log(len(secret.Data))

	assertions.Equal("app-env-secret", secret.Name)
	assertions.Empty(secret.Annotations)
	assertions.Empty(secret.Labels)
	assertions.Equal(v1.SecretType("Opaque"), secret.Type)

	assertions.Len(secret.Data, 4)
	assertions.Equal("9", string(secret.Data["SECRET_1"]))
	assertions.Equal("substitution_value-test", string(secret.Data["SECRET_2"]))
	assertions.Equal("aaa", string(secret.Data["SECRET_3"]))
	assertions.Equal("aaa", string(secret.Data["SECRET_4"]))

}
