package main

import (
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/shadowbq/nomnomlog/papertrail"
	"github.com/shadowbq/nomnomlog/syslog"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	config *viper.Viper
	flags  *pflag.FlagSet

	Version string

	ErrUsage = errors.New("usage")
)

const (
	envPrefix         = "nomnomlog"
	defaultConfigFile = "/etc/nomnomlog-config.yml"
)

// The global Config object for nomnomlog server. "mapstructure" tags
// signify the config file key names.
type Config struct {
	ConnectTimeout       time.Duration    `mapstructure:"connect_timeout"`
	KeepReconnecting     bool             `mapstructure:"keep_reconnecting"`
	WriteTimeout         time.Duration    `mapstructure:"write_timeout"`
	NewFileCheckInterval time.Duration    `mapstructure:"new_file_check_interval"`
	ExcludeFiles         []*regexp.Regexp `mapstructure:"exclude_files"`
	ExcludePatterns      []*regexp.Regexp `mapstructure:"exclude_patterns"`
	IncludePatterns      []*regexp.Regexp `mapstructure:"include_patterns"`
	LogLevels            string           `mapstructure:"log_levels"`
	DebugLogFile         string           `mapstructure:"debug_log_file"`
	PidFile              string           `mapstructure:"pid_file"`
	TcpMaxLineLength     int              `mapstructure:"tcp_max_line_length"`
	NoDetach             bool             `mapstructure:"no_detach"`
	TCP                  bool             `mapstructure:"tcp"`
	TLS                  bool             `mapstructure:"tls"`
	TruncateHostname     bool             `mapstructure:"truncate_hostname"`
	Files                []LogFile
	Hostname             string
	Severity             syslog.Severity
	Facility             syslog.Facility
	Poll                 bool
	Destination          struct {
		Host     string
		Port     int
		Protocol string
	}
	RootCAs *x509.CertPool
}

type LogFile struct {
	Path string
	Tag  string
}

func init() {
	initConfigAndFlags()
}

func initConfigAndFlags() {
	flags = pflag.NewFlagSet(envPrefix, pflag.ExitOnError)

	config = viper.New()
	config.SetEnvPrefix(envPrefix)

	// set defaults for configuration values that aren't provided by flags here:
	config.SetDefault("destination.protocol", "udp")
	config.SetDefault("tcp_max_line_length", 99990)
	config.SetDefault("debug_log_file", "/dev/null")
	config.SetDefault("connect_timeout", 30*time.Second)
	config.SetDefault("keep_reconnecting", false)
	config.SetDefault("write_timeout", 30*time.Second)

	// flag-only "configuration" values (help and version)
	flags.BoolP("help", "h", false, "Display this help message")
	flags.BoolP("version", "V", false, "Display version and exit")

	// set available commandline flags here:
	flags.StringP("configfile", "c", defaultConfigFile, "Path to config")
	config.BindPFlag("config_file", flags.Lookup("configfile"))

	flags.StringP("dest-host", "d", "", "Destination syslog hostname or IP")
	config.BindPFlag("destination.host", flags.Lookup("dest-host"))

	flags.IntP("dest-port", "p", 514, "Destination syslog port")
	config.BindPFlag("destination.port", flags.Lookup("dest-port"))

	flags.StringP("facility", "f", "user", "Facility")
	config.BindPFlag("facility", flags.Lookup("facility"))

	hostname, _ := os.Hostname()
	flags.String("hostname", hostname, "Local hostname to send from")
	config.BindPFlag("hostname", flags.Lookup("hostname"))

	flags.Bool("truncate-hostname", false, "Local truncate-hostname to send from")
	config.BindPFlag("truncate_hostname", flags.Lookup("truncate-hostname"))

	flags.String("pid-file", "", "Location of the PID file")
	config.BindPFlag("pid_file", flags.Lookup("pid-file"))

	flags.StringP("severity", "s", "notice", "Severity")
	config.BindPFlag("severity", flags.Lookup("severity"))

	flags.Bool("tcp", false, "Connect via TCP (no TLS)")
	config.BindPFlag("tcp", flags.Lookup("tcp"))

	flags.Bool("tls", false, "Connect via TCP with TLS")
	config.BindPFlag("tls", flags.Lookup("tls"))

	flags.Bool("poll", false, "Detect changes by polling instead of inotify")
	config.BindPFlag("poll", flags.Lookup("poll"))

	flags.Int("new-file-check-interval", 10, "How often to check for new files (seconds)")
	config.BindPFlag("new_file_check_interval", flags.Lookup("new-file-check-interval"))

	flags.String("debug-log-cfg", "", "The debug log file; overridden by -D/--no-detach")
	config.BindPFlag("debug_log_file", flags.Lookup("debug-log-cfg"))

	flags.String("log", "<root>=INFO", "Set loggo config, like: --log=\"<root>=DEBUG\"")
	config.BindPFlag("log_levels", flags.Lookup("log"))

	// only present this flag to systems that can daemonize
	if CanDaemonize {
		flags.BoolP("no-detach", "D", false, "Do not daemonize and detach from the terminal; overrides --debug-log-cfg")
		config.BindPFlag("no_detach", flags.Lookup("no-detach"))
	}

	// deprecated flags
	flags.Bool("no-eventmachine-tail", false, "No action, provided for backwards compatibility")
	flags.Bool("eventmachine-tail", false, "No action, provided for backwards compatibility")

	// bind env vars to config automatically
	config.AutomaticEnv()

}

