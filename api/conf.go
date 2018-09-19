package api

import (
	"fmt"

	"github.com/keydotcat/server/util"
)

type ConfMailSMTP struct {
	Server   string `toml:"server_url"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

type ConfMailSparkpost struct {
	Key string `toml:"key"`
	EU  bool
}

type ConfSessionRedis struct {
	Server string `toml:"server"`
	DBId   int    `toml:"db_id"`
}

type ConfCsrf struct {
	HashKey  string `toml:"hash_key"`
	BlockKey string `toml:"block_key"`
}

type Conf struct {
	Url           string `toml:"url"`
	Port          int    `toml:"port"`
	DB            string `toml:"db"`
	DBMaxConns    int
	DBType        string             `toml:"db_type"`
	MailSMTP      *ConfMailSMTP      `toml:"mail_smtp"`
	MailSparkpost *ConfMailSparkpost `toml:"mail_sparkpost"`
	MailFrom      string             `toml:"mail_from"`
	SessionRedis  *ConfSessionRedis  `toml:"session_redis"`
	Csrf          ConfCsrf           `toml:"csrf"`
}

func (c Conf) validate() error {
	if c.Port < 1 {
		return util.NewErrorf("Invalid port defined in the configuration")
	}
	if len(c.Url) == 0 {
		c.Url = fmt.Sprintf("http://localhost:%d", c.Port)
	}
	if len(c.DB) == 0 {
		return util.NewErrorf("Invalid db configuration")
	}
	if c.DBType != "postgresql" && c.DBType != "cockroackdb" {
		return util.NewErrorf("Invalid db type (%s)", c.DBType)
	}
	if len(c.MailFrom) == 0 {
		return util.NewErrorf("Invalid mail.from")
	}
	if len(c.Csrf.HashKey) != 32 && len(c.Csrf.HashKey) != 64 {
		return util.NewErrorf("Invalid csrf.hash_key. It has to be 32 or 64 characters long")
	}
	bl := len(c.Csrf.BlockKey)
	if bl != 0 && bl != 16 && bl != 24 && bl != 32 {
		return util.NewErrorf("Invalid csrf.block_key. It has to be 16, 24 or 32 characters long, or 0 to disable encryption")
	}
	if !TEST_MODE {
		smtp := c.MailSMTP != nil
		spark := c.MailSparkpost != nil
		if (!smtp && !spark) || (smtp && spark) {
			return util.NewErrorf("Either configure mail.smtp (%t) or mail.sparkpost (%t)", smtp, spark)
		}
		if smtp && len(c.MailSMTP.Server) == 0 {
			return util.NewErrorf("Invalid mail.smtp.server")
		}
		if spark && len(c.MailSparkpost.Key) == 0 {
			return util.NewErrorf("Invalid mail.sparkpost.key")
		}
	}
	if c.SessionRedis != nil && len(c.SessionRedis.Server) == 0 {
		return util.NewErrorf("Invalid session.redis.server")
	}
	return nil
}
