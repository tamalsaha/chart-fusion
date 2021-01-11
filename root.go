package main

import (
	"flag"
	"github.com/tamalsaha/chart-fusion/fuse"

	"github.com/spf13/cobra"
	v "gomodules.xyz/x/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	cliflag "k8s.io/component-base/cli/flag"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"kmodules.xyz/client-go/logs"
)

func NewRootCmd() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:               "chart-fusion",
		Short:             ``,
		Long:              ``,
		DisableAutoGenTag: true,
	}

	flags := rootCmd.PersistentFlags()
	// Normalize all flags that are coming from other packages or pre-configurations
	// a.k.a. change all "_" to "-". e.g. glog package
	flags.SetNormalizeFunc(cliflag.WordSepNormalizeFunc)

	kubeConfigFlags := genericclioptions.NewConfigFlags(true)
	kubeConfigFlags.AddFlags(flags)
	matchVersionKubeConfigFlags := cmdutil.NewMatchVersionFlags(kubeConfigFlags)
	matchVersionKubeConfigFlags.AddFlags(flags)

	flags.AddGoFlagSet(flag.CommandLine)
	logs.ParseFlags()

	f := cmdutil.NewFactory(matchVersionKubeConfigFlags)

	rootCmd.AddCommand(v.NewCmdVersion())
	rootCmd.AddCommand(NewCmdRun())
	rootCmd.AddCommand(NewCmdHelm())
	rootCmd.AddCommand(NewCmdFuse(f))
	rootCmd.AddCommand(fuse.NewCmdFuse())
	rootCmd.AddCommand(NewCmdMerge())
	rootCmd.AddCommand(NewCmdAppGen())
	rootCmd.AddCommand(NewCmdDocComment())

	return rootCmd
}
