package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chart"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"kubepack.dev/kubepack/apis"
	"kubepack.dev/kubepack/apis/kubepack/v1alpha1"
	"kubepack.dev/kubepack/pkg/lib"
	"kubepack.dev/lib-helm/repo"
	"sigs.k8s.io/application/api/app/v1beta1"
	"sigs.k8s.io/yaml"
	"sort"
)

func NewCmdAppGen() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "appgen",
		Short:             `Generate App`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return genApp()
		},
	}

	return cmd
}

func genApp() error {
	reg := repo.NewDiskCacheRegistry()

	sel := v1alpha1.ChartSelection{
		ChartRef:    v1alpha1.ChartRef{
			URL:  "https://charts.appscode.com/stable/",
			Name: "stash",
		},
		Version:     "v0.11.8",
		ReleaseName: "stash",
		Namespace:   "default",
	}
	chrt, err := reg.GetChart(sel.URL, sel.Name, sel.Version)
	if err != nil {
		return err
	}

	app := Result(sel, chrt.Chart)
	data, err := yaml.Marshal(app)
	if err != nil {
		return err
	}
	fmt.Println(string(data))

	return nil
}

func Result(sel v1alpha1.ChartSelection, chrt *chart.Chart) *v1beta1.Application {
	desc := lib.GetPackageDescriptor(chrt)

	p := v1alpha1.ApplicationPackage{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1alpha1.SchemeGroupVersion.String(),
			Kind:       "ApplicationPackage",
		},
		Bundle: sel.Bundle,
		Chart: v1alpha1.ChartRepoRef{
			Name:    sel.Name,
			URL:     sel.URL,
			Version: sel.Version,
		},
		Channel: v1alpha1.RegularChannel,
	}
	data, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	b := &v1beta1.Application{
		TypeMeta: metav1.TypeMeta{
			APIVersion: v1beta1.SchemeGroupVersion.String(),
			Kind:       "Application",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      sel.ReleaseName,
			Namespace: sel.Namespace,
			Labels:    nil, // TODO: ?
			Annotations: map[string]string{
				apis.LabelPackage: string(data),
			},
		},
		Spec: v1beta1.ApplicationSpec{
			Descriptor: v1beta1.Descriptor{
				Type:        chrt.Name(),
				Description: desc.Description,
				Icons:       v1alpha1.ConvertImageSpec(desc.Icons),
				Maintainers: v1alpha1.ConvertContactData(desc.Maintainers),
				Keywords:    desc.Keywords,
				Links:       v1alpha1.ConvertLink(desc.Links),
				Notes:       "",
				Version:     chrt.Metadata.AppVersion,
				Owners:      nil, // TODO: Add the user email who is installing this app
			},
			AddOwnerRef:   false,
			Info:          nil,
			AssemblyPhase: v1beta1.Ready,
		},
	}

	gks := []metav1.GroupKind{
		{
			Group: "kubedb.com",
			Kind:  "MongoDB",
		},
		{
			Group: "",
			Kind:  "Service",
		},
		{
			Group: "",
			Kind:  "Secret",
		},
		{
			Group: "apps",
			Kind:  "StatefulSet",
		},
		{
			Group: "policy",
			Kind:  "PodDisruptionBudget",
		},
		{
			Group: "policy",
			Kind:  "PodSecurityPolicy",
		},
		{
			Group: "rbac.authorization.k8s.io",
			Kind:  "Role",
		},
		{
			Group: "rbac.authorization.k8s.io",
			Kind:  "RoleBinding",
		},
		{
			Group: "monitoring.coreos.com",
			Kind:  "ServiceMonitor",
		},
		{
			Group: "stash.appscode.com",
			Kind:  "BackupConfiguration",
		},
		{
			Group: "stash.appscode.com",
			Kind:  "Repository",
		},
		{
			Group: "grafana.searchlight.dev",
			Kind:  "Dashboard",
		},
	}
	sort.Slice(gks, func(i, j int) bool {
		if gks[i].Group == gks[j].Group {
			return gks[i].Kind < gks[j].Kind
		}
		return gks[i].Group < gks[j].Group
	})
	b.Spec.ComponentGroupKinds = gks

	b.Spec.Selector = &metav1.LabelSelector{
		MatchLabels: map[string]string{
			"kubedb.com/kind": "Elasticsearch",
			"kubedb.com/name": "name",
		},
	}
	return b
}
