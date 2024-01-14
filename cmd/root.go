/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"github.com/spf13/cobra"
)

const (
	ReleaseVersion string = "0.3.4"
)

var (
	DatabaseType     string
	DatabaseHost     string
	DatabasePort     string
	DatabaseUser     string
	DatabasePass     string
	DatabaseName     string
	DatabaseTable    string
	DatabaseSslMode  string
	DatabaseRootCert string
	DatabaseSslCert  string
	DatabaseSslKey   string
	TimeZone         string
	bind             string
	port             uint16
	profile          bool
	version          bool

	rootCmd = &cobra.Command{
		Use:   "commands",
		Short: "Display command logs from a database.",
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

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		return err
	}

	return nil
}

func init() {
	rootCmd.Flags().StringVar(&DatabaseType, "database-type", "", "database type to connect to")
	rootCmd.Flags().StringVar(&DatabaseHost, "database-host", "", "database host to connect to")
	rootCmd.Flags().StringVar(&DatabasePort, "database-port", "", "database port to connect to")
	rootCmd.Flags().StringVar(&DatabaseUser, "database-user", "", "database user to connect as")
	rootCmd.Flags().StringVar(&DatabasePass, "database-pass", "", "database password to connect with")
	rootCmd.Flags().StringVar(&DatabaseName, "database-name", "", "database name to connect to")
	rootCmd.Flags().StringVar(&DatabaseTable, "database-table", "", "database table to query")
	rootCmd.Flags().StringVar(&DatabaseSslMode, "database-ssl-mode", "", "database ssl connection mode")
	rootCmd.Flags().StringVar(&DatabaseRootCert, "database-root-cert", "", "database ssl root certificate path")
	rootCmd.Flags().StringVar(&DatabaseSslCert, "database-ssl-cert", "", "database ssl connection certificate path")
	rootCmd.Flags().StringVar(&DatabaseSslKey, "database-ssl-key", "", "database ssl connection key path")
	rootCmd.Flags().StringVar(&TimeZone, "time-zone", "", "timezone to use")
	rootCmd.Flags().StringVarP(&bind, "bind", "b", "0.0.0.0", "address to bind to")
	rootCmd.Flags().Uint16VarP(&port, "port", "p", 8080, "port to listen on")
	rootCmd.Flags().BoolVar(&profile, "profile", false, "register net/http/pprof handlers")
	rootCmd.Flags().BoolVarP(&version, "version", "V", false, "display version and exit")
	rootCmd.Flags().SetInterspersed(true)

	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	rootCmd.SilenceErrors = true
	rootCmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	rootCmd.SetVersionTemplate("commands v{{.Version}}\n")
	rootCmd.Version = ReleaseVersion
}
