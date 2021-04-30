package cmds

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/keydotcat/keycatd/api"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
)

func runServer(c api.Conf) {
	apiHandler, err := api.NewAPIHandler(c)
	if err != nil {
		log.Fatalf("Could not parse configuration: %s", err)
	}
	handler := cors.New(cors.Options{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPatch, http.MethodHead, http.MethodPut},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	}).Handler(apiHandler)
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
