package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/v2"
	"github.com/opensearch-project/opensearch-go/v2/opensearchapi"
)

type SearchResponseHitsHit struct {
	Sort   []interface{}          `json:"sort"`
	Source map[string]interface{} `json:"_source"`
}

type SearchResponseHits struct {
	Hits []SearchResponseHitsHit `json:"hits"`
}

type SearchResponse struct {
	Hits SearchResponseHits `json:"hits"`
}

type Tailer struct {
	Client        *opensearch.Client
	ClibanaConfig ClibanaConfig
	SearchAfter   []interface{}
}

func NewTailer(client *opensearch.Client, clibanaConfig ClibanaConfig) *Tailer {
	return &Tailer{
		Client:        client,
		ClibanaConfig: clibanaConfig,
	}
}

func (t *Tailer) Tail() func(func(SearchResponseHitsHit) bool) {
	size := SearchRequestSize
	return func(yield func(SearchResponseHitsHit) bool) {
		for {
			requestBody := t.buildSearchRequestBody()
			request := opensearchapi.SearchRequest{
				Index: []string{t.ClibanaConfig.Index},
				Body:  requestBody,
				Sort:  []string{"@timestamp:asc"},
				Size:  &size,
			}

			response := doRequest[SearchResponse](t.Client, request)

			for _, hit := range response.Hits.Hits {
				t.SearchAfter = hit.Sort
				if !yield(hit) {
					break
				}
			}

			if len(response.Hits.Hits) != size {
				if t.ClibanaConfig.Search.Follow {
					time.Sleep(TailSleep * time.Second)
				} else {
					break
				}
			}
		}
	}
}

func (t *Tailer) buildSearchRequestBody() *strings.Reader {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"query_string": map[string]interface{}{
							"query": t.ClibanaConfig.Search.Query,
						},
					},
					map[string]interface{}{
						"range": map[string]interface{}{
							"@timestamp": map[string]interface{}{
								"gte": t.ClibanaConfig.Search.Start,
								"lte": t.ClibanaConfig.Search.End,
							},
						},
					},
				},
			},
		},
	}

	if t.SearchAfter != nil {
		query["search_after"] = t.SearchAfter
	}

	if len(t.ClibanaConfig.Search.Fields) > 0 {
		query["_source"] = append(t.ClibanaConfig.Search.Fields, "@timestamp")
	}

	body, err := json.Marshal(query)
	if err != nil {
		FatalError(fmt.Errorf("failed to marshal search request body to JSON: %w", err))
	}

	return strings.NewReader(string(body))
}
