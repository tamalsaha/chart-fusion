package main

import (
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
	m.Run()
	return nil
}
