package main

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type CatIndicesResponseItem struct {
	Index       string `json:"index"`
	Health      string `json:"health"`
	Status      string `json:"status"`
	Pri         string `json:"pri"`
	Rep         string `json:"rep"`
	DocsCount   string `json:"docs.count"`
	DocsDeleted string `json:"docs.deleted"`
	StoreSize   string `json:"store.size"`
	PriStorSize string `json:"pri.store.size"`
}

type CatIndicesResponse []CatIndicesResponseItem

func indices(client *opensearch.Client, clibanaConfig ClibanaConfig) {
	request := opensearchapi.CatIndicesRequest{
		Format: "json",
		Index:  []string{clibanaConfig.Index},
	}

	response := doRequest[CatIndicesResponse](client, request)

	writer := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	if !clibanaConfig.Indices.Quiet {
		columns := []string{"index", "health", "status", "pri", "rep", "docs.count", "docs.deleted", "store.size", "pri.store"}
		writer.Write([]byte(fmt.Sprintf("%s\n", strings.Join(columns, "\t"))))
	}

	for _, idx := range response {
		writer.Write([]byte(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
			idx.Index, idx.Health, idx.Status, idx.Pri, idx.Rep, idx.DocsCount, idx.DocsDeleted, idx.StoreSize, idx.PriStorSize,
		)))
	}

	writer.Flush()
}
