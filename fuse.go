package main

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"io/ioutil"
	api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"kmodules.xyz/resource-metadata/hub"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
	"text/template"
)

var (
	chartName   = "kubedb"
	releaseName = "kubedb-community"
	chartSchema = api.JSONSchemaProps{
		Type:       "object",
		Properties: map[string]api.JSONSchemaProps{},
	}
	registry = hub.NewRegistryOfKnownResources()
)

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

/*
{{- if contains .Chart.Name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
*/
func FullName(chartName, releaseName string) string {
	s := fmt.Sprintf("%s-%s", releaseName, chartName)
	if strings.Contains(releaseName, chartName) {
		s = releaseName
	}
	return strings.TrimSuffix(s[:min(len(s), 63)], "-")
}

var histogramGroupKind = map[schema.GroupKind]int{}
var histogramFilename = map[string]int{}

func NewCmdFuse(f cmdutil.Factory) *cobra.Command {
	var options resource.FilenameOptions
	cmd := &cobra.Command{
		Use:               "fuse",
		Short:             `Fuse YAMLs`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdNamespace, enforceNamespace, err := f.ToRawKubeConfigLoader().Namespace()
			if err != nil {
				return err
			}

			cmdNamespace = "xyz"
			enforceNamespace = false
			r := f.NewBuilder().
				Unstructured().
				Local().
				NamespaceParam(cmdNamespace).DefaultNamespace().
				FilenameParam(enforceNamespace, &options).
				ResourceTypeOrNameArgs(enforceNamespace, args...).
				SelectAllParam(false).
				LabelSelectorParam("").
				// Latest().
				Flatten().
				Do()
			err = r.Err()
			if err != nil {
				return err
			}

			model := ValuesModel{
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
				Objects: map[string]runtime.Object{},
			}

			err = r.Visit(func(info *resource.Info, err error) error {
				if err != nil {
					return err
				}

				ta, err := meta.TypeAccessor(info.Object)
				if err != nil {
					panic(err)
					return err
				}
				gv, err := schema.ParseGroupVersion(ta.GetAPIVersion())
				if err != nil {
					panic(err)
					return err
				}
				gk := schema.GroupKind{
					Group: gv.Group,
					Kind:  ta.GetKind(),
				}
				if v, ok := histogramGroupKind[gk]; ok {
					histogramGroupKind[gk] = v + 1
				} else {
					histogramGroupKind[gk] = 1
				}

				accessor, err := meta.Accessor(info.Object)
				if err != nil {
					panic(err)
					return err
				}

				n1, n2, n3 := NameChoices(ta.GetAPIVersion(), ta.GetKind(), accessor.GetName(), FullName(chartName, releaseName))
				if v, ok := histogramFilename[n1]; ok {
					histogramFilename[n1] = v + 1
				} else {
					histogramFilename[n1] = 1
				}
				if v, ok := histogramFilename[n2]; ok {
					histogramFilename[n2] = v + 1
				} else {
					histogramFilename[n2] = 1
				}
				if v, ok := histogramFilename[n3]; ok {
					histogramFilename[n3] = v + 1
				} else {
					histogramFilename[n3] = 1
				}

				return nil
			})

			data2, err := ioutil.ReadFile("values_x.yaml")
			if err != nil {
				return err
			}

			var prop api.JSONSchemaProps
			err = yaml.Unmarshal(data2, &prop)
			if err != nil {
				return err
			}
			chartSchema.Required = append(chartSchema.Required, "x")
			chartSchema.Properties["x"] = prop

			err = r.Visit(func(info *resource.Info, err error) error {
				if err != nil {
					return err
				}

				//data, err := yaml.Marshal(info.Object)
				//if err != nil {
				//	return err
				//}

				//fmt.Println(string(data))
				//fmt.Println("-------------------------------------|" +
				//	info.Name + "|" + info.Namespace + "|" + info.Source)

				ta, err := meta.TypeAccessor(info.Object)
				if err != nil {
					panic(err)
					return err
				}
				//gv, err := schema.ParseGroupVersion(ta.GetAPIVersion())
				//if err != nil {
				//	panic(err)
				//	return err
				//}
				//gk := schema.GroupKind{
				//	Group: gv.Group,
				//	Kind:  ta.GetKind(),
				//}

				accessor, err := meta.Accessor(info.Object)
				if err != nil {
					panic(err)
					return err
				}

				key, err := NG(ta.GetAPIVersion(), ta.GetKind(), accessor.GetName(), FullName(chartName, releaseName), histogramGroupKind)
				if err != nil {
					panic(err)
					return err
				}

				model.Objects[key] = info.Object

				gvr, err := registry.GVR(info.Object.GetObjectKind().GroupVersionKind())
				if err != nil {
					panic(err)
					return err
				}
				descriptor, err := registry.LoadByGVR(gvr)
				if err != nil {
					panic(err)
					return err
				}
				if descriptor.Spec.Validation != nil && descriptor.Spec.Validation.OpenAPIV3Schema != nil {
					chartSchema.Properties[key] = *descriptor.Spec.Validation.OpenAPIV3Schema
				} else {
					fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>> _______", gvr)
				}

				//err = ioutil.WriteFile(filepath.Join("charts", "templates", flect.Underscore(key)+".yaml"), data, 0644)
				//if err != nil {
				//	panic(err)
				//	return err
				//}

				objModel := ObjectModel{
					Key:    key,
					Object: info.Object,
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

				var f2 string
				n1, n2, n3 := NameChoices(ta.GetAPIVersion(), ta.GetKind(), accessor.GetName(), FullName(chartName, releaseName))
				if count, ok := histogramFilename[n1]; ok && count == 1 {
					f2 = n1
				} else if count, ok := histogramFilename[n2]; ok && count == 1 {
					f2 = n2
				} else if count, ok := histogramFilename[n3]; ok && count == 1 {
					f2 = n3
				}

				filename := filepath.Join("charts", "templates", f2+".yaml")
				f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					return err
				}
				defer f.Close()

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

			modelJSON, err := json.Marshal(model)
			if err != nil {
				return err
			}

			var data map[string]interface{}
			err = json.Unmarshal(modelJSON, &data)
			if err != nil {
				panic(err)
			}

			//str, err := yaml.Marshal(data)
			//if err != nil {
			//	panic(err)
			//}
			//fmt.Println(string(str))

			filename := "charts/values.yaml"
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

			// /home/tamal/go/src/kubedb.dev/profile-charts/charts/mongodb/values.openapiv3_schema.yaml

			data3, err := yaml.Marshal(chartSchema)
			if err != nil {
				return err
			}
			filename = filepath.Join("charts", "values.openapiv3_schema.yaml")
			err = ioutil.WriteFile(filename, data3, 0644)
			if err != nil {
				return err
			}

			return nil
		},
	}

	usage := "identifying the resource to update the annotation"
	AddFilenameOptionFlags(cmd, &options, usage)

	return cmd
}

func AddFilenameOptionFlags(cmd *cobra.Command, options *resource.FilenameOptions, usage string) {
	AddJsonFilenameFlag(cmd.Flags(), &options.Filenames, "Filename, directory, or URL to files "+usage)
	cmd.Flags().BoolVarP(&options.Recursive, "recursive", "R", options.Recursive, "Process the directory used in -f, --filename recursively. Useful when you want to manage related manifests organized within the same directory.")
}

func AddJsonFilenameFlag(flags *pflag.FlagSet, value *[]string, usage string) {
	flags.StringSliceVarP(value, "filename", "f", *value, usage)
	annotations := make([]string, 0, len(resource.FileExtensions))
	for _, ext := range resource.FileExtensions {
		annotations = append(annotations, strings.TrimLeft(ext, "."))
	}
	flags.SetAnnotation("filename", cobra.BashCompFilenameExt, annotations)
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
