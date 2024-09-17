/*
Copyright © 2024 Seednode <seednode@seedno.de>
*/

package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	ReleaseVersion string = "0.5.1"
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
	timeZone         string
	bind             string
	port             uint16
	profile          bool
	version          bool
)

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "commands",
		Short: "Display command logs from a database.",
		Args:  cobra.ExactArgs(0),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return initializeConfig(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := ServePage()
			if err != nil {
				return err
			}

			return nil
		},
	}

	rootCmd.Flags().StringVar(&databaseType, "db-type", "", "database type to connect to")
	rootCmd.Flags().StringVar(&databaseHost, "db-host", "", "database host to connect to")
	rootCmd.Flags().StringVar(&databasePort, "db-port", "", "database port to connect to")
	rootCmd.Flags().StringVar(&databaseUser, "db-user", "", "database user to connect as")
	rootCmd.Flags().StringVar(&databasePass, "db-pass", "", "database password to connect with")
	rootCmd.Flags().StringVar(&databaseName, "db-name", "", "database name to connect to")
	rootCmd.Flags().StringVar(&databaseTable, "db-table", "", "database table to query")
	rootCmd.Flags().StringVar(&databaseSslMode, "db-ssl-mode", "", "database ssl connection mode")
	rootCmd.Flags().StringVar(&databaseRootCert, "db-root-cert", "", "database ssl root certificate path")
	rootCmd.Flags().StringVar(&databaseSslCert, "db-ssl-cert", "", "database ssl connection certificate path")
	rootCmd.Flags().StringVar(&databaseSslKey, "db-ssl-key", "", "database ssl connection key path")
	rootCmd.Flags().StringVar(&timeZone, "timezone", "", "timezone to use")
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

	return rootCmd
}

func initializeConfig(cmd *cobra.Command) error {
	v := viper.New()

	v.SetConfigName("config")

	v.SetConfigType("yaml")

	v.AddConfigPath("/etc/commands/")
	v.AddConfigPath("$HOME/.config/commands")
	v.AddConfigPath(".")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return err
		}
	}

	v.SetEnvPrefix("commands")

	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	v.AutomaticEnv()

	bindFlags(cmd, v)

	return nil
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
