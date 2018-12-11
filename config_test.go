package main

import (
	"regexp"
	"testing"
	"time"

	"github.com/shadowbq/nomnomlog/papertrail"
	"github.com/shadowbq/nomnomlog/syslog"
	"github.com/stretchr/testify/assert"
)

func TestRawConfig(t *testing.T) {
	assert := assert.New(t)
	initConfigAndFlags()

	// pretend like some things were passed on the command line
	flags.Set("configfile", "test/config.yaml")
	flags.Set("tls", "true")

	c, err := NewConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(c.Destination.Host, "logs.papertrailapp.com")
	assert.Equal(c.Destination.Port, 514)
	assert.Equal(c.Destination.Protocol, "tls")
	// assert.Equal(c.IncludePatterns, []*regexp.Regexp{regexp.MustCompile("log only me"), regexp.MustCompile(`log o.{1,2} me`)})
	assert.Equal(c.ExcludePatterns, []*regexp.Regexp{regexp.MustCompile("don't log on me"), regexp.MustCompile(`do \w+ on me`)})
	assert.Equal(c.ExcludeFiles, []*regexp.Regexp{regexp.MustCompile(`\.DS_Store`)})
	assert.Equal(c.Files, []LogFile{
		{
			Path: "locallog.txt",
		},
		{
			Path: "/var/log/**/*.log",
		},
		{
			Tag:  "nginx",
			Path: "/var/log/nginx/nginx.log",
		},
		{
			Tag:  "apache",
			Path: "/var/log/httpd/access_log",
		},
	})
	assert.Equal(c.TcpMaxLineLength, 99991)
	assert.Equal(c.NewFileCheckInterval, 10*time.Second)
	assert.Equal(c.ConnectTimeout, 5*time.Second)
	assert.Equal(c.WriteTimeout, 30*time.Second)
	assert.Equal(c.TCP, false)
	assert.Equal(c.TLS, true)
	assert.Equal(c.LogLevels, "<root>=INFO")
	assert.Equal(c.PidFile, "/var/run/nomnomlog.pid")
	assert.Equal(c.DebugLogFile, "/dev/null")
	assert.Equal(c.NoDetach, false)
	sev, err := syslog.SeverityMap("notice")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(c.Severity, sev)
	fac, err := syslog.FacilityMap("user")
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(c.Facility, fac)
	assert.NotEqual(c.Hostname, "")
	assert.Equal(c.Poll, false)
	assert.Equal(c.RootCAs, papertrail.RootCA())
}

func TestNoConfigFile(t *testing.T) {
	assert := assert.New(t)
	initConfigAndFlags()

	flags.Set("dest-host", "localhost")
	flags.Set("dest-port", "999")

	c, err := NewConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}

	assert.NoError(c.Validate())
	assert.Equal("localhost", c.Destination.Host)
	assert.Equal(999, c.Destination.Port)
	assert.Equal("udp", c.Destination.Protocol)
}
