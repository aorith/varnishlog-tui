package cmd

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/log"

	"github.com/aorith/varnishlog-tui/internal/ui"
	"github.com/aorith/varnishlog-tui/internal/ui/components/queryloader"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	Version   string = "dev"
	Commit    string = "none"
	BuildTime string = "unknown"
)

var (
	debugMode   *bool
	showVersion *bool
	queriesFile *string
)

func init() {
	debugMode = flag.Bool("debug", false, "enable debug logging")
	showVersion = flag.Bool("version", false, "show version information and exit")
	queriesFile = flag.String("file", "", "path to a YAML file containing queries")
}

func Execute() {
	flag.Parse()

	if *showVersion {
		fmt.Printf("varnishlog-tui %s\ncommit %s\nbuilt at %s\n", Version, Commit, BuildTime)
		os.Exit(0)
	}

	if *debugMode {
		f, err := tea.LogToFile("debug.log", "debug")
		if err != nil {
			log.Fatalf("Could not open log file: %v", err)
		}
		defer f.Close()
		log.SetOutput(f)
		log.SetTimeFormat(time.RFC3339)
		log.SetReportCaller(true)
		log.SetLevel(log.DebugLevel)
	}

	var configQueries *queryloader.QueriesConfig
	var err error
	if queriesFile != nil && *queriesFile != "" {
		configQueries, err = queryloader.LoadQueriesFromYaml(*queriesFile)
		if err != nil {
			log.Fatalf("Error loading queries from YAML: %v", err)
		}
	}

	ui.StartUI(configQueries)
}
