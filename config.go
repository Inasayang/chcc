package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"gopkg.in/yaml.v2"
)

type APISite struct {
	Name        string `yaml:"name"`
	BaseURL     string `yaml:"base_url"`
	Token       string `yaml:"token"`
}

type Config struct {
	APISites       []APISite `yaml:"api_sites"`
	DefaultAPISite string    `yaml:"default_api_site"`
}

func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

func (c *Config) GetDefaultAPISite() *APISite {
	for _, site := range c.APISites {
		if site.Name == c.DefaultAPISite {
			return &site
		}
	}
	if len(c.APISites) > 0 {
		return &c.APISites[0]
	}
	return nil
}

func (c *Config) GetAPISiteByName(name string) *APISite {
	for _, site := range c.APISites {
		if site.Name == name {
			return &site
		}
	}
	return nil
}

func (c *Config) PrintConfig() {
	fmt.Printf("Default API Site: %s\n", c.DefaultAPISite)
	fmt.Println("Available API Sites:")
	for i, site := range c.APISites {
		fmt.Printf("  %d. %s\n", i+1, site.Name)
		fmt.Printf("     URL: %s\n", site.BaseURL)
		fmt.Printf("     Token: %s...\n", site.Token[:min(len(site.Token), 20)])
	}
}

func (c *Config) SaveConfig(filename string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}

	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (c *Config) AddOrUpdateAPISite(name, baseURL, token string) {
	for i, site := range c.APISites {
		if site.Name == name {
			c.APISites[i].BaseURL = baseURL
			c.APISites[i].Token = token
			return
		}
	}
	
	newSite := APISite{
		Name:    name,
		BaseURL: baseURL,
		Token:   token,
	}
	c.APISites = append(c.APISites, newSite)
}

func (c *Config) SetDefaultAPISite(name string) bool {
	for _, site := range c.APISites {
		if site.Name == name {
			c.DefaultAPISite = name
			return true
		}
	}
	return false
}

func (c *Config) SetEnvironmentVariables(siteName string) error {
	site := c.GetAPISiteByName(siteName)
	if site == nil {
		return fmt.Errorf("API site '%s' not found", siteName)
	}

	baseURL := site.BaseURL
	authToken := site.Token

	if runtime.GOOS == "windows" {
		return setWindowsEnvVars(baseURL, authToken)
	} else {
		return setUnixEnvVars(baseURL, authToken)
	}
}

func setWindowsEnvVars(baseURL, authToken string) error {
	// 设置当前进程环境变量
	os.Setenv("ANTHROPIC_BASE_URL", baseURL)
	os.Setenv("ANTHROPIC_AUTH_TOKEN", authToken)

	// 设置用户级别的持久化环境变量（setx默认设置用户变量）
	cmd1 := exec.Command("setx", "ANTHROPIC_BASE_URL", baseURL)
	if err := cmd1.Run(); err != nil {
		fmt.Printf("Warning: Failed to set persistent ANTHROPIC_BASE_URL: %v\n", err)
	}

	cmd2 := exec.Command("setx", "ANTHROPIC_AUTH_TOKEN", authToken)
	if err := cmd2.Run(); err != nil {
		fmt.Printf("Warning: Failed to set persistent ANTHROPIC_AUTH_TOKEN: %v\n", err)
	}

	return nil
}

func setUnixEnvVars(baseURL, authToken string) error {
	// 设置当前进程环境变量
	os.Setenv("ANTHROPIC_BASE_URL", baseURL)
	os.Setenv("ANTHROPIC_AUTH_TOKEN", authToken)

	// 设置用户级别的持久化环境变量
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	bashrcPath := homeDir + "/.bashrc"
	zshrcPath := homeDir + "/.zshrc"

	exportLines := fmt.Sprintf("\n# CHCC API Configuration\nexport ANTHROPIC_BASE_URL=\"%s\"\nexport ANTHROPIC_AUTH_TOKEN=\"%s\"\n", baseURL, authToken)

	for _, rcFile := range []string{bashrcPath, zshrcPath} {
		if _, err := os.Stat(rcFile); err == nil {
			f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_WRONLY, 0644)
			if err != nil {
				continue
			}
			f.WriteString(exportLines)
			f.Close()
		}
	}

	return nil
}

func (c *Config) RemoveAPISite(name string) bool {
	for i, site := range c.APISites {
		if site.Name == name {
			// 删除该站点
			c.APISites = append(c.APISites[:i], c.APISites[i+1:]...)
			
			// 如果删除的是默认站点，清空默认设置
			if c.DefaultAPISite == name {
				c.DefaultAPISite = ""
				// 如果还有其他站点，设置第一个为默认
				if len(c.APISites) > 0 {
					c.DefaultAPISite = c.APISites[0].Name
				}
			}
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
