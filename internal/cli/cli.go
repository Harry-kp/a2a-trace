package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Version information (set at build time)
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)

// Config holds CLI configuration
type Config struct {
	Port     int
	UIPort   int
	DBPath   string
	Verbose  bool
	NoUI     bool
	Command  []string
}

// ParseArgs parses command line arguments and returns a Config
func ParseArgs() (*Config, error) {
	cfg := &Config{}

	rootCmd := &cobra.Command{
		Use:   "a2a-trace [flags] -- <command> [args...]",
		Short: "A2A Trace - Visual debugger for A2A multi-agent systems",
		Long: `A2A Trace intercepts and visualizes Agent-to-Agent (A2A) protocol 
communications to help you debug multi-agent systems.

Usage:
  a2a-trace -- node my-agent.js
  a2a-trace -- python agent.py --port 8080
  a2a-trace --port 8081 -- ./my-go-agent

The command after '--' is run with HTTP_PROXY set to route traffic
through A2A Trace for inspection.`,
		Example: `  # Trace a Node.js agent
  a2a-trace -- node my-agent.js

  # Trace a Python agent with custom port
  a2a-trace --port 9000 -- python agent.py

  # Trace without opening UI
  a2a-trace --no-ui -- ./my-agent`,
		Version: formatVersion(),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Find the command after --
			dashIndex := -1
			for i, arg := range os.Args {
				if arg == "--" {
					dashIndex = i
					break
				}
			}

			if dashIndex == -1 || dashIndex == len(os.Args)-1 {
				return fmt.Errorf("no command specified after '--'\n\nUsage: a2a-trace [flags] -- <command> [args...]")
			}

			cfg.Command = os.Args[dashIndex+1:]
			return nil
		},
		SilenceUsage: true,
	}

	// Flags
	rootCmd.Flags().IntVarP(&cfg.Port, "port", "p", 8080, "Proxy port")
	rootCmd.Flags().IntVar(&cfg.UIPort, "ui-port", 0, "UI port (default: same as proxy port)")
	rootCmd.Flags().StringVar(&cfg.DBPath, "db", "", "SQLite database path (default: in-memory)")
	rootCmd.Flags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Verbose output")
	rootCmd.Flags().BoolVar(&cfg.NoUI, "no-ui", false, "Don't serve the web UI")

	// Parse without the -- and everything after it
	var argsToparse []string
	for _, arg := range os.Args[1:] {
		if arg == "--" {
			break
		}
		argsToparse = append(argsToparse, arg)
	}

	rootCmd.SetArgs(argsToparse)

	if err := rootCmd.Execute(); err != nil {
		return nil, err
	}

	// Set UI port to proxy port if not specified
	if cfg.UIPort == 0 {
		cfg.UIPort = cfg.Port
	}

	return cfg, nil
}

// formatVersion returns formatted version information
func formatVersion() string {
	return fmt.Sprintf("%s (commit: %s, built: %s)", Version, Commit, BuildDate)
}

// PrintBanner prints the startup banner
func PrintBanner(cfg *Config) {
	banner := `
   ___   ___   ___     ______                    
  / _ | |_  | / _ |   /_  __/_____ ___ _____ ___ 
 / __ | / __/ / __ |    / / / __/ _ '/ __/ -_|_-< 
/_/ |_|/____/_/ |_|    /_/ /_/  \_,_/\__/\__/___/ 
                                                  
`
	fmt.Print(banner)
	fmt.Printf("  Version: %s\n", Version)
	fmt.Printf("  Proxy:   http://127.0.0.1:%d\n", cfg.Port)
	if !cfg.NoUI {
		fmt.Printf("  UI:      http://127.0.0.1:%d/ui\n", cfg.UIPort)
	}
	fmt.Printf("  Command: %s\n", strings.Join(cfg.Command, " "))
	fmt.Println()
	fmt.Println("  📡 Intercepting A2A traffic...")
	fmt.Println()
}

// PrintError prints an error message
func PrintError(msg string, err error) {
	fmt.Fprintf(os.Stderr, "❌ %s: %v\n", msg, err)
}

// PrintSuccess prints a success message
func PrintSuccess(msg string) {
	fmt.Printf("✅ %s\n", msg)
}

// PrintInfo prints an info message
func PrintInfo(msg string) {
	fmt.Printf("ℹ️  %s\n", msg)
}

// PrintWarning prints a warning message
func PrintWarning(msg string) {
	fmt.Printf("⚠️  %s\n", msg)
}

