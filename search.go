package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/opensearch-project/opensearch-go/v2"
)

func search(client *opensearch.Client, clibanaConfig ClibanaConfig) {
	tailer := NewTailer(client, clibanaConfig)

	for hit := range tailer.Tail() {
		var output string

		if clibanaConfig.Search.Fields != nil {
			var values []string

			for _, field := range clibanaConfig.Search.Fields {
				if value, ok := getNestedField(hit.Source, field.Name); ok {
					if colorCode, ok := Colors[field.Color]; ok {
						colorer := color.New(colorCode)
						value = colorer.Sprint(value)
					}
					values = append(values, value)
				}
			}

			output = strings.Join(values, " ")
		} else {
			buf, err := json.Marshal(hit.Source)
			if err != nil {
				FatalError(fmt.Errorf("failed to marshal JSON: %w", err))
			}

			output = string(buf)
		}

		fmt.Println(output) //nolint:forbidigo
	}
}
