package core

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Options struct {
	Threads             *int
	ImageQuality        *int
	OutDir              *string
	SessionPath         *string
	SessionFileName     *string
	ReportFileName      *string
	CombineSessionPaths *[]string
	TemplatePath        *string
	Proxy               *string
	ChromePath          *string
	ChromeDevToolsURL   *string
	Resolution          *string
	Ports               *string
	ScanTimeout         *int
	HTTPTimeout         *int
	ScreenshotTimeout   *int
	Nmap                *bool
	UseRemoteChrome     *bool
	GenReport           *bool
	ClusterSimilar      *bool
	SaveBody            *bool
	Silent              *bool
	Debug               *bool
	Version             *bool
}

func ParseOptions() (Options, error) {
	var combineSessions []string
	combflag := flag.String("combine-sessions", "", "combine multi session files, sep by ','")
	combGlob := flag.String("combine-sessions-glob", "", "combine session file from file glob pattern")
	options := Options{
		Threads:             flag.Int("threads", 0, "Number of concurrent threads (default number of logical CPUs)"),
		ImageQuality:        flag.Int("quality", 70, "Screenshot image quality"),
		OutDir:              flag.String("out", ".", "Directory to write files to"),
		SessionPath:         flag.String("session", "", "Load Aquatone session file and generate HTML report"),
		SessionFileName:     flag.String("session-out-name", "aquatone_session.json", "Dump Aquatone session to json filename"),
		ReportFileName:      flag.String("report-out-name", "aquatone_report.html", "Generate HTML report filename"),
		CombineSessionPaths: &combineSessions,
		TemplatePath:        flag.String("template-path", "", "Path to HTML template to use for report"),
		Proxy:               flag.String("proxy", "", "Proxy to use for HTTP requests"),
		ChromePath:          flag.String("chrome-path", "", "Full path to the Chrome/Chromium executable to use. By default, aquatone will search for Chrome or Chromium"),
		ChromeDevToolsURL:   flag.String("chrome-dev-tools-url", "ws://localhost:3000", "When set UseRemoteChrome true, aquatone will use ChromeDevToolsURL to connect chrome dev tools"),
		Resolution:          flag.String("resolution", "1440,900", "screenshot resolution"),
		Ports:               flag.String("ports", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(MediumPortList)), ","), "[]"), "Ports to scan on hosts. Supported list aliases: small, medium, large, xlarge"),
		ScanTimeout:         flag.Int("scan-timeout", 100, "Timeout in miliseconds for port scans"),
		HTTPTimeout:         flag.Int("http-timeout", 3*1000, "Timeout in miliseconds for HTTP requests"),
		ScreenshotTimeout:   flag.Int("screenshot-timeout", 30*1000, "Timeout in miliseconds for screenshots"),
		Nmap:                flag.Bool("nmap", false, "Parse input as Nmap/Masscan XML"),
		UseRemoteChrome:     flag.Bool("use-remote-chrome", false, "Use remove chrome dev tools"),
		GenReport:           flag.Bool("report", true, "Generate report file"),
		SaveBody:            flag.Bool("save-body", true, "Save response bodies to files"),
		Silent:              flag.Bool("silent", false, "Suppress all output except for errors"),
		ClusterSimilar:      flag.Bool("similar", true, "Cluster page similarity"),
		Debug:               flag.Bool("debug", false, "Print debugging information"),
		Version:             flag.Bool("version", false, "Print current Aquatone version"),
	}
	flag.Parse()

	// 必须parse之后执行
	if *combflag == "" {
		combineSessions = nil
	} else {
		combineSessions = strings.Split(*combflag, ",")
		for i := range combineSessions {
			combineSessions[i] = strings.TrimSpace(combineSessions[i])
		}
	}

	if *combGlob != "" {
		files, err := filepath.Glob(*combGlob)
		if err != nil {
			fmt.Println("error glob session files:%s", err)
			os.Exit(1)
		} else {
			combineSessions = append(combineSessions, files...)
		}
	}

	return options, nil
}
