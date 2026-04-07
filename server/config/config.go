package config

import "time"

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	JWT      JWTConfig      `yaml:"jwt"`
	OTP      OTPConfig      `yaml:"otp"`
	Crypto   CryptoConfig   `yaml:"crypto"`
	Wallet   WalletConfig   `yaml:"wallet"`
	Cobo     CoboConfig     `yaml:"cobo"`
	Internal InternalAPIConfig `yaml:"internal"`
	Exchange ExchangeConfig    `yaml:"exchange"`
Google   GoogleConfig   `yaml:"google"`
	Worker   WorkerConfig   `yaml:"worker"`
	Log          LogConfig          `yaml:"log"`
	Notification NotificationConfig `yaml:"notification"`
}

type GoogleConfig struct {
	ClientID string `yaml:"client_id" env:"GOOGLE_CLIENT_ID"`
}

type ServerConfig struct {
	Port            int           `yaml:"port" env:"SERVER_PORT"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Host         string `yaml:"host" env:"DB_HOST"`
	Port         int    `yaml:"port" env:"DB_PORT"`
	User         string `yaml:"user" env:"DB_USER"`
	Password     string `yaml:"password" env:"DB_PASSWORD"`
	Name         string `yaml:"name" env:"DB_NAME"`
	SSLMode      string `yaml:"ssl_mode"`
	MaxOpenConns int    `yaml:"max_open_conns"`
	MaxIdleConns int    `yaml:"max_idle_conns"`
}

func (d DatabaseConfig) DSN() string {
	return "host=" + d.Host +
		" port=" + intToStr(d.Port) +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.Name +
		" sslmode=" + d.SSLMode
}

type RedisConfig struct {
	Addr     string `yaml:"addr" env:"REDIS_ADDR"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int    `yaml:"db"`
}

type JWTConfig struct {
	AccessSecret  string        `yaml:"access_secret" env:"JWT_ACCESS_SECRET"`
	RefreshSecret string        `yaml:"refresh_secret" env:"JWT_REFRESH_SECRET"`
	AccessExpiry  time.Duration `yaml:"access_expiry"`
	RefreshExpiry time.Duration `yaml:"refresh_expiry"`
	Issuer        string        `yaml:"issuer"`
}

type OTPConfig struct {
	Length   int           `yaml:"length"`
	Expiry   time.Duration `yaml:"expiry"`
	Issuer   string        `yaml:"issuer"`
	CoolDown time.Duration `yaml:"cool_down"`
}

type CryptoConfig struct {
	AESKey string `yaml:"aes_key" env:"AES_256_KEY"`
}

type WalletConfig struct {
	AddressCooldown time.Duration `yaml:"address_cooldown"`
	Currencies      []string      `yaml:"currencies"`
}

type CoboConfig struct {
	UseMock           bool              `yaml:"use_mock"`
	BaseURL           string            `yaml:"base_url"`
	APISecret         string            `yaml:"api_secret"`
	APIPubKey         string            `yaml:"api_pub_key"`
	WalletID          string            `yaml:"wallet_id"`
	WebhookPubKey     string            `yaml:"webhook_pub_key"`
	WithdrawAddresses map[string]string `yaml:"withdraw_addresses"` // network -> address
}

type ExchangeConfig struct {
	PollInterval time.Duration       `yaml:"poll_interval"`
	Upbit        ExchangeCredentials `yaml:"upbit"`
	Bithumb      ExchangeCredentials `yaml:"bithumb"`
	Binance      ExchangeCredentials `yaml:"binance"`
	Bybit        ExchangeCredentials `yaml:"bybit"`
}

type ExchangeCredentials struct {
	APIKey    string `yaml:"api_key" env-prefix:"true"`
	APISecret string `yaml:"api_secret" env-prefix:"true"`
	BaseURL   string `yaml:"base_url"`
}

type InternalAPIConfig struct {
	APIKey     string   `yaml:"api_key"`
	AllowedIPs []string `yaml:"allowed_ips"`
}

type WorkerConfig struct {
	PremiumFetchInterval time.Duration `yaml:"premium_fetch_interval"`
	SettlementCron       string        `yaml:"settlement_cron"`
	CleanupCron          string        `yaml:"cleanup_cron"`
}

type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

type NotificationConfig struct {
	UseMock     bool   `yaml:"use_mock"`
	FromAddress string `yaml:"from_address"`
	FromName    string `yaml:"from_name"`
	AWSRegion   string `yaml:"aws_region"`
}

func intToStr(n int) string {
	if n == 0 {
		return "0"
	}
	s := ""
	for n > 0 {
		s = string(rune('0'+n%10)) + s
		n /= 10
	}
	return s
}
