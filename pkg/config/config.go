package config

import (
	"github.com/gari8/sheryl/pkg/types"
	"gopkg.in/yaml.v3"
	"log/slog"
	"os"
	"strings"
	"time"
)

type Config struct {
	Steps []*Step         `yaml:"steps,omitempty"`
	Env   []*types.KeyVal `yaml:"env,omitempty"`
}

type Step struct {
	Name     string `yaml:"name,omitempty"`
	Cmd      string `yaml:"cmd,omitempty"`
	Delay    string `yaml:"delay,omitempty"`
	Retries  int    `yaml:"retries,omitempty"`
	Interval string `yaml:"interval,omitempty"`
}

func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	configStore := &struct {
		Steps []*Step           `yaml:"steps"`
		Env   map[string]string `yaml:"env"`
	}{}
	if err := unmarshal(&configStore); err != nil {
		return err
	}
	extraEnv := getAllEnv()
	var kvs []*types.KeyVal
	for k, v := range configStore.Env {
		// override yaml env
		if overrideVal, ok := extraEnv[k]; ok {
			v = overrideVal
		}
		kvs = append(kvs, &types.KeyVal{
			Key:   k,
			Value: v,
		})
	}
	c.Steps = configStore.Steps
	c.Env = kvs
	return nil
}

func Load(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
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

func getAllEnv() map[string]string {
	envMap := make(map[string]string)
	for _, env := range os.Environ() {
		envSet := strings.Split(env, "=")
		if len(envSet) > 1 {
			envMap[envSet[0]] = envSet[1]
		}
	}
	return envMap
}