// Read in configuration from environment, flags, and specified or default config file.
func NewConfigFromEnv() (*Config, error) {
	if err := flags.Parse(os.Args[1:]); err != nil {
		return nil, err
	}

	if h, _ := flags.GetBool("help"); h {
		usage()
		return nil, ErrUsage
	}

	if v, _ := flags.GetBool("version"); v {
		version()
		return nil, ErrUsage
	}

	c := &Config{}

	// read in config file if it's there
	configFile := config.GetString("config_file")
	config.SetConfigFile(configFile)
	if err := config.ReadInConfig(); err != nil && configFile != defaultConfigFile {
		return nil, err
	}

	// override daemonize setting for platforms that don't support it
	if !CanDaemonize {
		config.Set("no_daemonize", true)
	}

	// unmarshal environment config into our Config object here
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           c,
		WeaklyTypedInput: true,
		DecodeHook:       decodeHook,
	})
	if err != nil {
		return nil, err
	}

	if err = decoder.Decode(config.AllSettings()); err != nil {
		return nil, err
	}

	// explicitly set destination fields since they are nested
	c.Destination.Host = config.GetString("destination.host")
	c.Destination.Port = config.GetInt("destination.port")
	c.Destination.Protocol = config.GetString("destination.protocol")

	// explicitly set destination protocol if we've asked for tcp or tls
	if c.TLS {
		c.Destination.Protocol = "tls"
	}
	if c.TCP {
		c.Destination.Protocol = "tcp"
	}

	// truncate hostname if requested in configuration
	if c.TruncateHostname {
		i := strings.Index(c.Hostname, ".")
		if i != -1 {
			c.Hostname = c.Hostname[:i]
		}
	}

	// https://github.com/shadowbq/nomnomlog/issues/17
	// Make this more extensible
	// add the papertrail root CA if necessary
	if c.Destination.Protocol == "tls" && c.Destination.Host == "logs.papertrailapp.com" {
		c.RootCAs = papertrail.RootCA()
	}

	// figure out where to create a pidfile if none was configured
	if c.PidFile == "" {
		c.PidFile = getPidFile()
	}

	// collect any extra args passed on the command line and add them to our file list
	for _, file := range flags.Args() {
		files, err := decodeLogFiles([]interface{}{file})
		if err != nil {
			return nil, err
		}

		c.Files = append(c.Files, files...)
	}

	return c, nil
}

func (c *Config) Validate() error {

	// Symptom of bad config, but not the correct error in all cases.
	// GH Issue #4
	if c.Destination.Host == "" {
		return fmt.Errorf("No destination hostname specified")
	}

	if c.NewFileCheckInterval < 1*time.Second {
		return fmt.Errorf("new_file_check_interval is too small, try setting >= 1")
	}

	return nil
}

func decodeDuration(f interface{}) (time.Duration, error) {
	var (
		i   int
		err error
	)

	switch val := f.(type) {
	case string:
		i, err = strconv.Atoi(val)
		if err != nil {
			return 0, err
		}

	case int:
		i = val

	case time.Duration:
		return val, nil

	default:
		return 0, fmt.Errorf("Invalid duration: %#v", val)
	}

	return time.Duration(i) * time.Second, nil
}

func decodeRegexps(f interface{}) ([]*regexp.Regexp, error) {
	rs, ok := f.([]interface{})
	if !ok {
		return nil, fmt.Errorf("Invalid input type for regular expression %#v", f)
	}

	exps := make([]*regexp.Regexp, len(rs))
	for i, r := range rs {
		str, ok := r.(string)
		if !ok {
			return nil, fmt.Errorf("Invalid input type for regular expression %#v", r)
		}

		exp, err := regexp.Compile(str)
		if err != nil {
			return nil, err
		}

		exps[i] = exp
	}

	return exps, nil
}

