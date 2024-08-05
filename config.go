package main

import (
	"fmt"
	"strings"

	"github.com/alexflint/go-arg"
)

type SearchConfig struct {
	Fields fieldList `arg:"-o,env:CLIBANA_FIELDS" help:"Comma-separated list of fields to output"`
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
	Host     string          `arg:"-H,required,env:CLIBANA_HOST" help:"http[s]://host[:port] or aws://cluster-name"`
	Index    string          `arg:"-i,required,env:CLIBANA_INDEX" help:"Index pattern"`
	AuthType string          `arg:"-a,env:CLIBANA_AUTH" default:"basic" help:"Authentication type: aws or basic"`
	Username string          `arg:"-u,env:CLIBANA_USER" help:"Username for basic authentication"`
	Password string          `arg:"-p,env:CLIBANA_PASSWORD" help:"Password for basic authentication"`
	Debug    bool            `arg:"-d,env:CLIBANA_DEBUG" help:"Enable debug output"`
	Search   *SearchConfig   `arg:"subcommand:search" help:"Search indices matching the index pattern"`
	Mappings *MappingsConfig `arg:"subcommand:mappings" help:"Show field mappings in the indices matching the index pattern"`
	Indices  *IndicesConfig  `arg:"subcommand:indices" help:"List indices matching the index pattern"`
}

func NewClibanaConfig() ClibanaConfig {
	var clibanaConfig ClibanaConfig

	arg.MustParse(&clibanaConfig)

	return clibanaConfig
}

func (ClibanaConfig) Description() string {
	return "Clibana - OpenSearch log tailer"
}

func (ClibanaConfig) Epilogue() string {
	return "For more information, see https://github.com/ivoronin/clibana"
}

func (ClibanaConfig) Version() string {
	return fmt.Sprintf("clibana %s (commit: %s, build date: %s)", version, commit, date)
}

type fieldList []string

func (c *fieldList) UnmarshalText(text []byte) error { //nolint:unparam
	*c = strings.Split(string(text), ",")

	return nil
}
