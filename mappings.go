package main

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"text/tabwriter"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type Field struct {
	Name string
	Type string
}

type Mappings struct {
	Properties map[string]interface{} `json:"properties"`
}

type IndicesGetMappingResponseItem struct {
	Mappings Mappings `json:"mappings"`
}

type IndicesGetMappingResponse map[string]IndicesGetMappingResponseItem

func mappings(client *opensearch.Client, clibanaConfig ClibanaConfig) {
	request := opensearchapi.IndicesGetMappingRequest{
		Index: []string{clibanaConfig.Index},
	}
	response := doRequest[IndicesGetMappingResponse](client, request)

	var keys []Field

	for _, item := range response {
		keys = append(keys, getIndexMapKeysFlattened(item.Mappings.Properties, "")...)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].Name < keys[j].Name
	})

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !clibanaConfig.Mappings.Quiet {
		writer.Write([]byte("field\ttype\n"))
	}

	for _, key := range slices.Compact(keys) {
		writer.Write([]byte(fmt.Sprintf("%s\t%s\n", key.Name, key.Type)))
	}

	writer.Flush()
}

func getIndexMapKeysFlattened(props map[string]interface{}, prefix string) []Field {
	var keys []Field

	for key, value := range props {
		switch v := value.(type) {
		case map[string]interface{}:
			if _, ok := v["properties"]; ok {
				properties := mustAssertType[map[string]interface{}](v["properties"])
				for _, subfield := range getIndexMapKeysFlattened(properties, key+".") {
					keys = append(keys, Field{
						Name: prefix + subfield.Name,
						Type: subfield.Type,
					})
				}
			} else {
				typ := mustAssertType[string](v["type"])
				keys = append(keys, Field{
					Name: prefix + key,
					Type: typ,
				})
			}
		default:
			panic(fmt.Sprintf("unexpected type: %+v", value))
		}
	}

	return keys
}
