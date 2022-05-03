package chart

import (
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/helm"
	"github.com/gruntwork-io/terratest/modules/k8s"
	"github.com/gruntwork-io/terratest/modules/logger"
	"github.com/gruntwork-io/terratest/modules/random"
)

func TestHelmRequireNoExtraValuesForChartTemplating(t *testing.T) {
	t.Parallel()
	namespaceName := "medieval-" + strings.ToLower(random.UniqueId())
	l := logger.Default
	l.Logf(t, "Namespace: %s\n", namespaceName)

	options := &helm.Options{
		SetValues:      map[string]string{},
		KubectlOptions: k8s.NewKubectlOptions("", "", namespaceName),
	}

	_ = helm.RenderTemplate(t, options, "../../", "test", nil)
}
