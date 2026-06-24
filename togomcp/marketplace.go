package togomcp

// Marketplace tools — let an agent discover and install togo ecosystem items
// (plugins, agents, skills) over MCP. Read-only: they return the exact
// `togo install …` command rather than shelling out.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
)

const (
	mktReposURL  = "https://to-go.dev/repos.json"
	mktPluginOrg = "togo-framework"
	mktAssetRepo = "togo-framework/claude-togo"
)

type mktRepo struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Kind        string `json:"kind"`
}

type mktGHContent struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type mktItem struct {
	Name, Kind, Description, Install string
}

func mktHTTPJSON(url string, v any) error {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, v)
}

// mktGHList lists the *.md asset names under a claude-togo dir (agents|commands).
func mktGHList(dir string) []string {
	var items []mktGHContent
	url := fmt.Sprintf("https://api.github.com/repos/%s/contents/%s", mktAssetRepo, dir)
	if err := mktHTTPJSON(url, &items); err != nil {
		return nil
	}
	var names []string
	for _, it := range items {
		if it.Type == "file" && strings.HasSuffix(it.Name, ".md") && !strings.EqualFold(it.Name, "README.md") {
			names = append(names, strings.TrimSuffix(it.Name, ".md"))
		}
	}
	sort.Strings(names)
	return names
}

// marketplaceItems aggregates plugins (to-go.dev catalog) + agents + skills (claude-togo).
func marketplaceItems() []mktItem {
	var items []mktItem
	var repos []mktRepo
	if err := mktHTTPJSON(mktReposURL, &repos); err == nil {
		for _, r := range repos {
			if r.Kind != "" && r.Kind != "plugin" {
				continue
			}
			name := r.Slug
			if name == "" {
				name = r.Name
			}
			items = append(items, mktItem{name, "plugin", r.Description, "togo install " + mktPluginOrg + "/" + name})
		}
	}
	for _, a := range mktGHList("agents") {
		items = append(items, mktItem{a, "agent", "togo AI agent (.claude/agents)", "togo install agent:" + a})
	}
	for _, sk := range mktGHList("commands") {
		items = append(items, mktItem{sk, "skill", "togo skill — Claude Code command /" + sk, "togo install skill:" + sk})
	}
	return items
}

func mktMatchKind(itemKind, filter string) bool {
	return filter == "" || strings.EqualFold(itemKind, filter)
}

// registerMarketplaceTools wires the marketplace_* tools onto the server.
func (s *Server) registerMarketplaceTools() {
	AddTool(s, "marketplace_search",
		"Search the togo marketplace — plugins, agents, and skills. Optional `kind` filter: plugin|agent|skill. Returns matches with their `togo install` command.",
		func(_ context.Context, in struct {
			Query string `json:"query"`
			Kind  string `json:"kind,omitempty"`
		}) (string, error) {
			q := strings.ToLower(in.Query)
			var out []string
			for _, it := range marketplaceItems() {
				if !mktMatchKind(it.Kind, in.Kind) {
					continue
				}
				if q == "" || strings.Contains(strings.ToLower(it.Name+" "+it.Description), q) {
					out = append(out, fmt.Sprintf("%-7s %-26s %s\n        install: %s", it.Kind, it.Name, it.Description, it.Install))
				}
			}
			sort.Strings(out)
			if len(out) == 0 {
				return "No marketplace matches.", nil
			}
			return strings.Join(out, "\n"), nil
		})

	AddTool(s, "marketplace_get",
		"Get details for a togo marketplace item (plugin/agent/skill) by name, including its install command.",
		func(_ context.Context, in struct {
			Name string `json:"name"`
			Kind string `json:"kind,omitempty"`
		}) (string, error) {
			for _, it := range marketplaceItems() {
				if it.Name == in.Name && mktMatchKind(it.Kind, in.Kind) {
					return fmt.Sprintf("%s  (%s)\n%s\ninstall: %s", it.Name, it.Kind, it.Description, it.Install), nil
				}
			}
			return "", fmt.Errorf("no marketplace item named %q", in.Name)
		})

	AddTool(s, "marketplace_install_hint",
		"Return the exact `togo install …` command to install a marketplace item (plugin/agent/skill).",
		func(_ context.Context, in struct {
			Name string `json:"name"`
			Kind string `json:"kind,omitempty"`
		}) (string, error) {
			for _, it := range marketplaceItems() {
				if it.Name == in.Name && mktMatchKind(it.Kind, in.Kind) {
					return it.Install, nil
				}
			}
			return "", fmt.Errorf("no marketplace item named %q", in.Name)
		})
}
