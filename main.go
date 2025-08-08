package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

func getConfigFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Error getting user home directory: %v", err)
	}
	return filepath.Join(homeDir, ".chcc.yaml")
}

var rootCmd = &cobra.Command{
	Use:   "chcc",
	Short: "CHCC - API Site Configuration Manager",
	Long: `A CLI tool for managing API site configurations.
You can add, update, list API sites and set default sites.`,
	Run: func(cmd *cobra.Command, args []string) {
		listAPISites()
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all API sites",
	Long:  "Display all configured API sites and show the default site",
	Run: func(cmd *cobra.Command, args []string) {
		listAPISites()
	},
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add or update an API site",
	Long:  "Add a new API site or update an existing one with name, URL, and token",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		url, _ := cmd.Flags().GetString("url")
		token, _ := cmd.Flags().GetString("token")
		addAPISite(name, url, token)
	},
}

var setDefaultCmd = &cobra.Command{
	Use:   "set-default",
	Short: "Set default API site",
	Long:  "Set a specific API site as the default",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		setDefaultAPISite(name)
	},
}

var removeCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove an API site",
	Long:  "Remove an API site from the configuration",
	Run: func(cmd *cobra.Command, args []string) {
		name, _ := cmd.Flags().GetString("name")
		removeAPISite(name)
	},
}

func init() {
	addCmd.Flags().StringP("name", "n", "", "API site name (required)")
	addCmd.Flags().StringP("url", "u", "", "API site base URL (required)")
	addCmd.Flags().StringP("token", "t", "", "API site token (required)")
	addCmd.MarkFlagRequired("name")
	addCmd.MarkFlagRequired("url")
	addCmd.MarkFlagRequired("token")

	setDefaultCmd.Flags().StringP("name", "n", "", "API site name (required)")
	setDefaultCmd.MarkFlagRequired("name")

	removeCmd.Flags().StringP("name", "n", "", "API site name (required)")
	removeCmd.MarkFlagRequired("name")

	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(setDefaultCmd)
	rootCmd.AddCommand(removeCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func loadConfig() *Config {
	configFile := getConfigFilePath()
	
	// 检查文件是否存在
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// 如果配置文件不存在，创建一个新的空配置
		config := &Config{
			APISites:       []APISite{},
			DefaultAPISite: "",
		}
		fmt.Printf("Creating new config file at: %s\n", configFile)
		return config
	}
	
	config, err := LoadConfig(configFile)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}
	return config
}

func listAPISites() {
	config := loadConfig()
	
	if len(config.APISites) == 0 {
		fmt.Println("No API sites configured.")
		fmt.Println("\nUse 'chcc add --help' to add your first API site.")
		return
	}

	fmt.Println("=== CHCC Configuration ===")
	config.PrintConfig()

	fmt.Println("\n=== Default API Site ===")
	defaultSite := config.GetDefaultAPISite()
	if defaultSite != nil {
		fmt.Printf("Name: %s\n", defaultSite.Name)
		fmt.Printf("URL: %s\n", defaultSite.BaseURL)
		fmt.Printf("Token: %s...\n", defaultSite.Token[:min(len(defaultSite.Token), 10)])
	} else {
		fmt.Println("No default API site set")
		fmt.Println("Use 'chcc set-default --name <site-name>' to set a default site.")
	}
}

func addAPISite(name, url, token string) {
	config := loadConfig()

	existing := config.GetAPISiteByName(name)
	if existing != nil {
		fmt.Printf("Updating existing API site: %s\n", name)
	} else {
		fmt.Printf("Adding new API site: %s\n", name)
	}

	config.AddOrUpdateAPISite(name, url, token)

	if config.DefaultAPISite == "" {
		config.DefaultAPISite = name
		fmt.Printf("Set %s as default API site (first site added)\n", name)
	}

	configFile := getConfigFilePath()
	err := config.SaveConfig(configFile)
	if err != nil {
		log.Fatalf("Error saving config: %v", err)
	}

	fmt.Println("Configuration saved successfully!")
}

