package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/alexflint/go-arg"
)

type SearchConfig struct {
	Fields fieldList `arg:"-F,env:CLIBANA_FIELDS" help:"List of fields to output. Optionally, the field output color can be set" placeholder:"FIELD_NAME[:COLOR],..."`
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
	URL      string          `arg:"-u,required,env:CLIBANA_URL" help:"http[s]://host[:port] or aws://cluster-name"`
	Index    string          `arg:"-i,required,env:CLIBANA_INDEX" help:"Index pattern"`
	AuthType string          `arg:"-a,env:CLIBANA_AUTH" help:"Authentication type: aws or basic"`
	Username string          `arg:"-U,env:CLIBANA_USER" help:"Username for basic authentication"`
	Password string          `arg:"-p,env:CLIBANA_PASSWORD" help:"Password for basic authentication"`
	Debug    bool            `arg:"-d,env:CLIBANA_DEBUG" help:"Enable debug output"`
	Search   *SearchConfig   `arg:"subcommand:search" help:"Search indices matching the index pattern"`
	Mappings *MappingsConfig `arg:"subcommand:mappings" help:"Show field mappings in the indices matching the index pattern"`
	Indices  *IndicesConfig  `arg:"subcommand:indices" help:"List indices matching the index pattern"`
}

func NewClibanaConfig() ClibanaConfig {
	var clibanaConfig ClibanaConfig

	arg.MustParse(&clibanaConfig)

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
			"Supported color names:",
			"black, red, green, yellow, blue, magenta, cyan, white",
			"hiblack, hired, higreen, hiyellow, hiblue, himagenta, hicyan, hiwhite",
			"",
			"For more information, see https://github.com/ivoronin/clibana",
		}, "\n")
}

func (ClibanaConfig) Version() string {
	return fmt.Sprintf("clibana %s (commit: %s, build date: %s)", version, commit, date)
}

type fieldListItem struct {
	Name  string
	Color string
}

type fieldList []fieldListItem

func (c *fieldList) UnmarshalText(text []byte) error { //nolint:unparam
	parts := strings.Split(string(text), ",")
	for _, part := range parts {
		name, color, _ := strings.Cut(part, ":")
		*c = append(*c, fieldListItem{Name: name, Color: color})
	}

	return nil
}
