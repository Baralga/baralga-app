package shared

import (
	"log"
	"strings"
	"time"
)

type Config struct {
	BindPort   string `default:"8080"`
	Webroot    string `default:"http://localhost:8080"`
	Db         string `default:"postgres://postgres:postgres@localhost:5432/baralga"`
	DbMaxConns int32  `default:"3"`
	Env        string `default:"dev"`

	JWTSecret  string `default:"secret"`
	JWTExpiry  string `default:"24h"`
	CSRFSecret string `default:"CSRFsecret"`

	SMTPServername string `default:"localhost:1025"`
	SMTPFrom       string `default:"smtp.from@baralga.com"`
	SMTPUser       string `default:"smtp.user@baralga.com"`
	SMTPPassword   string `default:"SMTPPassword"`

	DataProtectionURL string `default:"#"`

	GithubClientId     string `default:"asdfasdf"`
	GithubClientSecret string `default:""`
	GithubRedirectURL  string `default:"http://localhost:8080/github/callback"`

	GoogleClientId     string `default:"asfdasf"`
	GoogleClientSecret string `default:""`
	GoogleRedirectURL  string `default:"http://localhost:8080/google/callback"`
}

func (c *Config) ExpiryDuration() time.Duration {
	expiryDuration, err := time.ParseDuration(c.JWTExpiry)
	if err != nil {
		log.Printf("could not parse jwt expiry %s", c.JWTExpiry)
		expiryDuration = time.Duration(24 * time.Hour)
	}
	return expiryDuration
}

func (c *Config) IsProduction() bool {
	return strings.ToLower(c.Env) == "production"
}
