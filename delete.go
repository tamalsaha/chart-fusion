package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/cli-runtime/pkg/resource"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/yaml"
	"strings"
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

			err = r.Visit(func(info *resource.Info, err error) error {
				if err != nil {
					return err
				}

				data, err := yaml.Marshal(info.Object)
				if err != nil {
					return err
				}

				fmt.Println(string(data))

				return nil
			})

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
