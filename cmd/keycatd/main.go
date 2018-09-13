package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/codahale/http-handlers/debug"
	"github.com/codahale/http-handlers/logging"
	"github.com/codahale/http-handlers/recovery"
	"github.com/keydotcat/server/api"
	"github.com/keydotcat/server/util"
	"github.com/rs/cors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func processConf(cfgFile string) api.Conf {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigName("keycatd")
		viper.AddConfigPath("/etc/keycatd")
		viper.AddConfigPath(".")
	}
	viper.SetDefault("port", 27623)
	viper.SetDefault("url", "http://localhost:27623")
	viper.SetDefault("db", "keycat")
	viper.SetDefault("db_type", "postgresql")
	viper.SetDefault("csrf.hash_key", "")
	viper.SetDefault("csrf.block_key", "")
	viper.SetDefault("session.redis.server", "")
	viper.SetDefault("session.redis.db_id", 0)
	viper.SetDefault("mail.from", "")
	viper.SetDefault("mail.smtp.server", "")
	viper.SetDefault("mail.smtp.user", "")
	viper.SetDefault("mail.smtp.password", "")
	viper.SetDefault("mail.sparkpost.key", "")
	viper.SetEnvPrefix("KEYCATD")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Fatal error while loading config file: %s \n", err)
	}
	c := api.Conf{}
	c.Url = viper.GetString("url")
	c.Port = viper.GetInt("port")
	c.DB = viper.GetString("db")
	c.DBType = viper.GetString("db_type")
	c.MailFrom = viper.GetString("mail.from")
	c.Csrf.HashKey = viper.GetString("csrf.hash_key")
	c.Csrf.BlockKey = viper.GetString("csrf.block_key")
	if len(viper.GetString("mail.smtp.server")) > 0 {
		c.MailSMTP = &api.ConfMailSMTP{
			Server:   viper.GetString("mail.smtp.server"),
			User:     viper.GetString("mail.smtp.user"),
			Password: viper.GetString("mail.smtp.password"),
		}
	}
	if len(viper.GetString("mail.smtp.sparkpost")) > 0 {
		c.MailSparkpost = &api.ConfMailSparkpost{
			Key: viper.GetString("mail.sparkpost.key"),
		}
	}
	if srv := viper.GetString("session.redis.server"); len(srv) > 0 {
		c.SessionRedis = &api.ConfSessionRedis{srv, viper.GetInt("session.redis.db_id")}
	}
	return c
}

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
	handler = debug.Wrap(handler)
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

func main() {
	var rootCmd = &cobra.Command{
		Use:   "keycatd",
		Short: "Keycat service to provide identities management",
		Run: func(cmd *cobra.Command, args []string) {
			flags := cmd.Flags()
			showVer, err := flags.GetBool("version")
			if err != nil {
				log.Fatalf("Could not get version flag: %s", err)
				return
			}
			if showVer {
				log.Printf("Keycat server version is %s (web %s)", util.GetServerVersion(), util.GetWebVersion())
				return
			}
			cfgFile, err := flags.GetString("config")
			if err != nil {
				log.Fatalf("Could not get config file: %s", err)
				return
			}
			c := processConf(cfgFile)
			runServer(c)
		},
	}
	rootCmd.PersistentFlags().Bool("version", false, "Show version")
	rootCmd.PersistentFlags().String("config", "", "Configuration file (default is ./keycatd.yaml)")
	rootCmd.Execute()
}
