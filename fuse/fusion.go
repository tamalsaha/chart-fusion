package fuse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/gobuffalo/flect"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chart"
	crdv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ylib "k8s.io/apimachinery/pkg/util/yaml"
	"kmodules.xyz/resource-metadata/hub"
	"sigs.k8s.io/yaml"
)

var (
	sampleDir   = "/home/tamal/go/src/github.com/tamalsaha/chart-fusion/fuse/samples"
	chartDir    = "/home/tamal/go/src/github.com/tamalsaha/chart-fusion/fuse/charts"
	chartName   = "mongodb-editor"
	releaseName = "default"
	chartSchema = crdv1.JSONSchemaProps{
		Type:       "object",
		Properties: map[string]crdv1.JSONSchemaProps{},
	}
	model = ValuesModel{
		X: X{
			Release: Release{
				Name:      "kubedb-community",
				Namespace: "demo",
				Service:   "Helm",
			},
			Chart: Chart{
				Name:       "kubedb",
				Version:    "v0.15.0",
				AppVersion: "v0.15.0",
			},
		},
		Objects: map[string]*unstructured.Unstructured{},
	}
	modelValues = map[string]*unstructured.Unstructured{}
	registry = hub.NewRegistryOfKnownResources()
)

type Release struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Service   string `json:"service"`
}

type Chart struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	AppVersion string `json:"appVersion"`
}

type X struct {
	Release Release `json:"release"`
	Chart   Chart   `json:"chart"`
}

type ValuesModel struct {
	X       X                         `json:"x"`

	Objects map[string]*unstructured.Unstructured `json:"objects"`
}

type ObjectModel struct {
	Key    string                     `json:"key"`
	Object *unstructured.Unstructured `json:"object"`
}

func NewCmdFuse() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "fuse-chart",
		Short:             `Fuse YAMLs`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			tplDir := filepath.Join(chartDir, chartName, "templates")
			err := os.MkdirAll(tplDir, 0755)
			if err != nil {
				return err
			}
			err = GenerateChartMetadata()
			if err != nil {
				return err
			}

			err = ProcessResource(filepath.Join(sampleDir, chartName), func(obj *unstructured.Unstructured) error {
				rsKey, err := resourceKey(obj.GetAPIVersion(), obj.GetKind(), chartName, obj.GetName())
				if err != nil {
					return err
				}
				_, _, rsFilename := resourceFilename(obj.GetAPIVersion(), obj.GetKind(), chartName, obj.GetName())

				// values
				model.Objects[rsKey] = obj
				modelValues[rsKey] = obj

				// schema
				gvr, err := registry.GVR(obj.GetObjectKind().GroupVersionKind())
				if err != nil {
					return err
				}
				descriptor, err := registry.LoadByGVR(gvr)
				if err != nil {
					return err
				}
				if descriptor.Spec.Validation != nil && descriptor.Spec.Validation.OpenAPIV3Schema != nil {
					delete(descriptor.Spec.Validation.OpenAPIV3Schema.Properties, "status")
					chartSchema.Properties[rsKey] = *descriptor.Spec.Validation.OpenAPIV3Schema
				}

				// templates
				filename := filepath.Join(tplDir, rsFilename+".yaml")
				f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					return err
				}
				defer f.Close()

				objModel := ObjectModel{
					Key:    rsKey,
					Object: obj,
				}
				modelJSON, err := json.Marshal(objModel)
				if err != nil {
					return err
				}

				var data map[string]interface{}
				err = json.Unmarshal(modelJSON, &data)
				if err != nil {
					panic(err)
				}

				localTplFile := "/home/tamal/go/src/github.com/tamalsaha/chart-fusion/templates/object.yaml"
				funcMap := sprig.TxtFuncMap()
				funcMap["toYaml"] = toYAML
				funcMap["toJson"] = toJSON
				tpl := template.Must(template.New(filepath.Base(localTplFile)).Funcs(funcMap).ParseFiles(localTplFile))
				err = tpl.Execute(f, &data)
				if err != nil {
					return err
				}

				return nil
			})
			if err != nil {
				return err
			}

			{
				data3, err := yaml.Marshal(chartSchema)
				if err != nil {
					return err
				}
				schemaFilename := filepath.Join(chartDir, chartName, "values.openapiv3_schema.yaml")
				err = ioutil.WriteFile(schemaFilename, data3, 0644)
				if err != nil {
					return err
				}
			}

			{
				modelJSON, err := json.Marshal(model)
				if err != nil {
					return err
				}

				var data map[string]interface{}
				err = json.Unmarshal(modelJSON, &data)
				if err != nil {
					panic(err)
				}

				filename := filepath.Join(chartDir, chartName, "values.yaml")
				f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					return err
				}
				defer f.Close()

				localTplFile := "/home/tamal/go/src/github.com/tamalsaha/chart-fusion/templates/values.yaml"
				funcMap := sprig.TxtFuncMap()
				funcMap["toYaml"] = toYAML
				funcMap["toJson"] = toJSON
				tpl := template.Must(template.New(filepath.Base(localTplFile)).Funcs(funcMap).ParseFiles(localTplFile))
				err = tpl.Execute(f, &data)
				if err != nil {
					return err
				}
			}
			{
				data, err := yaml.Marshal(modelValues)
				if err != nil {
					panic(err)
				}

				filename := filepath.Join(chartDir, chartName, "values.yaml")
				err = ioutil.WriteFile(filename, data, 0644)
				if err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&sampleDir, "sample-dir", sampleDir, "Sample dir")
	cmd.Flags().StringVar(&chartDir, "chart-dir", chartDir, "Charts dir")
	cmd.Flags().StringVar(&chartName, "chart-name", chartName, "Charts name")
	cmd.Flags().StringVar(&releaseName, "release-name", releaseName, "Release name")

	return cmd
}

