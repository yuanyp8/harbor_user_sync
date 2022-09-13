package config

import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/yuanyp8/log"
	"sync"
)

const (
	PROJECT = "projects"
	MEMBERS = "members"
)

var defaultConfig *Config = &Config{
	SourceRepo:      nil,
	DestinationRepo: nil,
	locker:          sync.Mutex{},
}

type Repo struct {
	Url      string `mapstructure:"url" validate:"required"`
	Api      string `mapstructure:"api" validate:"required"`
	UserName string `mapstructure:"username" validate:"required"`
	Password string `mapstructure:"password" validate:"required"`
}

type Config struct {
	SourceRepo      *Repo `mapstructure:"source_repo" validate:"required"`
	DestinationRepo *Repo `mapstructure:"destination_repo" validate:"required"`
	locker          sync.Mutex
}

func (c *Config) LoadConf(configFile string) error {
	// load config from params path filename
	defer log.Sync()

	vip := viper.New()
	vip.SetConfigFile(configFile)

	if err := vip.ReadInConfig(); err != nil {
		log.Error("error loading config", log.String("filename", configFile))
		return err
	}

	if err := vip.Unmarshal(c); err != nil {
		log.Error("error unmarshal config", log.String("filename", configFile))
		return err
	}
	return nil
}

func C() *Config {
	return defaultConfig
}

func (r *Repo) Addr() string {
	return fmt.Sprintf("%s/%s", r.Url, r.Api)
}

func init() {
	// log.ResetDefault(log.New(os.Stdout, log.DebugLevel, log.WithCaller(false)))
	tops := []log.TeeOption{
		{Filename: "log/success.log",
			Ropt: log.RotateOptions{
				MaxSize:    2,
				MaxAge:     2,
				MaxBackups: 2,
				Compress:   false,
			},
			Lef: func(lvl log.Level) bool {
				return lvl <= log.InfoLevel
			},
		},
		{Filename: "log/error.log",
			Ropt: log.RotateOptions{
				MaxSize:    2,
				MaxAge:     2,
				MaxBackups: 2,
				Compress:   false,
			},
			Lef: func(lvl log.Level) bool {
				return lvl > log.InfoLevel
			},
		},
	}

	logger := log.NewLoggerTeeWithRotate(tops)
	log.ResetDefault(logger)
}
