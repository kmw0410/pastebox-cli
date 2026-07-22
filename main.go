package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

var (
	version = "dev"
	commit  = "unknown"
)

const usageText = `Usage:
  pb [options] [file|-]
  pb show [--password] <code|url>
  pb clone [options] <code|url>
  pb delete <code|delete-url>
  pb config show
  pb config set server <URL>
  pb config validate
  pb update
  pb version

Upload options:
  --permanent       keep the paste permanently
  --once            delete after the first successful view
  --expires VALUE   expire after a duration such as 30m, 12h, or 7d
  --password        prompt for password protection
  --code VALUE      use a custom paste code
  --label VALUE     attach a paste label
  --quiet           print only the public URL
  --json            print the JSON response
`

const showUsageText = `Usage:
  pb show [--password] <code|url>

Show options:
  --password  prompt for the paste password
`

const cloneUsageText = `Usage:
  pb clone [options] <code|url>

Clone options:
  --source-password           prompt for the source paste password
  --permanent                 keep the cloned paste permanently
  --once                      delete after the first successful view
  --expires VALUE             expire after a duration such as 30m, 12h, or 7d
  --password                  prompt for password protection for the clone
  --code VALUE                use a custom code for the clone
  --quiet                     print only the cloned paste URL
  --json                      print the JSON response
`

const deleteUsageText = `Usage:
  pb delete <code|delete-url>
`

const configUsageText = `Usage:
  pb config show
  pb config set server <URL>
  pb config validate
`

const updateUsageText = `Usage:
  pb update
`

const (
	connectTimeout        = 30 * time.Second
	tlsHandshakeTimeout   = 10 * time.Second
	responseHeaderTimeout = 30 * time.Second
)

type application struct {
	stdin         io.Reader
	stdout        io.Writer
	stderr        io.Writer
	httpClient    *http.Client
	configPath    string
	stdinTTY      bool
	ctx           context.Context
	osReleasePath string
	goarch        string
	releaseAPIURL string
	runCommand    func(context.Context, string, ...string) error
	lookPath      func(string) (string, error)
	effectiveUID  func() int
	readPassword  func(string) (string, error)
}

func main() {
	stdinTTY := false
	if info, err := os.Stdin.Stat(); err == nil {
		stdinTTY = info.Mode()&os.ModeCharDevice != 0
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	app := application{
		stdin:      os.Stdin,
		stdout:     os.Stdout,
		stderr:     os.Stderr,
		httpClient: newHTTPClient(),
		configPath: defaultConfigPath(),
		stdinTTY:   stdinTTY,
		ctx:        ctx,
		readPassword: func(prompt string) (string, error) {
			return readTerminalPassword(os.Stderr, prompt)
		},
	}
	os.Exit(app.run(os.Args[1:]))
}

func newHTTPClient() *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&net.Dialer{
		Timeout:   connectTimeout,
		KeepAlive: 30 * time.Second,
	}).DialContext
	transport.TLSHandshakeTimeout = tlsHandshakeTimeout
	transport.ResponseHeaderTimeout = responseHeaderTimeout
	return &http.Client{Transport: transport}
}

func (a application) requestContext() context.Context {
	if a.ctx != nil {
		return a.ctx
	}
	return context.Background()
}

func (a application) run(args []string) int {
	if len(args) > 0 {
		switch args[0] {
		case "help", "--help", "-h":
			fmt.Fprint(a.stdout, usageText)
			return 0
		case "version", "--version":
			fmt.Fprintf(a.stdout, "pb %s (commit %s)\n", version, commit)
			return 0
		case "config":
			return a.runConfig(args[1:])
		case "show":
			return a.runShow(args[1:])
		case "clone":
			return a.runClone(args[1:])
		case "delete":
			return a.runDelete(args[1:])
		case "update":
			return a.runUpdate(args[1:])
		}
	}
	return a.runUpload(args)
}

func (a application) runConfig(args []string) int {
	switch {
	case len(args) == 1 && (args[0] == "--help" || args[0] == "-h"):
		fmt.Fprint(a.stdout, configUsageText)
		return 0

	case len(args) == 1 && args[0] == "show":
		cfg, err := loadConfig(a.configPath)
		if err != nil {
			fmt.Fprintln(a.stderr, err)
			return 2
		}
		fmt.Fprintf(a.stdout, "config: %s\nserver_url: %s\n", a.configPath, cfg.ServerURL)
		return 0

	case len(args) == 1 && args[0] == "validate":
		cfg, err := loadConfig(a.configPath)
		if err != nil {
			fmt.Fprintln(a.stderr, err)
			return 2
		}
		fmt.Fprintf(a.stdout, "valid config: %s\nserver_url: %s\n", a.configPath, cfg.ServerURL)
		return 0

	case len(args) == 3 && args[0] == "set" && args[1] == "server":
		serverURL, err := validateServerURL(a.configPath, args[2])
		if err != nil {
			fmt.Fprintln(a.stderr, err)
			return 2
		}
		if err := saveConfig(a.configPath, config{ServerURL: serverURL}); err != nil {
			fmt.Fprintln(a.stderr, err)
			return 2
		}
		fmt.Fprintf(a.stdout, "updated config: %s\nserver_url: %s\n", a.configPath, serverURL)
		return 0

	default:
		fmt.Fprint(a.stderr, configUsageText)
		return 2
	}
}

