package main

import (
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
	"io/ioutil"
)

func NewCmdDocComment() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "doc-comment",
		Short:             `Test lib-helm`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			filename := "/home/tamal/go/src/github.com/tamalsaha/chart-fusion/charts/values.yaml"
			data, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}

			var root yaml.Node
			err = yaml.Unmarshal(data, &root)
			if err != nil {
				return err
			}

			for _, keyNode := range root.Content {
				keyNode.LineComment = "# +doc-gen:break"
			}

			data, err = yaml.Marshal(&root)
			if err != nil {
				return err
			}

			return ioutil.WriteFile(filename, data, 0644)
		},
	}

	return cmd
}
