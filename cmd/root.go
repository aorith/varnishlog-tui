package cmd

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/charmbracelet/log"

	"github.com/aorith/varnishlog-tui/internal/ui"
	"github.com/aorith/varnishlog-tui/internal/ui/components/queryloader"
	tea "github.com/charmbracelet/bubbletea"
)

var (
	Version   string
	Commit    string
	BuildTime string
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

	if info, ok := debug.ReadBuildInfo(); ok {
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				Commit = setting.Value
			case "vcs.time":
				BuildTime = setting.Value
			case "vcs.modified":
				if setting.Value == "true" {
					Commit += "-modified"
				}
			}
		}
		Version = info.Main.Version
	}
}

func Execute() {

	flag.Parse()

	if *showVersion {
		fmt.Printf("Version: %s\n", Version)
		fmt.Printf("Commit: %s\n", Commit)
		fmt.Printf("Build Time: %s\n", BuildTime)
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