func (a application) runUpload(args []string) int {
	flags := flag.NewFlagSet("upload", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var opts uploadOptions
	flags.BoolVar(&opts.permanent, "permanent", false, "")
	flags.BoolVar(&opts.once, "once", false, "")
	flags.StringVar(&opts.expires, "expires", "", "")
	promptPassword := flags.Bool("password", false, "")
	flags.StringVar(&opts.code, "code", "", "")
	flags.StringVar(&opts.label, "label", "", "")
	flags.BoolVar(&opts.quiet, "quiet", false, "")
	flags.BoolVar(&opts.jsonOutput, "json", false, "")
	if err := flags.Parse(args); err != nil {
		fmt.Fprintf(a.stderr, "invalid arguments: %v\n", err)
		return 2
	}
	if err := opts.validate(); err != nil {
		fmt.Fprintf(a.stderr, "invalid arguments: %v\n", err)
		return 2
	}
	if flags.NArg() > 1 {
		fmt.Fprintln(a.stderr, "invalid arguments: provide one file or use stdin")
		return 2
	}
	if len(args) == 0 && a.stdinTTY {
		created, err := initializeConfig(a.configPath)
		if err != nil {
			fmt.Fprintln(a.stderr, err)
			return 2
		}
		if created {
			fmt.Fprintf(a.stdout, "created config: %s\nRun pb config set server <URL> before using pb.\n", a.configPath)
			return 0
		}
	}

	filename := ""
	reader := a.stdin
	var closeInput io.Closer
	if flags.NArg() == 1 && flags.Arg(0) != "-" {
		file, err := os.Open(flags.Arg(0))
		if err != nil {
			fmt.Fprintf(a.stderr, "cannot open %q: %v\n", flags.Arg(0), err)
			return 2
		}
		reader = file
		closeInput = file
		filename = file.Name()
	} else if a.stdinTTY {
		fmt.Fprintln(a.stderr, "no input: provide a file or pipe text to pb")
		return 2
	}
	if closeInput != nil {
		defer closeInput.Close()
	}

	cfg, err := loadConfig(a.configPath)
	if err != nil {
		fmt.Fprintln(a.stderr, err)
		return 2
	}
	if *promptPassword {
		opts.newPassword, err = a.promptNewPassword()
		if err != nil {
			fmt.Fprintf(a.stderr, "cannot read password: %v\n", err)
			return 2
		}
	}
	result, raw, err := upload(a.requestContext(), a.httpClient, cfg, reader, filename, opts)
	if err != nil {
		fmt.Fprintln(a.stderr, err)
		return 1
	}
	if err := writeUploadOutput(a.stdout, result, raw, opts); err != nil {
		fmt.Fprintf(a.stderr, "write output: %v\n", err)
		return 1
	}
	return 0
}

func (a application) runShow(args []string) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(a.stdout, showUsageText)
		return 0
	}

	flags := flag.NewFlagSet("show", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	promptPassword := flags.Bool("password", false, "")
	if err := flags.Parse(args); err != nil {
		fmt.Fprintf(a.stderr, "invalid arguments: %v\n", err)
		return 2
	}
	if flags.NArg() != 1 || strings.TrimSpace(flags.Arg(0)) == "" {
		fmt.Fprint(a.stderr, showUsageText)
		return 2
	}

	cfg, err := loadConfig(a.configPath)
	if err != nil {
		fmt.Fprintln(a.stderr, err)
		return 2
	}
	password := ""
	if *promptPassword {
		password, err = a.promptExistingPassword("Paste password: ")
		if err != nil {
			fmt.Fprintf(a.stderr, "cannot read password: %v\n", err)
			return 2
		}
	}
	if err := getPaste(a.requestContext(), a.httpClient, cfg, flags.Arg(0), password, a.stdout); err != nil {
		fmt.Fprintln(a.stderr, err)
		return 1
	}
	return 0
}

func (a application) runClone(args []string) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(a.stdout, cloneUsageText)
		return 0
	}

	flags := flag.NewFlagSet("clone", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	var opts uploadOptions
	promptSourcePassword := flags.Bool("source-password", false, "")
	flags.BoolVar(&opts.permanent, "permanent", false, "")
	flags.BoolVar(&opts.once, "once", false, "")
	flags.StringVar(&opts.expires, "expires", "", "")
	promptPassword := flags.Bool("password", false, "")
	flags.StringVar(&opts.code, "code", "", "")
	flags.BoolVar(&opts.quiet, "quiet", false, "")
	flags.BoolVar(&opts.jsonOutput, "json", false, "")
	if err := flags.Parse(args); err != nil {
		fmt.Fprintf(a.stderr, "invalid arguments: %v\n", err)
		return 2
	}
	if err := opts.validate(); err != nil {
		fmt.Fprintf(a.stderr, "invalid arguments: %v\n", err)
		return 2
	}
	if flags.NArg() != 1 || strings.TrimSpace(flags.Arg(0)) == "" {
		fmt.Fprint(a.stderr, cloneUsageText)
		return 2
	}

	cfg, err := loadConfig(a.configPath)
	if err != nil {
		fmt.Fprintln(a.stderr, err)
		return 2
	}
	sourcePassword := ""
	if *promptSourcePassword {
		sourcePassword, err = a.promptExistingPassword("Source paste password: ")
		if err != nil {
			fmt.Fprintf(a.stderr, "cannot read source password: %v\n", err)
			return 2
		}
	}
	if *promptPassword {
		opts.newPassword, err = a.promptNewPassword()
		if err != nil {
			fmt.Fprintf(a.stderr, "cannot read new password: %v\n", err)
			return 2
		}
	}
	result, raw, err := clonePaste(a.requestContext(), a.httpClient, cfg, flags.Arg(0), sourcePassword, opts)
	if err != nil {
		fmt.Fprintln(a.stderr, err)
		return 1
	}
	if err := writeUploadOutput(a.stdout, result, raw, opts); err != nil {
		fmt.Fprintf(a.stderr, "write output: %v\n", err)
		return 1
	}
	return 0
}
