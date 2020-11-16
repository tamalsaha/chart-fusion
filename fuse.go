package main

import (
	"encoding/json"
	"fmt"
	"github.com/Masterminds/sprig"
	"github.com/gobuffalo/flect"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
	"text/template"
)

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

				data, err := yaml.Marshal(info.Object)
				if err != nil {
					return err
				}

				//fmt.Println(string(data))
				//fmt.Println("-------------------------------------|" +
				//	info.Name + "|" + info.Namespace + "|" + info.Source)

				accessor, err := meta.Accessor(info.Object)
				if err != nil {
					panic(err)
					return err
				}

				model.Objects[flect.Camelize(accessor.GetName())] = info.Object

				err = ioutil.WriteFile(filepath.Join("charts", "templates", flect.Underscore(accessor.GetName())+".yaml"), data, 0644)
				if err != nil {
					panic(err)
					return err
				}

				return nil
			})

			modelJSON, err := json.Marshal(model)
			if err != nil {
				return err
			}

			// fmt.Println(string(modelJSON))

			var data map[string]interface{}
			err = json.Unmarshal(modelJSON, &data)
			if err != nil {
				panic(err)
			}

			 str, err := yaml.Marshal(data)
			if err != nil {
				panic(err)
			}
			fmt.Println(string(str))


			var tpl *template.Template
			localTplFile := "/home/tamal/go/src/github.com/tamalsaha/chart-fusion/templates/values.yaml"
			funcMap := sprig.TxtFuncMap()
			funcMap["toYaml"] = toYAML
			funcMap["toJson"] = toJSON
			tpl = template.Must(template.New(filepath.Base(localTplFile)).Funcs(funcMap).ParseFiles(localTplFile))
			err = tpl.Execute(os.Stdout, &data)
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
