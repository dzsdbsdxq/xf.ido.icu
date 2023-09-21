package config

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"go.uber.org/zap/zapcore"
	"os"
)

// 系统配置，对应yml
// viper内置了mapstructure, yml文件用"-"区分单词, 转为驼峰方便

// Conf 全局配置变量
var Conf = new(config)

type config struct {
	System      *SystemConfig      `mapstructure:"system" json:"system"`
	Logs        *LogsConfig        `mapstructure:"logs" json:"logs"`
	RateLimit   *RateLimitConfig   `mapstructure:"rate-limit" json:"rateLimit"`
	XunFei      *XunFeiConfig      `mapstructure:"xun-fei" json:"xun-fei"`
	XunFeiSpark *XunFeiSparkConfig `mapstructure:"xun-fei-spark" json:"xun-fei-spark"`
	Encode      *EncodeConfig      `mapstructure:"encode" json:"encode"`
}

func InitConfig() {
	workDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("读取应用目录失败:%s \n", err))
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(workDir + "/")
	// 读取配置信息
	err = viper.ReadInConfig()

	// 热更新配置
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		// 将读取的配置信息保存至全局变量Conf
		if err := viper.Unmarshal(Conf); err != nil {
			panic(fmt.Errorf("初始化配置文件失败:%s \n", err))
		}
	})

	if err != nil {
		panic(fmt.Errorf("读取配置文件失败:%s \n", err))
	}
	// 将读取的配置信息保存至全局变量Conf
	if err := viper.Unmarshal(Conf); err != nil {
		panic(fmt.Errorf("初始化配置文件失败:%s \n", err))
	}
}

type SystemConfig struct {
	Mode string `mapstructure:"mode" json:"mode"`
	Port int    `mapstructure:"port" json:"port"`
	Host string `mapstructure:"host" json:"host"`
}

type LogsConfig struct {
	Level      zapcore.Level `mapstructure:"level" json:"level"`
	Path       string        `mapstructure:"path" json:"path"`
	MaxSize    int           `mapstructure:"max-size" json:"maxSize"`
	MaxBackups int           `mapstructure:"max-backups" json:"maxBackups"`
	MaxAge     int           `mapstructure:"max-age" json:"maxAge"`
	Compress   bool          `mapstructure:"compress" json:"compress"`
}
type RateLimitConfig struct {
	FillInterval int64 `mapstructure:"fill-interval" json:"fillInterval"`
	Capacity     int64 `mapstructure:"capacity" json:"capacity"`
}
type XunFeiConfig struct {
	HostUrlV1 string `mapstructure:"host-url-v1" json:"host-url-v1"`
	HostUrlV2 string `mapstructure:"host-url-v2" json:"host-url-v2"`
	Appid     string `mapstructure:"appid" json:"appid"`
	Secret    string `mapstructure:"secret" json:"secret"`
	Key       string `mapstructure:"key" json:"key"`
}
type XunFeiSparkConfig struct {
	HostUrlSpark string `mapstructure:"host-url-spark" json:"host-url-spark"`
	Appid        string `mapstructure:"appid" json:"appid"`
	Assistants   string `mapstructure:"assistants" json:"assistants"`
	Secret       string `mapstructure:"secret" json:"secret"`
	Key          string `mapstructure:"key" json:"key"`
}
type EncodeConfig struct {
	AesKey string `mapstructure:"aes-key" json:"aes-key"`
	AesIv  string `mapstructure:"aes-iv" json:"aes-iv"`
}
