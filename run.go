package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"gopkg.in/macaron.v1"
)

func NewCmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "run",
		Short:             `Run Wizard backend demo server`,
		DisableAutoGenTag: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunServer()
		},
	}

	return cmd
}

func RunServer() error {
	m := macaron.New()
	m.Use(macaron.Logger())
	m.Use(macaron.Recovery())
	m.Use(macaron.Renderer())

	m.Group("/api", func() {

	})
	m.Get("/", func() string {
		return "Hello world!"
	})
	m.Get("/hello/abc", func(ctx *macaron.Context) string {
		return "abc"
	})
	m.Get("/hello/*", func(ctx *macaron.Context) string {
		return "Hello " + ctx.Params("*") + "|" + ctx.Params("*")
	})
	m.Get("/user/*.*", func(ctx *macaron.Context) string {
		return fmt.Sprintf("Last part is: %s, Ext: %s", ctx.Params(":path"), ctx.Params(":ext"))
	})
	m.Run()
	return nil
}
