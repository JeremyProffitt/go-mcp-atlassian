package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/atlassian"
	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/atlassian/confluence"
	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/atlassian/jira"
	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/logging"
	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/mcp"
	"github.com/nicholasgriffintn/go-mcp-atlassian/pkg/tools"
)

// Version is set at build time
var Version = "dev"

func main() {
	// Load environment variables from ~/.mcp_env if it exists
	// This must happen before flag parsing so env vars are available for defaults
	logging.LoadEnvFile()

	// Define command-line flags
	logDir := flag.String("log-dir", "", "Log directory path")
	logLevel := flag.String("log-level", "info", "Log level (off, error, warn, info, access, debug)")
	httpMode := flag.Bool("http", false, "Run in HTTP mode (default: stdio)")
	port := flag.Int("port", 3000, "HTTP port (shorthand: -p)")
	flag.IntVar(port, "p", 3000, "HTTP port")
	host := flag.String("host", "127.0.0.1", "HTTP host (shorthand: -H)")
	flag.StringVar(host, "H", "127.0.0.1", "HTTP host")
	showVersion := flag.Bool("version", false, "Show version")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "go-mcp-atlassian - MCP server for Atlassian Jira and Confluence\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		fmt.Fprintf(os.Stderr, "  -log-dir string\n")
		fmt.Fprintf(os.Stderr, "        Log directory path\n")
		fmt.Fprintf(os.Stderr, "  -log-level string\n")
		fmt.Fprintf(os.Stderr, "        Log level: off, error, warn, info, access, debug (default \"info\")\n")
		fmt.Fprintf(os.Stderr, "  --http\n")
		fmt.Fprintf(os.Stderr, "        Run in HTTP mode (default: stdio)\n")
		fmt.Fprintf(os.Stderr, "  -p, --port int\n")
		fmt.Fprintf(os.Stderr, "        HTTP port (default 3000)\n")
		fmt.Fprintf(os.Stderr, "  -H, --host string\n")
		fmt.Fprintf(os.Stderr, "        HTTP host (default \"127.0.0.1\")\n")
		fmt.Fprintf(os.Stderr, "  --version\n")
		fmt.Fprintf(os.Stderr, "        Show version\n")
		fmt.Fprintf(os.Stderr, "  --help\n")
		fmt.Fprintf(os.Stderr, "        Show help\n")
		fmt.Fprintf(os.Stderr, "\nEnvironment Variables:\n")
		fmt.Fprintf(os.Stderr, "  JIRA_URL             Jira instance URL\n")
		fmt.Fprintf(os.Stderr, "  JIRA_USERNAME        Jira username (for Cloud)\n")
		fmt.Fprintf(os.Stderr, "  JIRA_API_TOKEN       Jira API token (for Cloud)\n")
		fmt.Fprintf(os.Stderr, "  JIRA_PERSONAL_TOKEN  Jira Personal Access Token (for Server/DC)\n")
		fmt.Fprintf(os.Stderr, "  CONFLUENCE_URL       Confluence instance URL\n")
		fmt.Fprintf(os.Stderr, "  CONFLUENCE_USERNAME  Confluence username (for Cloud)\n")
		fmt.Fprintf(os.Stderr, "  CONFLUENCE_API_TOKEN Confluence API token (for Cloud)\n")
		fmt.Fprintf(os.Stderr, "  CONFLUENCE_PERSONAL_TOKEN Confluence Personal Access Token (for Server/DC)\n")
		fmt.Fprintf(os.Stderr, "  READ_ONLY_MODE       Enable read-only mode (true/false)\n")
		fmt.Fprintf(os.Stderr, "  MCP_LOG_QUERIES      Log all queries to queries/ subfolder (true/false, default: false)\n")
	}

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("go-mcp-atlassian version %s\n", Version)
		os.Exit(0)
	}

	// Load configuration from environment
	jiraConfig := atlassian.NewJiraConfig()
	confluenceConfig := atlassian.NewConfluenceConfig()

	// Read READ_ONLY_MODE environment variable
	readOnlyMode := false
	if v := os.Getenv("READ_ONLY_MODE"); v != "" {
		readOnlyMode, _ = strconv.ParseBool(v)
	}

	// Read MCP_LOG_QUERIES environment variable (off by default)
	logQueries := false
	if v := os.Getenv("MCP_LOG_QUERIES"); v != "" {
		logQueries, _ = strconv.ParseBool(v)
	}

	// Validate that at least one client can be configured
	var jiraClient *jira.Client
	var confluenceClient *confluence.Client
	var err error

	if jiraConfig.URL != "" {
		jiraClient, err = jira.NewClient(jiraConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating Jira client: %v\n", err)
			os.Exit(1)
		}
	}

	if confluenceConfig.URL != "" {
		confluenceClient, err = confluence.NewClient(confluenceConfig)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating Confluence client: %v\n", err)
			os.Exit(1)
		}
	}

	// At least one client must be configured
	if jiraClient == nil && confluenceClient == nil {
		fmt.Fprintf(os.Stderr, "Error: At least one of JIRA_URL or CONFLUENCE_URL must be configured\n")
		os.Exit(1)
	}

	// Determine log directory and level sources for startup info
	logDirSource := logging.SourceDefault
	logLevelSource := logging.SourceDefault
	atlassianURLSource := logging.SourceEnvironment

	actualLogDir := *logDir
	addAppSubfolder := false
	if actualLogDir == "" {
		if envVal := os.Getenv("MCP_LOG_DIR"); envVal != "" {
			actualLogDir = envVal
			logDirSource = logging.SourceEnvironment
			addAppSubfolder = true // User specified a shared log directory
		} else {
			actualLogDir = logging.DefaultLogDir("go-mcp-atlassian")
		}
	} else {
		logDirSource = logging.SourceFlag
		addAppSubfolder = true // User specified a shared log directory
	}

	if *logLevel != "info" {
		logLevelSource = logging.SourceFlag
	}

	// Initialize logger
	loggerConfig := logging.Config{
		LogDir:          actualLogDir,
		AppName:         "go-mcp-atlassian",
		Level:           logging.ParseLevel(*logLevel),
		LogQueries:      logQueries,
		AddAppSubfolder: addAppSubfolder,
	}

	logger, err := logging.NewLogger(loggerConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Close()

	// Determine the Atlassian URL for logging
	atlassianURL := ""
	if jiraConfig.URL != "" {
		atlassianURL = jiraConfig.URL
	}
	if confluenceConfig.URL != "" {
		if atlassianURL != "" {
			atlassianURL += ", " + confluenceConfig.URL
		} else {
			atlassianURL = confluenceConfig.URL
		}
	}

	// Log startup information
	startupInfo := logging.GetStartupInfo(
		Version,
		logging.ConfigValue{Value: actualLogDir, Source: logDirSource},
		logging.ConfigValue{Value: *logLevel, Source: logLevelSource},
		logging.ConfigValue{Value: atlassianURL, Source: atlassianURLSource},
	)
	logger.LogStartup(startupInfo)

	// Defer shutdown logging
	defer func() {
		logger.LogShutdown("normal exit")
	}()

	// Create MCP server
	server := mcp.NewServer("go-mcp-atlassian", Version)

	// Set up telemetry callbacks
	server.SetToolCallCallback(func(name string, args map[string]interface{}, duration time.Duration, success bool) {
		logger.ToolCall(name, args, duration, success)
	})

	server.SetErrorCallback(func(err error, context string) {
		logger.Error("Error in %s: %v", context, err)
	})

	// Register Confluence tools if client exists
	if confluenceClient != nil {
		confluenceRegistry := tools.NewRegistry(confluenceClient, logger, readOnlyMode)
		confluenceRegistry.RegisterAll(server)
		logger.Info("Registered Confluence tools")
	}

	// Register Jira tools if client exists
	if jiraClient != nil {
		jiraRegistry := tools.NewJiraRegistry(jiraClient, logger, readOnlyMode)
		jiraRegistry.RegisterAll(server)
		logger.Info("Registered Jira tools")
	}

	// Set up graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Run server
	errChan := make(chan error, 1)

	go func() {
		if *httpMode {
			addr := fmt.Sprintf("%s:%d", *host, *port)
			logger.Info("Starting HTTP server on %s", addr)
			errChan <- server.RunHTTP(addr)
		} else {
			logger.Info("Starting stdio server")
			errChan <- server.Run()
		}
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		logger.Info("Received signal: %v", sig)
		logger.LogShutdown(fmt.Sprintf("received signal: %v", sig))
		os.Exit(0)
	case err := <-errChan:
		if err != nil {
			logger.Error("Server error: %v", err)
			logger.LogShutdown(fmt.Sprintf("server error: %v", err))
			os.Exit(1)
		}
	}
}
