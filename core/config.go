package core

import (
	"encoding/json"
	"errors"
	"os"
)

type config struct {
	configName string
	LogsDir    string                  `json:"logsDir"`
	Module     map[string]ModuleConfig `json:"module"`
	HttpPort   uint16                  `json:"httpPort"`
	Ext        ExtraConfig
}

type ModuleConfig struct {
	Enable bool        `json:"enable"`
	Config interface{} `json:"config"`
}
type ExtraConfig struct {
}

func (c *config) getConfig(module *Module) interface{} {
	if moduleConf, has := c.Module[module.name]; has {
		return moduleConf.Config
	}
	return nil
}

func (c *config) saveConfig(module *Module, conf interface{}) {
	if moduleConf, has := c.Module[module.name]; has {
		moduleConf.Config = conf
		c.Module[module.name] = moduleConf
	}
}
func (c *config) initConfig(module *Module, conf interface{}) {
	c.Module[module.name] = ModuleConfig{
		Enable: true,
		Config: conf,
	}
}

func (c *config) load() error {
	if c.configName == "" {
		return errors.New("配置文件名为空")
	}

	file, err := os.ReadFile(c.configName)
	if err != nil {
		return err
	}
	return json.Unmarshal(file, c)
}

func (c *config) save() error {
	marshal, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(c.configName, marshal, 0644)
}

func Unmarshal(src, dst interface{}) error {
	marshal, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(marshal, dst)
}
