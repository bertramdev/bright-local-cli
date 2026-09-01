package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/bertramdev/bright-local-cli/internal/api"
	"github.com/spf13/cobra"
)

func newLocationsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "locations", Short: "Read BrightLocal locations"}
	cmd.AddCommand(newListCmd("list", "List locations", "/manage/v1/locations"))
	cmd.AddCommand(newGetCmd("get <location-id>", "Get a location", "/manage/v1/locations/"))
	return cmd
}

func newClientsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "clients", Short: "Read BrightLocal clients"}
	cmd.AddCommand(newListCmd("list", "List clients", "/manage/v1/clients"))
	cmd.AddCommand(newGetCmd("get <client-id>", "Get a client", "/manage/v1/clients/"))
	return cmd
}

func newCategoriesCmd() *cobra.Command {
	var query string
	cmd := &cobra.Command{
		Use:   "categories <country>",
		Short: "List business categories for an ISO 3166-1 alpha-2 country code",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			values := url.Values{}
			if query != "" {
				values.Set("query", query)
			}
			return runGet(cmd.Context(), cmd.OutOrStdout(), "/manage/v1/business-categories/"+url.PathEscape(strings.ToUpper(args[0])), values)
		},
	}
	cmd.Flags().StringVarP(&query, "query", "q", "", "Filter categories by name")
	return cmd
}

func newAPICmd() *cobra.Command {
	cmd := &cobra.Command{Use: "api", Short: "Make read-only Management API requests"}
	cmd.AddCommand(&cobra.Command{
		Use:   "get <path>",
		Short: "GET a BrightLocal Management API path",
		Long:  "GET a path below /manage/v1. Repeat --query for query parameters.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !strings.HasPrefix(args[0], "/manage/v1/") {
				return fmt.Errorf("path must start with /manage/v1/")
			}
			values, err := cmd.Flags().GetStringArray("query")
			if err != nil {
				return err
			}
			query, err := parseQuery(values)
			if err != nil {
				return err
			}
			return runGet(cmd.Context(), cmd.OutOrStdout(), args[0], query)
		},
	})
	cmd.Commands()[0].Flags().StringArrayP("query", "q", nil, "Query parameter in key=value form (repeatable)")
	return cmd
}

func newListCmd(use, short, path string) *cobra.Command {
	var page, perPage int
	var query, clientType string
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if page < 1 {
				return fmt.Errorf("page must be at least 1")
			}
			if perPage < 1 || perPage > 100 {
				return fmt.Errorf("per-page must be between 1 and 100")
			}
			values := url.Values{}
			if page > 0 {
				values.Set("page", strconv.Itoa(page))
			}
			if perPage > 0 {
				values.Set("num_per_page", strconv.Itoa(perPage))
			}
			if query != "" {
				values.Set("query", query)
			}
			if clientType != "" {
				values.Set("type", clientType)
			}
			return runGet(cmd.Context(), cmd.OutOrStdout(), path, values)
		},
	}
	cmd.Flags().IntVarP(&page, "page", "p", 1, "Page number")
	cmd.Flags().IntVar(&perPage, "per-page", 10, "Results per page (1-100)")
	cmd.Flags().StringVarP(&query, "query", "q", "", "Free-text search")
	if path == "/manage/v1/clients" {
		cmd.Flags().StringVar(&clientType, "type", "", "Comma-separated client types: na, lead, client")
	}
	return cmd
}

func newGetCmd(use, short, pathPrefix string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGet(cmd.Context(), cmd.OutOrStdout(), pathPrefix+url.PathEscape(args[0]), nil)
		},
	}
}

func runGet(ctx context.Context, output io.Writer, path string, query url.Values) error {
	key := apiKey
	if key == "" {
		key = os.Getenv("BRIGHTLOCAL_API_KEY")
	}
	client, err := api.New(key, baseURL, nil)
	if err != nil {
		return err
	}
	body, err := client.Get(ctx, path, query)
	if err != nil {
		return err
	}
	return printJSON(output, body)
}

func printJSON(output io.Writer, body []byte) error {
	var value any
	if err := json.Unmarshal(body, &value); err != nil {
		_, writeErr := fmt.Fprintln(output, string(body))
		return writeErr
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}

func parseQuery(values []string) (url.Values, error) {
	query := url.Values{}
	for _, value := range values {
		key, val, ok := strings.Cut(value, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid query %q: use key=value", value)
		}
		query.Add(key, val)
	}
	return query, nil
}
