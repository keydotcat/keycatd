package cmds

import (
	"log"

	"github.com/keydotcat/server/api"
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
	viper.SetDefault("db.maxconns", 0)
	viper.SetDefault("db.type", "postgresql")
	viper.SetDefault("only_invited", false)
	viper.SetDefault("csrf.hash_key", "")
	viper.SetDefault("csrf.block_key", "")
	viper.SetDefault("session.redis.server", "")
	viper.SetDefault("session.redis.db_id", 0)
	viper.SetDefault("mail.from", "")
	viper.SetDefault("mail.smtp.server", "")
	viper.SetDefault("mail.smtp.user", "")
	viper.SetDefault("mail.smtp.password", "")
	viper.SetDefault("mail.sparkpost.key", "")
	viper.SetDefault("mail.sparkpost.eu", false)
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
	c.DBType = viper.GetString("db.type")
	if len(c.DBType) == 0 {
		c.DBType = "postgresql"
	}
	c.DBMaxConns = viper.GetInt("db.maxconns")
	c.OnlyInvited = viper.GetBool("only_invited")
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
	if len(viper.GetString("mail.sparkpost.key")) > 0 {
		c.MailSparkpost = &api.ConfMailSparkpost{
			Key: viper.GetString("mail.sparkpost.key"),
			EU:  viper.GetBool("mail.sparkpost.eu"),
		}
	}
	if srv := viper.GetString("session.redis.server"); len(srv) > 0 {
		c.SessionRedis = &api.ConfSessionRedis{srv, viper.GetInt("session.redis.db_id")}
	}
	return c
}
