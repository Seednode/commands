/*
Copyright © 2023 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"log"

	"github.com/spf13/cobra"
)

const (
	Version string = "0.1.0"
)

var (
	bind    string
	port    uint16
	verbose bool
	version bool

	rootCmd = &cobra.Command{
		Use:   "commands",
		Short: "Display command log from a database.",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			err := ServePage()
			if err != nil {
				return err
			}

			return nil
		},
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&bind, "bind", "b", "0.0.0.0", "address to bind to")
	rootCmd.Flags().Uint16VarP(&port, "port", "p", 8080, "port to listen on")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "log http errors to stdout")
	rootCmd.Flags().BoolVarP(&version, "version", "V", false, "display version and exit")
	rootCmd.Flags().SetInterspersed(true)

	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.SilenceErrors = true
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	rootCmd.SetVersionTemplate("commands v{{.Version}}\n")
	rootCmd.Version = Version
}
