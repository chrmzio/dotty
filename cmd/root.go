/*
Copyright © 2025 dotty <chrmzio@pm.me>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/chrmzio/dotty/internal/config"
	"github.com/spf13/cobra"
)

var (
	cfgManager *config.Manager

	configFile string
	verbose    bool
)

var rootCmd = &cobra.Command{
	Use:   "dotty",
	Short: "Dotty manages dotfiles the best she can",
	Long: `Dotty manages your dotfiles by syncing them with a Git repository.

Configuration file is located at ~/.config/dotty/config.toml and contains:
  - repo_url: URL to your GitHub repository
  - dotfiles: List of file paths to manage

Example configuration:
  repo_url = "https://github.com/username/dotfiles"
  dotfiles = [
    "~/.bashrc",
    "~/.vimrc",
    "~/.config/nvim/init.vim"
  ]`,
	PersistentPreRunE: initializeConfig,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no subcommand is provided, show the current configuration
		cfg := cfgManager.Get()

		if cfg.RepoURL == "" && len(cfg.Dotfiles) == 0 {
			fmt.Println("No configuration found. Use 'dotty init' to set up your dotfiles.")
			return nil
		}

		fmt.Printf("Configuration file: %s\n", cfgManager.GetPath())
		fmt.Printf("Repository URL: %s\n", cfg.RepoURL)

		if len(cfg.Dotfiles) > 0 {
			fmt.Println("Managed dotfiles:")
			for _, dotfile := range cfg.Dotfiles {
				fmt.Printf("  - %s\n", dotfile)
			}
		} else {
			fmt.Println("No dotfiles configured.")
		}

		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configFile, "config", "c", "",
		"config file (default is $HOME/.config/dotty/config.toml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false,
		"verbose output")

	rootCmd.AddCommand(newInitCmd())
	rootCmd.AddCommand(newAddCmd())
	rootCmd.AddCommand(newRemoveCmd())
	rootCmd.AddCommand(newSyncCmd())
	rootCmd.AddCommand(newStatusCmd())
}

func initializeConfig(cmd *cobra.Command, args []string) error {
	var err error

	if configFile != "" {
		cfgManager, err = config.NewManagerWithPath(configFile)
	} else {
		cfgManager, err = config.NewManager()
	}

	if err != nil {
		return fmt.Errorf("failed to initialize configuration: %w", err)
	}

	if verbose {
		fmt.Printf("Using config file: %s\n", cfgManager.GetPath())
	}

	return nil
}

func newInitCmd() *cobra.Command {
	var repoURL string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize dotty configuration",
		Long:  "Initialize dotty with a repository URL for your dotfiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			if repoURL == "" {
				return fmt.Errorf("repository URL is required")
			}

			if err := cfgManager.SetRepoURL(repoURL); err != nil {
				return fmt.Errorf("failed to set repository URL: %w", err)
			}

			fmt.Printf("Initialized dotty with repository: %s\n", repoURL)
			return nil
		},
	}

	cmd.Flags().StringVarP(&repoURL, "repo", "r", "", "Repository URL")
	cmd.MarkFlagRequired("repo")

	return cmd
}

func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add [file...]",
		Short: "Add dotfiles to manage",
		Long:  "Add one or more dotfiles to be managed by dotty",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, file := range args {
				if err := cfgManager.AddDotfile(file); err != nil {
					return fmt.Errorf("failed to add %s: %w", file, err)
				}
				fmt.Printf("Added: %s\n", file)
			}
			return nil
		},
	}
}

func newRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove [file...]",
		Short:   "Remove dotfiles from management",
		Aliases: []string{"rm"},
		Args:    cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, file := range args {
				if err := cfgManager.RemoveDotfile(file); err != nil {
					return fmt.Errorf("failed to remove %s: %w", file, err)
				}
				fmt.Printf("Removed: %s\n", file)
			}
			return nil
		},
	}
}

func newSyncCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "sync",
		Short: "Sync dotfiles with repository",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := cfgManager.Get()
			if cfg.RepoURL == "" {
				return fmt.Errorf("no repository URL configured, run 'dotty init' first")
			}

			fmt.Printf("Syncing with %s...\n", cfg.RepoURL)
			// TODO: Implement sync logic
			return nil
		},
	}
}

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show status of managed dotfiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := cfgManager.Get()

			if len(cfg.Dotfiles) == 0 {
				fmt.Println("No dotfiles configured")
				return nil
			}

			expandedPaths, err := cfgManager.GetExpandedDotfiles()
			if err != nil {
				return fmt.Errorf("failed to expand paths: %w", err)
			}

			fmt.Println("Dotfile status:")
			for _, dotfile := range cfg.Dotfiles {
				expandedPath := expandedPaths[dotfile]

				// Check if file exists using expanded path
				info, err := os.Stat(expandedPath)
				if err == nil {
					if info.IsDir() {
						fmt.Printf("  📁 %s (directory)\n", dotfile)
					} else {
						fmt.Printf("  ✓ %s\n", dotfile)
					}
				} else if os.IsNotExist(err) {
					fmt.Printf("  ✗ %s (not found)\n", dotfile)
					if verbose {
						fmt.Printf("     Expanded to: %s\n", expandedPath)
					}
				} else {
					fmt.Printf("  ? %s (error: %v)\n", dotfile, err)
				}
			}

			return nil
		},
	}
}
