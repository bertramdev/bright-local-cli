package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	apiKey  string
	baseURL string
)

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "bright-local",
		Short: "A command-line client for BrightLocal",
		Long:  "A safe, read-only command-line client for BrightLocal's Management API.",
	}
	root.PersistentFlags().StringVar(&apiKey, "api-key", "", "BrightLocal API key (default: BRIGHTLOCAL_API_KEY)")
	root.PersistentFlags().StringVar(&baseURL, "base-url", "https://api.brightlocal.com", "BrightLocal API base URL")
	_ = root.PersistentFlags().MarkHidden("base-url")
	root.AddCommand(newLocationsCmd(), newClientsCmd(), newCategoriesCmd(), newAPICmd())
	return root
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
