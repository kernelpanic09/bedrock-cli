package cmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available AWS profiles",
	Long: `List profiles from ~/.aws/config and ~/.aws/credentials.
The active profile (set via 'bedrock-cli profile use') is marked with an arrow.`,
	RunE: runProfileList,
}

func runProfileList(cmd *cobra.Command, args []string) error {
	profiles, err := collectAWSProfiles()
	if err != nil {
		return err
	}
	if len(profiles) == 0 {
		out.Dim("No AWS profiles found in ~/.aws/config or ~/.aws/credentials.")
		return nil
	}

	active := viper.GetString("aws-profile")

	if flagJSON {
		type row struct {
			Name   string `json:"name"`
			Active bool   `json:"active"`
		}
		var rows []row
		for _, p := range profiles {
			rows = append(rows, row{Name: p, Active: p == active})
		}
		enc := json.NewEncoder(out.Writer())
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	}

	headers := []string{"Profile", "Active"}
	var rows [][]string
	for _, p := range profiles {
		marker := ""
		if p == active {
			marker = "yes"
		}
		rows = append(rows, []string{p, marker})
	}
	out.Table(headers, rows)
	return nil
}

// collectAWSProfiles parses profile names from ~/.aws/config and ~/.aws/credentials.
func collectAWSProfiles() ([]string, error) {
	seen := make(map[string]bool)
	var profiles []string

	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolving home dir: %w", err)
	}

	for _, relPath := range []string{".aws/config", ".aws/credentials"} {
		path := filepath.Join(home, relPath)
		f, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if !strings.HasPrefix(line, "[") || !strings.HasSuffix(line, "]") {
				continue
			}
			name := line[1 : len(line)-1]
			// ~/.aws/config uses [profile NAME]; credentials uses [NAME].
			name = strings.TrimPrefix(name, "profile ")
			if name == "default" || name == "" {
				name = "default"
			}
			if !seen[name] {
				seen[name] = true
				profiles = append(profiles, name)
			}
		}
		f.Close()
	}
	return profiles, nil
}
