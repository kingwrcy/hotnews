package provider

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/samber/do"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type AppConfig struct {
	Port         int    `yaml:"port" env:"PORT" env-default:"7832"`
	DB           string `yaml:"db" env:"DB" env-default:"/app/data.sqlite"`
	CookieSecret string `yaml:"cookieSecret" env:""`
	RedisAddress string `yaml:"redisAddress" env:"localhost:6379"`
}

func NewRepository(i *do.Injector) (*gorm.DB, error) {
	appConfig := do.MustInvoke[*AppConfig](i)
	db, err := gorm.Open(mysql.Open(appConfig.DB), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Info),
		TranslateError: true})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func NewAppConfig(i *do.Injector) (*AppConfig, error) {
	var cfg AppConfig
	err := cleanenv.ReadConfig("config.yml", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
