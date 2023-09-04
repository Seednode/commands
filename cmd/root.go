/*
Copyright © 2023 Seednode <seednode@seedno.de>
*/

package web

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/spf13/cobra"

	db "seedno.de/seednode/commands-web/db"
	utils "seedno.de/seednode/commands-web/utils"
	"seedno.de/seednode/commands-web/web"
)

const (
	Version string = "0.2.0"
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

func Execute() error {
	err := rootCmd.Execute()
	if err != nil {
		return err
	}

	return nil
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

func ServeVersion() httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		data := []byte(fmt.Sprintf("commands v%s\n", Version))

		w.Header().Write(bytes.NewBufferString("Content-Length: " + strconv.Itoa(len(data))))

		w.Write(data)
	}
}

func ServePage() error {
	timezone, err := utils.GetEnvVar("TZ", false)
	if err != nil {
		timezone = "UTC"
	}

	time.Local, err = time.LoadLocation(timezone)
	if err != nil {
		return err
	}

	bindHost, err := net.LookupHost(bind)
	if err != nil {
		return err
	}

	bindAddr := net.ParseIP(bindHost[0])
	if bindAddr == nil {
		return errors.New("invalid bind address provided")
	}

	dbType, err := utils.GetEnvVar("COMMANDS_DB_TYPE", false)
	if err != nil {
		return err
	}
	if dbType != "cockroachdb" && dbType != "postgresql" {
		return errors.New("invalid database type specified")
	}

	databaseURL, err := db.GetDatabaseURL(dbType)
	if err != nil {
		return err
	}

	tableName, err := utils.GetEnvVar("COMMANDS_DB_TABLE", false)
	if err != nil {
		return err
	}

	database := &db.Database{
		Url:   databaseURL,
		Table: tableName,
	}

	mux := httprouter.New()

	mux.PanicHandler = web.ServerErrorHandler()

	mux.GET("/", web.ServePageHandler(database))

	mux.GET("/version", ServeVersion())

	srv := &http.Server{
		Addr:         net.JoinHostPort(bind, strconv.Itoa(int(port))),
		Handler:      mux,
		IdleTimeout:  10 * time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Minute,
	}

	err = srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
