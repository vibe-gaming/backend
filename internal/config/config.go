package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `env:"ENV" env-required:"true"`
	LogLevel   string `env:"LOG_LEVEL" env-default:"info" env-description:"logging level, debug, info, etc."`
	HttpServer HttpServer
	Database   Database
	Limiter    Limiter
	Auth       AuthConfig
	SMTP       SMTPConfig
	Email      EmailConfig
	Cache      Cache
}

type HttpServer struct {
	Port           string        `env:"HTTP_PORT" env-default:"8080"`
	Timeout        time.Duration `env:"HTTP_TIMEOUT" env-default:"4s"`
	IdleTimeout    time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
	SwaggerEnabled bool          `env:"HTTP_SWAGGER_ENABLED" env-default:"false"`
}

type Database struct {
	Net                string        `env:"DB_NET" env-default:"tcp"`
	Server             string        `env:"DB_SERVER" env-required:"true"`
	DBName             string        `env:"DB_NAME" env-required:"true"`
	User               string        `env:"DB_USER" env-required:"true"`
	Password           string        `env:"DB_PASSWORD" env-required:"true"`
	TimeZone           string        `env:"DB_TIMEZONE"`
	Timeout            time.Duration `env:"DB_TIMEOUT" env-default:"2s"`
	MaxIdleConnections int           `env:"DB_MAX_IDLE_CONNECTIONS" env-default:"40"`
	MaxOpenConnections int           `env:"DB_MAX_OPEN_CONNECTIONS" env-default:"40"`
}

type Limiter struct {
	RPS   int           `env:"LIMITER_RPS" env-default:"10"`
	Burst int           `env:"LIMITER_BURST" env-default:"20"`
	TTL   time.Duration `env:"LIMITER_TTL" env-default:"10m"`
}

type AuthConfig struct {
	JWT                    JWTConfig
	PasswordSalt           string `env:"AUTH_PASSWORD_SALT" env-required:"true"`
	VerificationCodeLength int    `env:"AUTH_VERIFICATION_CODE_LENGTH" env-default:"6"`
}

type JWTConfig struct {
	AccessTokenTTL  time.Duration `env:"JWT_ACCESS_TOKEN_TTL" env-default:"1m"`
	RefreshTokenTTL time.Duration `env:"JWT_REFRESH_TOKEN_TTL" env-default:"240h"`
	SigningKey      string        `env:"JWT_SIGNING_KEY" env-required:"true"`
}

type SMTPConfig struct {
	Host string `env:"SMTP_HOST" env-required:"true"`
	Port int    `env:"SMTP_PORT" env-required:"true"`
	From string `env:"SMTP_FROM" env-required:"true"`
	Pass string `env:"SMTP_PASS" env-required:"true"`
}

type EmailConfig struct {
	Enabled   bool `env:"EMAIL_ENABLED" env-default:"false"`
	Templates EmailTemplates
}

type EmailTemplates struct {
	Verification string `env:"EMAIL_TEMPLATE_VERIFICATION" env-required:"true"`
}

type Cache struct {
	Type  string `env:"REDIS_TYPE" env-required:"true" env-description:"specifies provider, one of redis/redisCluster"`
	Redis struct {
		Address  string `env:"REDIS_ADDR" env-default:"" env-description:"redis host:port single instance"`
		Password string `env:"REDIS_PASSWORD" env-default:"" env-description:"redis password if exists"`
		PoolSize int    `env:"REDIS_POOL_SIZE" env-default:"70" env-description:"max tcp connections pool size"`
	}
	RedisCluster struct {
		Addresses []string `env:"REDIS_CLUSTER_ADDRS" env-default:"" env-description:"redis cluster nodes: ['172.27.29.90:7000','172.27.29.91:7001'', '172.27.29.92:7002'']"`
		Password  string   `env:"REDIS_PASSWORD" env-default:"" env-description:"redis password if exists"`
		PoolSize  int      `env:"REDIS_POOL_SIZE" env-default:"70" env-description:"max tcp connections pool size"`
	}
}

func MustLoad() *Config {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read config from environment: %s", err)
	}

	return &cfg
}
