package main

import (
	"github.com/spf13/cobra"
	"io/ioutil"
	api "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

func NewCmdMerge() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "merge",
		Short:             `Merge YAMLs`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			data, err := ioutil.ReadFile("values_x.yaml")
			if err != nil {
				return err
			}

			var prop api.JSONSchemaProps
			err = yaml.Unmarshal(data, &prop)
			if err != nil {
				panic(err)
			}

			return nil
		},
	}

	return cmd
}
