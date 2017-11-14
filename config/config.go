package main

import(
	"fmt"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	Counter []Instance `yaml:"count"`
	XXX map[string]interface{} `yaml:",inline"`
}

type Instance struct {
	Name string `yaml:"name"`
	Groups map[string]Group `yaml:"group"`
	XXX map[string]interface{} `yaml:",inline"`
}

type Group struct {
	Targets []string `yaml:"targets"`
	Kind string `yaml:"kind"`
	XXX map[string]interface{} `yaml:",inline"`
}

func checkOverflow(m map[string]interface{}, ctx string) error {
	if len(m) > 0 {
		var keys []string
		for k := range m {
			keys = append(keys, k)
		}
		return fmt.Errorf("unknown fields in %s: %s", ctx, strings.Join(keys, ", "))
	}
	return nil
}

func (s *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Config
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	if err := checkOverflow(s.XXX, "config"); err != nil {
		return err
	}
	return nil
}

func (s *Instance) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Instance
	if err := unmarshal((*plain)(s)); err != nil {
		return err
	}
	if err := checkOverflow(s.XXX, "instance"); err != nil {
		return err
	}
	return nil
}

type SafeConfig struct {
	sync.RWMutex
	C *Config
}

func (sc *SafeConfig) ReloadConfig(confFile string) (err error) {
	var c = &Config{}

	yamlFile, err := ioutil.ReadFile(confFile)
	if err != nil {
		return fmt.Errorf("Error reading config file: %s", err)
	}

	if err := yaml.Unmarshal(yamlFile, c); err != nil {
		return fmt.Errorf("Error parsing config file: %s", err)
	}

	sc.Lock()
	sc.C = c
	sc.Unlock()
	
	return nil
}