func decodeLogFiles(f interface{}) ([]LogFile, error) {
	var (
		files []LogFile
	)

	vals, ok := f.([]interface{})
	if !ok {
		return files, fmt.Errorf("Invalid input type for files: %#v", f)
	}

	for _, v := range vals {
		switch val := v.(type) {
		case string:
			lf := strings.Split(val, "=")
			switch len(lf) {
			case 2:
				files = append(files, LogFile{Tag: lf[0], Path: lf[1]})
			case 1:
				files = append(files, LogFile{Path: val})
			default:
				return files, fmt.Errorf("Invalid log file name %s", val)
			}

		case map[interface{}]interface{}:
			var (
				tag  string
				path string
			)

			tag, _ = val["tag"].(string)
			path, _ = val["path"].(string)

			if path == "" {
				return files, fmt.Errorf("Invalid log file %#v", val)
			}

			files = append(files, LogFile{Tag: tag, Path: path})

		default:
			panic(vals)
		}
	}

	return files, nil
}

func decodeFacility(p interface{}) (interface{}, error) {
	ps, ok := p.(string)
	if !ok {
		return nil, fmt.Errorf("Invalid SYSLOG Facility String: %#v", p)
	}

	pri, err := syslog.FacilityMap(ps)
	if err == nil {
		return pri, nil
	}

	return nil, fmt.Errorf("%s: %s", err.Error(), ps)
}

func decodeSeverity(p interface{}) (interface{}, error) {
	ps, ok := p.(string)
	if !ok {
		return nil, fmt.Errorf("Invalid SYSLOG Severity String: %#v", p)
	}

	pri, err := syslog.SeverityMap(ps)
	if err == nil {
		return pri, nil
	}

	return nil, fmt.Errorf("%s: %s", err.Error(), ps)
}

func decodeHook(from reflect.Type, to reflect.Type, data interface{}) (interface{}, error) {
	switch to {
	case reflect.TypeOf([]LogFile{}):
		return decodeLogFiles(data)
	case reflect.TypeOf([]*regexp.Regexp{}):
		return decodeRegexps(data)
	case reflect.TypeOf(syslog.Facility(0)):
		return decodeFacility(data)
	case reflect.TypeOf(syslog.Severity(0)):
		return decodeSeverity(data)
	case reflect.TypeOf(time.Duration(0)):
		return decodeDuration(data)
	}

	return data, nil
}

func getPidFile() string {
	pidFiles := []string{
		"/var/run/nomnomlog.pid",
		os.Getenv("HOME") + "/run/nomnomlog.pid",
		os.Getenv("HOME") + "/tmp/nomnomlog.pid",
		os.Getenv("HOME") + "/nomnomlog.pid",
		os.TempDir() + "/nomnomlog.pid",
		os.Getenv("TMPDIR") + "/nomnomlog.pid",
	}

	for _, f := range pidFiles {
		dir := filepath.Dir(f)
		dirStat, err := os.Stat(dir)
		if err != nil || dirStat == nil || !dirStat.IsDir() {
			continue
		}

		fd, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE, 0644)
		if err != nil {
			continue
		}
		fd.Close()

		return f
	}

	return "/tmp/nomnomlog.pid"
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s %s:\n", envPrefix, Version)
	flags.PrintDefaults()
}

func version() {
	fmt.Fprintf(os.Stderr, "%s %s\n", envPrefix, Version)
}

func regexpArrToStringArr(regpattern []*regexp.Regexp) []string {
	var rsc []string
	for _, element := range regpattern {
		rsc = append(rsc, element.String())
	}
	return rsc
}

func sizeOfCertPool(pool *x509.CertPool) (mysize int) {
	// This should not be needed. LAW OF DEMETER, check nils
	//
	// defer func() {
	//	// recover from panic if one occurs.. len(pool.DEMETER) panics because of no default value of pool.
	//	if recover() != nil {
	//		mysize = 0
	//	}
	// }()
	if pool != nil {
		mysize = len(pool.Subjects())
	}

	return mysize
}

func dumpConfig(c *Config) {

	fmt.Fprintf(os.Stderr, "Running Configuration: \n")
	type ConfigAlias Config
	jsonConfig, err := json.MarshalIndent(&struct {
		ExcludeFiles    []string
		ExcludePatterns []string
		IncludePatterns []string
		RootCAs         int
		*ConfigAlias
	}{
		ExcludeFiles:    regexpArrToStringArr(c.ExcludeFiles),
		ExcludePatterns: regexpArrToStringArr(c.ExcludePatterns),
		IncludePatterns: regexpArrToStringArr(c.IncludePatterns),
		RootCAs:         sizeOfCertPool(c.RootCAs),
		ConfigAlias:     (*ConfigAlias)(c),
	}, "", "\t")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(jsonConfig))

}
