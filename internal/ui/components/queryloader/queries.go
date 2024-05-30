package queryloader

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/aorith/varnishlog-tui/internal/ui/state"
	"github.com/aorith/varnishlog-tui/internal/util"
	"gopkg.in/yaml.v3"
)

// Query represents a single varnishlog query
type Query struct {
	Name   string `yaml:"name"`
	Script string `yaml:"script"`
}

// QueriesConfig represents a collection of queries
type QueriesConfig struct {
	Queries []Query `yaml:"queries"`
}

// FilterValue satisfaces list.Item interface
func (q Query) FilterValue() string {
	return q.Name
}

// Title satisfaces list.DefaultItem interface
func (q Query) Title() string {
	return q.Name
}

// Description satisfaces list.DefaultItem interface
func (q Query) Description() string {
	return strings.ReplaceAll(util.ParseVarnishlogArgs(q.Script), "\n", " ")
}

func (q Query) newQueryEditorData() state.NewQueryEditorScriptMsg {
	return state.NewQueryEditorScriptMsg(q.Script)
}

// QueryToYamlLines converts a Query into a YAML string split into lines.
// It returns a slice of strings, each representing a line of the YAML output.
// If an error occurs during marshalling, the returned slice contains an error message.
func QueryToYamlLines(q Query) []string {
	qc := QueriesConfig{Queries: []Query{q}}
	yamlData, err := yaml.Marshal(&qc)
	if err != nil {
		return []string{fmt.Sprintf("query: marshal error: %s", err.Error())}
	}
	return strings.Split(string(yamlData), "\n")
}

func LoadQueriesFromYaml(filename string) (*QueriesConfig, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open YAML file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("could not read YAML file: %w", err)
	}

	var config QueriesConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal YAML: %w", err)
	}

	return &config, nil
}
