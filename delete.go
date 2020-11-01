package main

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func NewCmdFuse(clientGetter genericclioptions.RESTClientGetter) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "fuse",
		Short:             `Fuse YAMLs`,
		DisableAutoGenTag: true,
		Run: func(cmd *cobra.Command, args []string) {
			
		},
	}
	return cmd
}
