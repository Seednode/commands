/*
Copyright © 2025 Seednode <seednode@seedno.de>
*/

package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	ReleaseVersion string = "1.1.1"
)

var (
	databaseType     string
	databaseHost     string
	databasePort     string
	databaseUser     string
	databasePass     string
	databaseName     string
	databaseTable    string
	databaseSslMode  string
	databaseRootCert string
	databaseSslCert  string
	databaseSslKey   string
	bind             string
	port             uint16
	profile          bool
	scheme           string = "http"
	tlsCert          string
	tlsKey           string
	verbose          bool
	version          bool
)

func main() {
	cmd := &cobra.Command{
		Use:   "commands",
		Short: "Display command logs from a database.",
		Args:  cobra.ExactArgs(0),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			initializeConfig(cmd)

			if tlsCert == "" && tlsKey != "" || tlsCert != "" && tlsKey == "" {
				return errors.New("TLS certificate and keyfile must both be specified to enable HTTPS")
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return ServePage()
		},
	}

	cmd.Flags().StringVar(&databaseType, "db-type", "", "database type to connect to")
	cmd.Flags().StringVar(&databaseHost, "db-host", "", "database host to connect to")
	cmd.Flags().StringVar(&databasePort, "db-port", "", "database port to connect to")
	cmd.Flags().StringVar(&databaseUser, "db-user", "", "database user to connect as")
	cmd.Flags().StringVar(&databasePass, "db-pass", "", "database password to connect with")
	cmd.Flags().StringVar(&databaseName, "db-name", "", "database name to connect to")
	cmd.Flags().StringVar(&databaseTable, "db-table", "", "database table to query")
	cmd.Flags().StringVar(&databaseSslMode, "db-ssl-mode", "", "database ssl connection mode")
	cmd.Flags().StringVar(&databaseRootCert, "db-root-cert", "", "database ssl root certificate path")
	cmd.Flags().StringVar(&databaseSslCert, "db-ssl-cert", "", "database ssl connection certificate path")
	cmd.Flags().StringVar(&databaseSslKey, "db-ssl-key", "", "database ssl connection key path")
	cmd.Flags().StringVarP(&bind, "bind", "b", "0.0.0.0", "address to bind to")
	cmd.Flags().Uint16VarP(&port, "port", "p", 8080, "port to listen on")
	cmd.Flags().BoolVar(&profile, "profile", false, "register net/http/pprof handlers")
	cmd.Flags().StringVar(&tlsCert, "tls-cert", "", "path to TLS certificate")
	cmd.Flags().StringVar(&tlsKey, "tls-key", "", "path to TLS keyfile")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "display additional output")
	cmd.Flags().BoolVarP(&version, "version", "V", false, "display version and exit")
	cmd.Flags().SetInterspersed(true)

	cmd.CompletionOptions.HiddenDefaultCmd = true

	cmd.Flags().SetInterspersed(true)

	cmd.SetHelpCommand(&cobra.Command{
		Hidden: true,
	})

	cmd.SetVersionTemplate("commands v{{.Version}}\n")

	cmd.SilenceErrors = true

	cmd.Version = ReleaseVersion

	log.SetFlags(0)

	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func initializeConfig(cmd *cobra.Command) {
	v := viper.New()

	v.SetEnvPrefix("commands")

	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	v.AutomaticEnv()

	bindFlags(cmd, v)
}

func bindFlags(cmd *cobra.Command, v *viper.Viper) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		configName := strings.ReplaceAll(f.Name, "-", "_")

		if !f.Changed && v.IsSet(configName) {
			val := v.Get(configName)
			cmd.Flags().Set(f.Name, fmt.Sprintf("%v", val))
		}
	})
}
