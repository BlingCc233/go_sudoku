package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	// 配置文件路径
	configPath string
)

type Config struct {
	ListenAddr string   `mapstructure:"listen"`
	RemoteAddr string   `mapstructure:"remote"`
	ObfDomain  []string `mapstructure:"obf_domain"`
}

func init() {
	home, _ := homedir.Dir()
	// 默认的配置文件名称
	configFilename := ".lightsocks.yaml"
	// 如果用户有传配置文件，就使用用户传入的配置文件
	if len(os.Args) == 2 {
		configFilename = os.Args[1]
	}
	configPath = path.Join(home, configFilename)

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
}

// 保存配置到配置文件
func (config *Config) SaveConfig() {
	viper.Set("listen", config.ListenAddr)
	viper.Set("remote", config.RemoteAddr)
	viper.Set("obf_domain", config.ObfDomain)
	err := viper.WriteConfigAs(configPath)
	if err != nil {
		fmt.Errorf("保存配置到文件 %s 出错: %s", configPath, err)
	} else {
		log.Printf("保存配置到文件 %s 成功\n", configPath)
	}
}

// 读取配置文件
func (config *Config) ReadConfig() {
	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		var configFileNotFoundError viper.ConfigFileNotFoundError
		if errors.As(err, &configFileNotFoundError) {
			log.Printf("配置文件 %s 不存在，使用默认配置\n", configPath)
		}
	} else {
		log.Printf("从文件 %s 中读取配置\n", configPath)
		err := viper.Unmarshal(config)
		if err != nil {
			log.Fatalf("格式不合法的 YAML 配置文件: %s", err)
		}
	}
}
