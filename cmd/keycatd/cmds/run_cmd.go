package cmds

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/codahale/http-handlers/logging"
	"github.com/codahale/http-handlers/recovery"
	"github.com/keydotcat/keycatd/api"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
)

func runServer(c api.Conf) {
	apiHandler, err := api.NewAPIHandler(c)
	if err != nil {
		log.Fatalf("Could not parse configuration: %s", err)
	}
	handler := cors.AllowAll().Handler(apiHandler)
	logHandler := logging.Wrap(handler, os.Stdout)
	logHandler.Start()
	defer logHandler.Stop()
	handler = recovery.Wrap(logHandler, recovery.LogOnPanic)
	s := &http.Server{
		Addr:           fmt.Sprintf(":%d", c.Port),
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Printf("Listening at %s", s.Addr)
	log.Fatal(s.ListenAndServe())
}

func RunCmd(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	cfgFile, err := flags.GetString("config")
	if err != nil {
		log.Fatalf("Could not get config file: %s", err)
		return
	}
	c := processConf(cfgFile)
	runServer(c)
}