func GenerateChartMetadata() error {
	chartMeta := chart.Metadata{
		Name:        chartName,
		Home:        "https://byte.builders",
		Version:     "v0.1.0",
		Description: "Ui Wizard Chart",
		Keywords:    []string{"appscode"},
		Maintainers: []*chart.Maintainer{
			{
				Name:  "AppsCode Engineering",
				Email: "support@appscode.com",
				URL:   "https://appscode.com",
			},
		},
	}
	data4, err := yaml.Marshal(chartMeta)
	if err != nil {
		return err
	}
	filename := filepath.Join(chartDir, chartName, "Chart.yaml")
	return ioutil.WriteFile(filename, data4, 0644)
}

type ResourceFn func(obj *unstructured.Unstructured) error

func ProcessResource(dir string, fn ResourceFn) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		fmt.Println(">>> ", path)
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		reader := ylib.NewYAMLOrJSONDecoder(bytes.NewReader(data), 2048)
		for {
			var obj unstructured.Unstructured
			err := reader.Decode(&obj)
			if err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			if obj.IsList() {
				if err := obj.EachListItem(func(item runtime.Object) error {
					return fn(item.(*unstructured.Unstructured))
				}); err != nil {
					return err
				}
			} else {
				if err := fn(&obj); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func resourceKey(apiVersion, kind, chartName, name string) (string, error) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		return "", err
	}

	groupPrefix := gv.Group
	groupPrefix = strings.TrimSuffix(groupPrefix, ".k8s.io")
	groupPrefix = strings.TrimSuffix(groupPrefix, ".kubernetes.io")
	//groupPrefix = strings.TrimSuffix(groupPrefix, ".x-k8s.io")
	groupPrefix = strings.Replace(groupPrefix, ".", "_", -1)
	groupPrefix = flect.Pascalize(groupPrefix)

	var nameSuffix string
	nameSuffix = strings.TrimPrefix(chartName, name)
	nameSuffix = strings.Replace(nameSuffix, ".", "-", -1)
	nameSuffix = strings.Trim(nameSuffix, "-")
	nameSuffix = flect.Pascalize(nameSuffix)

	return flect.Camelize(groupPrefix + kind + nameSuffix), nil
}

func resourceFilename(apiVersion, kind, chartName, name string) (string, string, string) {
	gv, err := schema.ParseGroupVersion(apiVersion)
	if err != nil {
		panic(err)
	}

	groupPrefix := gv.Group
	groupPrefix = strings.TrimSuffix(groupPrefix, ".k8s.io")
	groupPrefix = strings.TrimSuffix(groupPrefix, ".kubernetes.io")
	//groupPrefix = strings.TrimSuffix(groupPrefix, ".x-k8s.io")
	groupPrefix = strings.Replace(groupPrefix, ".", "_", -1)
	groupPrefix = flect.Pascalize(groupPrefix)

	var nameSuffix string
	nameSuffix = strings.TrimPrefix(chartName, name)
	nameSuffix = strings.Replace(nameSuffix, ".", "-", -1)
	nameSuffix = strings.Trim(nameSuffix, "-")
	nameSuffix = flect.Pascalize(nameSuffix)

	return flect.Underscore(kind), flect.Underscore(kind + nameSuffix), flect.Underscore(groupPrefix + kind + nameSuffix)
}

// toYAML takes an interface, marshals it to yaml, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toYAML(v interface{}) string {
	data, err := yaml.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return strings.TrimSuffix(string(data), "\n")
}

// toJSON takes an interface, marshals it to json, and returns a string. It will
// always return a string, even on marshal error (empty string).
//
// This is designed to be called from a template.
func toJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		// Swallow errors inside of a template.
		return ""
	}
	return string(data)
}
