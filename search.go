package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/opensearch-project/opensearch-go/v2"
)

func search(client *opensearch.Client, clibanaConfig ClibanaConfig) {
	tailer := NewTailer(client, clibanaConfig)

	for hit := range tailer.Tail() {
		var output string

		if clibanaConfig.Search.Fields != nil {
			var values []string

			for _, field := range clibanaConfig.Search.Fields {
				if strValue, ok := getNestedField(hit.Source, field); ok {
					values = append(values, strValue)
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
