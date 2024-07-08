package provider

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/samber/do"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type AppConfig struct {
	Port            int    `env:"PORT" env-default:"32919"`
	DB              string `env:"DB"`
	CookieSecret    string `env:"COOKIE_SECRET" env-default:"UbnpjqcvDJ8mDCB"`
	StaticCdnPrefix string `env:"STATIC_CDN_PREFIX" env-default:"/static"`
	AvatarCdn       string `env:"AVATAR_CDN" env-default:"https://gravatar.cooluc.com/avatar/"`
}

func NewRepository(i *do.Injector) (*gorm.DB, error) {
	appConfig := do.MustInvoke[*AppConfig](i)
	db, err := gorm.Open(postgres.Open(appConfig.DB), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Info),
		TranslateError: true})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func NewAppConfig(i *do.Injector) (*AppConfig, error) {
	var cfg AppConfig

	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
