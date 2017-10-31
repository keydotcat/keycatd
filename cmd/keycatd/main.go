package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/codahale/http-handlers/debug"
	"github.com/codahale/http-handlers/logging"
	"github.com/codahale/http-handlers/recovery"
	"github.com/keydotcat/backend/api"
	"github.com/rs/cors"
	"github.com/spf13/viper"
)

func enableStackDump() {
	dump := make(chan os.Signal)
	go func() {
		stack := make([]byte, 16*1024)
		for _ = range dump {
			n := runtime.Stack(stack, true)
			fmt.Fprintf(os.Stderr, "==== %s\n%s\n====\n", time.Now(), stack[0:n])
		}
	}()
	signal.Notify(dump, syscall.SIGUSR1)
}

func processConf() api.Conf {
	viper.SetConfigName("keycatd")
	viper.AddConfigPath(".")
	viper.SetDefault("port", 27623)
	viper.SetDefault("url", "http://localhost:27623")
	viper.SetDefault("db", "")
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
	c.SessionRedis.Server = viper.GetString("session.redis.server")
	c.SessionRedis.DBId = viper.GetInt("session.redis.db_id")
	return c
}

func main() {
	c := processConf()
	apiHandler, err := api.NewAPIHandler(c)
	if err != nil {
		log.Fatalf("Could not parse configuration: %s", err)
	}
	handler := cors.Default().Handler(apiHandler)
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
	enableStackDump()
	log.Fatal(s.ListenAndServe())
}
