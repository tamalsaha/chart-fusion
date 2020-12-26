package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"kubepack.dev/lib-helm/repo"
	"time"
)

func NewCmdHelm() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "helm",
		Short:             `Test lib-helm`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			timeIT(testLibHelm)
			time.Sleep(5 * time.Second)
			timeIT(testLibHelm)
			time.Sleep(5 * time.Second)
			timeIT(testLibHelm)

			return nil
		},
	}

	return cmd
}

func timeIT(f func() error) {
	start := time.Now()
	f()
	d := time.Since(start)
	fmt.Println("Duration", d)
}

func testLibHelm() error {
	reg := repo.NewDiskCacheRegistry()
	chrt, err := reg.GetChart("https://charts.appscode.com/stable/", "stash", "v0.11.8")
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(chrt.Values, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}
