package config

import (
	"fmt"
	"os"
	"strings"
	"time"


	"github.com/go-viper/mapstructure/v2"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

func Load() (*Config, error) {
	k := koanf.New(".")

	if err := k.Load(file.Provider("config/config.yaml"), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("load default config: %w", err)
	}

	appEnv := os.Getenv("APP_ENV")
	if appEnv != "" {
		envFile := fmt.Sprintf("config/config.%s.yaml", appEnv)
		if _, err := os.Stat(envFile); err == nil {
			if err := k.Load(file.Provider(envFile), yaml.Parser()); err != nil {
				return nil, fmt.Errorf("load %s config: %w", appEnv, err)
			}
		}
	}

	if err := k.Load(env.Provider("SOVEREIGN_", ".", func(s string) string {
		return strings.ReplaceAll(
			strings.ToLower(strings.TrimPrefix(s, "SOVEREIGN_")),
			"_", ".",
		)
	}), nil); err != nil {
		return nil, fmt.Errorf("load env config: %w", err)
	}

	var cfg Config
	if err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{
		DecoderConfig: &mapstructure.DecoderConfig{
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.TextUnmarshallerHookFunc(),
			),
			Result:           &cfg,
			WeaklyTypedInput: true,
			TagName:          "yaml",
		},
	}); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	// koanf + mapstructure v2 嵌套 struct 兼容性问题 workaround：
	// 对含 string 字段的子结构单独 decode
	subConfigs := []struct {
		key    string
		target interface{}
	}{
		{"cobo", &cfg.Cobo},
		{"internal", &cfg.Internal},
		{"worker", &cfg.Worker},
		{"google", &cfg.Google},
		{"kyc", &cfg.KYC},
	}
	for _, sc := range subConfigs {
		if sub := k.Cut(sc.key); len(sub.Keys()) > 0 {
			// koanf 用 "." 分隔符会把嵌套 map 的 key 展平（如 withdraw_addresses.BEP20）
			// 需要重建嵌套结构
			data := rebuildNestedMap(sub.All())

			decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
				DecodeHook: mapstructure.ComposeDecodeHookFunc(
					mapstructure.StringToTimeDurationHookFunc(),
				),
				Result:           sc.target,
				WeaklyTypedInput: true,
				TagName:          "yaml",
			})
			if err := decoder.Decode(data); err != nil {
				return nil, fmt.Errorf("unmarshal %s config: %w", sc.key, err)
			}
		}
	}

	setDefaults(&cfg)

	return &cfg, nil
}

func setDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 15 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 15 * time.Second
	}
	if cfg.Server.ShutdownTimeout == 0 {
		cfg.Server.ShutdownTimeout = 30 * time.Second
	}
	if cfg.JWT.AccessExpiry == 0 {
		cfg.JWT.AccessExpiry = 15 * time.Minute
	}
	if cfg.JWT.RefreshExpiry == 0 {
		cfg.JWT.RefreshExpiry = 7 * 24 * time.Hour
	}
	if cfg.OTP.Length == 0 {
		cfg.OTP.Length = 6
	}
	if cfg.OTP.Expiry == 0 {
		cfg.OTP.Expiry = 5 * time.Minute
	}
	if cfg.OTP.CoolDown == 0 {
		cfg.OTP.CoolDown = 60 * time.Second
	}
	if cfg.Wallet.AddressCooldown == 0 {
		cfg.Wallet.AddressCooldown = 24 * time.Hour
	}
	if cfg.Exchange.PollInterval == 0 {
		cfg.Exchange.PollInterval = 2 * time.Second
	}
	if cfg.Worker.PremiumFetchInterval == 0 {
		cfg.Worker.PremiumFetchInterval = 2 * time.Second
	}
}

// rebuildNestedMap 将 koanf 展平的 "a.b" key 重建为嵌套 map
func rebuildNestedMap(flat map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	nested := make(map[string]map[string]interface{})

	for k, v := range flat {
		parts := strings.SplitN(k, ".", 2)
		if len(parts) == 2 {
			if nested[parts[0]] == nil {
				nested[parts[0]] = make(map[string]interface{})
			}
			nested[parts[0]][parts[1]] = v
		} else {
			result[k] = v
		}
	}

	for k, v := range nested {
		result[k] = v
	}

	return result
}
