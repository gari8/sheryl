package config

import (
	"github.com/gari8/sheryl/pkg/types"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"time"
)

type Config struct {
	Steps []*Step         `mapstructure:"steps"`
	Env   []*types.KeyVal `mapstructure:"env" yaml:"env,omitempty"`
}

type Step struct {
	Name     string `mapstructure:"name" yaml:"name,omitempty"`
	Cmd      string `mapstructure:"cmd" yaml:"cmd,omitempty"`
	Delay    string `mapstructure:"delay" yaml:"delay,omitempty"`
	Retries  int    `mapstructure:"retries" yaml:"retries,omitempty"`
	Interval string `mapstructure:"interval" yaml:"interval,omitempty"`
}

func Load(filePath string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(filePath)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}
	var config Config
	if err := v.Unmarshal(&config, viper.DecodeHook(mapToKeyValHook(getAllEnv()))); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *Config) ToTypesSteps() ([]*types.Step, error) {
	var result []*types.Step
	var env []string
	for _, e := range c.Env {
		env = append(env, e.String())
	}
	for _, s := range c.Steps {
		step := &types.Step{
			Name: s.Name,
			Cmd:  s.Cmd,
			Env:  env,
		}
		if s.Delay != "" {
			delay, err := time.ParseDuration(s.Delay)
			if err != nil {
				return nil, err
			}
			step.Delay = delay
		}
		if s.Retries > 0 {
			step.Retries = s.Retries
			if s.Interval != "" {
				interval, err := time.ParseDuration(s.Interval)
				if err != nil {
					return nil, err
				}
				step.Interval = interval
			}
		}
		result = append(result, step)
	}
	return result, nil
}

func (c *Config) String() string {
	b, err := yaml.Marshal(c)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
	return string(b)
}

func mapToKeyValHook(extraMaps ...map[string]any) mapstructure.DecodeHookFuncType {
	return func(
		from reflect.Type,
		to reflect.Type,
		data any,
	) (any, error) {
		if from.Kind() == reflect.Map && from.Key().Kind() == reflect.String && to == reflect.TypeOf([]*types.KeyVal{}) {
			original, ok := data.(map[string]any)
			if !ok {
				return data, nil
			}

			// 外部から受け取ったマップも受け取る
			// 優先度は yaml > extra
			for _, m := range extraMaps {
				for k, v := range m {
					if _, ok := original[k]; !ok {
						original[k] = v
					}
				}
			}

			var result []*types.KeyVal
			for k, v := range original {
				value, ok := v.(string)
				if !ok {
					continue
				}
				result = append(result, &types.KeyVal{Key: strings.ToUpper(k), Value: value})
			}
			return result, nil
		}
		return data, nil
	}
}

func getAllEnv() map[string]any {
	envMap := make(map[string]any)
	for _, env := range os.Environ() {
		envSet := strings.Split(env, "=")
		if len(envSet) > 1 {
			envMap[envSet[0]] = envSet[1]
		}
	}
	return envMap
}
