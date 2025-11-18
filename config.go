package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/alexflint/go-arg"
)

type SearchConfig struct {
	Fields fieldList `arg:"-F,env:CLIBANA_FIELDS" help:"Comma-separated list of fields to output" placeholder:"FIELD_NAME,..."`
	Follow bool      `arg:"-f" help:"Enable live tailing of logs"`
	Query  string    `arg:"positional" default:"*" help:"Query string"`
	Start  string    `arg:"-s" default:"now-5m" help:"Start time"`
	End    string    `arg:"-e" default:"now" help:"End time"`
}

type MappingsConfig struct {
	Quiet bool `arg:"-q" help:"Do not show headers"`
}

type IndicesConfig struct {
	Quiet bool `arg:"-q" help:"Do not show headers"`
}

type ClibanaConfig struct {
	URL          string          `arg:"-u,required,env:CLIBANA_URL" help:"http[s]://host[:port] or aws://cluster-name"`
	Index        string          `arg:"-i,required,env:CLIBANA_INDEX" help:"Index pattern"`
	AuthType     string          `arg:"-a,env:CLIBANA_AUTH" help:"Authentication type: aws or basic"`
	Username     string          `arg:"-U,env:CLIBANA_USER" help:"Username for basic authentication"`
	PasswordFile string          `arg:"--password-file" help:"Path to file containing password for basic authentication"`
	Password     string          `arg:"-"`
	Debug        bool            `arg:"-d,env:CLIBANA_DEBUG" help:"Enable debug output"`
	Search       *SearchConfig   `arg:"subcommand:search" help:"Search indices matching the index pattern"`
	Mappings     *MappingsConfig `arg:"subcommand:mappings" help:"Show field mappings in the indices matching the index pattern"`
	Indices      *IndicesConfig  `arg:"subcommand:indices" help:"List indices matching the index pattern"`
}

func NewClibanaConfig() ClibanaConfig {
	var clibanaConfig ClibanaConfig

	arg.MustParse(&clibanaConfig)

	// Читаем пароль из файла если указан
	if clibanaConfig.PasswordFile != "" {
		password, err := os.ReadFile(clibanaConfig.PasswordFile)
		if err != nil {
			FatalError(fmt.Errorf("failed to read password file: %w", err))
		}
		clibanaConfig.Password = strings.TrimSpace(string(password))
	}

	// Парсим URL для определения схемы
	parsedURL, err := url.Parse(clibanaConfig.URL)
	if err != nil {
		FatalError(fmt.Errorf("failed to parse URL: %w", err))
	}

	// Резолвим AWS URL в настоящий HTTPS endpoint
	if parsedURL.Scheme == "aws" {
		clibanaConfig.URL = resolveAWSDomainEndpoint(parsedURL.Host)
	}

	// Автоматически выбираем тип авторизации если он не задан явно
	if clibanaConfig.AuthType == "" {
		// Используем AWS auth только если была схема aws:// и не заданы username/password
		if parsedURL.Scheme == "aws" && clibanaConfig.Username == "" && clibanaConfig.Password == "" {
			clibanaConfig.AuthType = AuthTypeAWS
		} else {
			clibanaConfig.AuthType = AuthTypeBasic
		}
	}

	return clibanaConfig
}

func (ClibanaConfig) Description() string {
	return "Clibana - OpenSearch log tailer"
}

func (ClibanaConfig) Epilogue() string {
	return strings.Join(
		[]string{
			"Query Examples:",
			"  *                               Match all logs",
			"  error                           Search in all fields",
			"  level:ERROR                     Field-specific search",
			"  pod_name:nginx*                 Wildcard search",
			`  message:"out of memory"         Phrase search`,
			"  level:ERROR AND pod:web-*       Boolean AND",
			"  level:ERROR AND NOT pod:test-*  Boolean AND NOT",
			"",
			"Time Format Examples:",
			"  -s now-5m                    Last 5 minutes (default)",
			"  -s now-2h -e now-1h          1-2 hours ago",
			"  -s 2024-01-15T10:00:00Z      Absolute timestamp",
			"",
			"For more information, see https://github.com/ivoronin/clibana",
		}, "\n")
}

func (ClibanaConfig) Version() string {
	return fmt.Sprintf("clibana %s (commit: %s, build date: %s)", version, commit, date)
}

type fieldListItem struct {
	Name string
}

type fieldList []fieldListItem

func (c *fieldList) UnmarshalText(text []byte) error { //nolint:unparam
	parts := strings.Split(string(text), ",")
	for _, part := range parts {
		*c = append(*c, fieldListItem{Name: strings.TrimSpace(part)})
	}

	return nil
}
