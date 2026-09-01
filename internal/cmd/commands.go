package cmd

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/tomleelong/bright-local-cli/internal/api"
	"github.com/spf13/cobra"
)

func newLocationsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "locations", Short: "Manage BrightLocal locations"}
	cmd.AddCommand(newListCmd("list", "List locations", "/manage/v1/locations"))
	cmd.AddCommand(newGetCmd("get <location-id>", "Get a location", "/manage/v1/locations/"))
	cmd.AddCommand(
		newWriteCmd("create", "Create a location", "create", "location", "POST", "/manage/v1/locations"),
		newWriteCmd("update <location-id>", "Update a location", "update", "location", "PATCH", "/manage/v1/locations/%s"),
	)
	return cmd
}

func newClientsCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "clients", Short: "Manage BrightLocal clients"}
	cmd.AddCommand(newListCmd("list", "List clients", "/manage/v1/clients"))
	cmd.AddCommand(newGetCmd("get <client-id>", "Get a client", "/manage/v1/clients/"))
	cmd.AddCommand(
		newWriteCmd("create", "Create a client", "create", "client", "POST", "/manage/v1/clients"),
		newWriteCmd("update <client-id>", "Update a client", "update", "client", "PATCH", "/manage/v1/clients/%s"),
	)
	return cmd
}

func newRankTrackerCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "rank-tracker", Short: "Read Local Rank Tracker reports"}
	reports := &cobra.Command{Use: "reports", Short: "Read Local Rank Tracker reports"}
	reports.AddCommand(
		newPathCmd("list", "List rank-tracker reports", "/manage/v1/lrt/reports"),
		newPathCmd("get <report-id>", "Get a rank-tracker report", "/manage/v1/lrt/reports/%s"),
		newPathCmd("history <report-id>", "Get rank-tracker report history", "/manage/v1/lrt/reports/%s/history"),
		newPathCmd("result <report-id>", "Get the latest rank-tracker report result", "/manage/v1/lrt/reports/%s/result"),
	)
	cmd.AddCommand(reports)
	return cmd
}

func newSearchGridCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "search-grid", Short: "Read Local Search Grid reports and rankings"}
	reports := &cobra.Command{Use: "reports", Short: "Read Local Search Grid reports"}
	reports.AddCommand(
		newPathCmd("list", "List search-grid reports", "/manage/v1/lsg/reports"),
		newPathCmd("get <report-id>", "Get a search-grid report", "/manage/v1/lsg/reports/%s"),
	)
	runs := &cobra.Command{Use: "runs", Short: "Read Local Search Grid report runs"}
	runs.AddCommand(
		newPathCmd("list <report-id> <keyword-id>", "List runs for a keyword", "/manage/v1/lsg/reports/%s/keywords/%s/runs"),
		newPathCmd("get <report-id> <run-id> <keyword-id>", "Get a keyword run", "/manage/v1/lsg/reports/%s/runs/%s/keywords/%s"),
	)
	rankings := &cobra.Command{Use: "rankings", Short: "Read Local Search Grid rankings"}
	rankings.AddCommand(
		newPathCmd("competitors <report-id> <run-id> <keyword-id>", "List competitors for a keyword run", "/manage/v1/lsg/reports/%s/runs/%s/keywords/%s/competitors"),
		newPathCmd("competitor <report-id> <run-id> <keyword-id> <competitor-id>", "Get competitor rankings", "/manage/v1/lsg/reports/%s/runs/%s/keywords/%s/competitors/%s"),
	)
	cmd.AddCommand(reports, runs, rankings)
	return cmd
}

func newReputationCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "reputation", Short: "Read Reputation Manager reports and reviews"}
	reports := &cobra.Command{Use: "reports", Short: "Read Reputation Manager reports"}
	reports.AddCommand(
		newPathCmd("list", "List reputation reports", "/manage/v1/rm/reports"),
		newPathCmd("get <report-id>", "Get a reputation report", "/manage/v1/rm/reports/%s"),
		newPathCmd("reviews <report-id>", "Get reviews for a reputation report", "/manage/v1/rm/reports/%s/reviews"),
	)
	cmd.AddCommand(reports)
	return cmd
}

func newCitationBuilderCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "citation-builder", Short: "Read Citation Builder campaigns"}
	cmd.AddCommand(
		newPathCmd("list", "List Citation Builder campaigns", "/manage/v1/citation-builder"),
		newPathCmd("get <campaign-id>", "Get a Citation Builder campaign", "/manage/v1/citation-builder/%s"),
	)
	return cmd
}

func newReferenceCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "reference", Short: "Read BrightLocal reference data"}
	cmd.AddCommand(
		newPathCmd("time-options", "Get available time options", "/manage/v1/time-options"),
		newPathCmd("white-label-profiles", "List white-label profiles", "/manage/v1/white-label-profiles"),
	)
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

func newPathCmd(use, short, pathTemplate string) *cobra.Command {
	argCount := strings.Count(pathTemplate, "%s")
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.ExactArgs(argCount),
		RunE: func(cmd *cobra.Command, args []string) error {
			escapedArgs := make([]any, len(args))
			for i, arg := range args {
				escapedArgs[i] = url.PathEscape(arg)
			}
			path := fmt.Sprintf(pathTemplate, escapedArgs...)
			values, err := cmd.Flags().GetStringArray("query")
			if err != nil {
				return err
			}
			query, err := parseQuery(values)
			if err != nil {
				return err
			}
			return runGet(cmd.Context(), cmd.OutOrStdout(), path, query)
		},
	}
	cmd.Flags().StringArrayP("query", "q", nil, "Query parameter in key=value form (repeatable)")
	return cmd
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

func newWriteCmd(use, short, operation, resource, method, pathTemplate string) *cobra.Command {
	argCount := strings.Count(pathTemplate, "%s")
	var data string
	var confirm, dryRun bool
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  "Send a JSON request to the BrightLocal Management API. Requires --confirm unless --dry-run is used.",
		Args:  cobra.ExactArgs(argCount),
		RunE: func(cmd *cobra.Command, args []string) error {
			body, err := validateJSONObject(data)
			if err != nil {
				return err
			}
			path := pathTemplate
			if argCount > 0 {
				escapedArgs := make([]any, len(args))
				for i, arg := range args {
					escapedArgs[i] = url.PathEscape(arg)
				}
				path = fmt.Sprintf(pathTemplate, escapedArgs...)
			}
			target := resource
			if len(args) > 0 {
				target += " " + args[0]
			}
			return runWrite(cmd, method, operation, target, path, body, confirm, dryRun)
		},
	}
	cmd.Flags().StringVar(&data, "data", "", "JSON request body")
	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm this write operation")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show the request without sending it")
	_ = cmd.MarkFlagRequired("data")
	return cmd
}

func validateJSONObject(data string) ([]byte, error) {
	body := []byte(data)
	if path, ok := strings.CutPrefix(data, "@"); ok {
		var err error
		body, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read --data file: %w", err)
		}
	}
	var value map[string]json.RawMessage
	if err := json.Unmarshal(body, &value); err != nil || value == nil {
		return nil, fmt.Errorf("--data must be a JSON object")
	}
	return body, nil
}

func runWrite(cmd *cobra.Command, method, operation, target, path string, body []byte, confirmed, dryRun bool) error {
	if dryRun {
		return printDryRun(cmd.OutOrStdout(), method, path, body)
	}
	if !confirmed {
		return fmt.Errorf("refusing to %s %s without --confirm", operation, target)
	}
	if isInteractive(cmd.InOrStdin()) {
		fmt.Fprintf(cmd.ErrOrStderr(), "About to %s %s. Proceed? [y/N]: ", operation, target)
		answer, err := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
		if err != nil && len(answer) == 0 {
			return fmt.Errorf("read confirmation: %w", err)
		}
		if strings.ToLower(strings.TrimSpace(answer)) != "y" && strings.ToLower(strings.TrimSpace(answer)) != "yes" {
			return fmt.Errorf("write cancelled")
		}
	}

	key := apiKey
	if key == "" {
		key = os.Getenv("BRIGHTLOCAL_API_KEY")
	}
	client, err := api.New(key, baseURL, nil)
	if err != nil {
		return err
	}
	response, err := client.Write(cmd.Context(), method, path, body)
	if err != nil {
		return err
	}
	return printJSON(cmd.OutOrStdout(), response)
}

func printDryRun(output io.Writer, method, path string, body []byte) error {
	request := struct {
		DryRun bool            `json:"dry_run"`
		Method string          `json:"method"`
		Path   string          `json:"path"`
		Body   json.RawMessage `json:"body"`
	}{
		DryRun: true,
		Method: method,
		Path:   path,
		Body:   body,
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(request)
}

func isInteractive(input io.Reader) bool {
	file, ok := input.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
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