func setDefaultAPISite(name string) {
	config := loadConfig()

	if config.SetDefaultAPISite(name) {
		fmt.Printf("Set %s as default API site\n", name)
		
		configFile := getConfigFilePath()
		err := config.SaveConfig(configFile)
		if err != nil {
			log.Fatalf("Error saving config: %v", err)
		}
		
		fmt.Println("Configuration saved successfully!")

		fmt.Println("Setting user environment variables...")
		err = config.SetEnvironmentVariables(name)
		if err != nil {
			fmt.Printf("Warning: Failed to set environment variables: %v\n", err)
			fmt.Println("You may need to set them manually:")
			site := config.GetAPISiteByName(name)
			if site != nil {
				fmt.Printf("  ANTHROPIC_BASE_URL=%s\n", site.BaseURL)
				fmt.Printf("  ANTHROPIC_AUTH_TOKEN=%s\n", site.Token)
			}
		} else {
			fmt.Println("User environment variables set successfully!")
			fmt.Println("To apply this change to your current terminal session, run the following command:")
			site := config.GetAPISiteByName(name)
			if site != nil {
				switch runtime.GOOS {
				case "windows":
					fmt.Println("\nFor Command Prompt (cmd.exe):")
					fmt.Printf("set ANTHROPIC_BASE_URL=%s\n", site.BaseURL)
					fmt.Printf("set ANTHROPIC_AUTH_TOKEN=%s\n", site.Token)
					fmt.Println("\nFor PowerShell:")
					fmt.Printf("$env:ANTHROPIC_BASE_URL=\"%s\"\n", site.BaseURL)
					fmt.Printf("$env:ANTHROPIC_AUTH_TOKEN=\"%s\"\n", site.Token)
				case "linux", "darwin":
					fmt.Println("\nFor Bash/Zsh:")
					fmt.Printf("export ANTHROPIC_BASE_URL=%s\n", site.BaseURL)
					fmt.Printf("export ANTHROPIC_AUTH_TOKEN=%s\n", site.Token)
				default:
					fmt.Println("\nUnsupported OS. Please set the environment variables manually.")
				}
			}
		}
	} else {
		fmt.Printf("Error: API site '%s' not found\n", name)
		if len(config.APISites) > 0 {
			fmt.Println("Available API sites:")
			for _, site := range config.APISites {
				fmt.Printf("  - %s\n", site.Name)
			}
		} else {
			fmt.Println("No API sites configured. Use 'chcc add' to add one first.")
		}
	}
}

func removeAPISite(name string) {
	config := loadConfig()

	if config.RemoveAPISite(name) {
		fmt.Printf("Removed API site: %s\n", name)
		
		configFile := getConfigFilePath()
		err := config.SaveConfig(configFile)
		if err != nil {
			log.Fatalf("Error saving config: %v", err)
		}
		
		fmt.Println("Configuration saved successfully!")
		
		// 如果删除后还有站点且有新的默认站点，更新环境变量
		if len(config.APISites) > 0 && config.DefaultAPISite != "" {
			fmt.Printf("New default API site: %s\n", config.DefaultAPISite)
			fmt.Println("Updating user environment variables...")
			err = config.SetEnvironmentVariables(config.DefaultAPISite)
			if err != nil {
				fmt.Printf("Warning: Failed to update environment variables: %v\n", err)
			} else {
				fmt.Println("User environment variables updated successfully!")
			}
		} else if len(config.APISites) == 0 {
			fmt.Println("No API sites remaining. You may want to clear environment variables manually:")
			fmt.Println("  ANTHROPIC_BASE_URL")
			fmt.Println("  ANTHROPIC_AUTH_TOKEN")
		}
	} else {
		fmt.Printf("Error: API site '%s' not found\n", name)
		if len(config.APISites) > 0 {
			fmt.Println("Available API sites:")
			for _, site := range config.APISites {
				fmt.Printf("  - %s\n", site.Name)
			}
		} else {
			fmt.Println("No API sites configured. Use 'chcc add' to add one first.")
		}
	}
}
