package config

import (
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	Database    Database `mapstructure:"database"`
	OldDatabase Database `mapstructure:"old_database"`
	Server      Server   `mapstructure:"server"`
	Security    Security `mapstructure:"security"`
	OAuth       OAuth    `mapstructure:"oauth"`
	Email       Email    `mapstructure:"email"`
}

type Database struct {
	Type     string `mapstructure:"type"` // mysql or sqlite
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	Path     string `mapstructure:"path"` // SQLite database file path
}

type Server struct {
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	URL          string `mapstructure:"url"`
	TLSCertFile  string `mapstructure:"tls_cert_file"`
	TLSKeyFile   string `mapstructure:"tls_key_file"`
	ForceHTTPS   bool   `mapstructure:"force_https"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
	IdleTimeout  int    `mapstructure:"idle_timeout"`
}

type Security struct {
	SecretKey         string `mapstructure:"secret_key"`
	CSRFKey           string `mapstructure:"csrf_key"`
	SessionName       string `mapstructure:"session_name"`
	HCaptchaSiteKey   string `mapstructure:"hcaptcha_site_key"`
	HCaptchaSecretKey string `mapstructure:"hcaptcha_secret_key"`
	MagicLinkHMACKey  string `mapstructure:"magic_link_hmac_key"`
}

type OAuth struct {
	Google GoogleOAuth `mapstructure:"google"`
}

type GoogleOAuth struct {
	ClientID     string `mapstructure:"client_id"`
	ClientSecret string `mapstructure:"client_secret"`
	RedirectURL  string `mapstructure:"redirect_url"`
}

type Email struct {
	SendGridAPIKey string `mapstructure:"sendgrid_api_key"`
	FromEmail      string `mapstructure:"from_email"`
	FromName       string `mapstructure:"from_name"`
}

func Load() *Config {
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.url", "http://localhost:8080")
	viper.SetDefault("server.force_https", false)
	viper.SetDefault("server.read_timeout", 15)
	viper.SetDefault("server.write_timeout", 15)
	viper.SetDefault("server.idle_timeout", 60)
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.path", "./uspavalia.db")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.name", "uspavalia")
	viper.SetDefault("old_database.host", "localhost")
	viper.SetDefault("old_database.port", 3306)
	viper.SetDefault("old_database.name", "uspavalia_old")
	viper.SetDefault("security.session_name", "uspavalia_session")
	viper.SetDefault("email.from_name", "USP Avalia")

	viper.SetEnvPrefix("USPAVALIA")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		logrus.Fatalf("Unable to decode config: %v", err)
	}

	return &config
}
