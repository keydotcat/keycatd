package api

type ConfMailSMTP struct {
	Server   string `toml:"server_url"`
	User     string `toml:"user"`
	Password string `toml:"password"`
}

type ConfMailSparkpost struct {
	Key string `toml:"key"`
}

type ConfSessionRedis struct {
	Server string `toml:"server"`
	DBId   int    `toml:"db_id"`
}

type ConfCsrf struct {
	HashKey string `toml:"hash_key"`
	BlobKey string `toml:"blob_key"`
}

type Conf struct {
	URL           string             `toml:"url"`
	DB            string             `toml:"db"`
	MailSMTP      *ConfMailSMTP      `toml:"mail_smtp"`
	MailSparkpost *ConfMailSparkpost `toml:"mail_sparkpost"`
	MailFrom      string             `toml:"mail_from"`
	SessionRedis  ConfSessionRedis   `toml:"session_redis"`
	Csrf          ConfCsrf           `toml:"csrf"`
}
