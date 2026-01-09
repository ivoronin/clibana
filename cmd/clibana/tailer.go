package main

import (
	"context"
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
	Client              *opensearch.Client
	ClibanaConfig       ClibanaConfig
	SearchAfter         []interface{}
	currentPollInterval time.Duration
}

func NewTailer(client *opensearch.Client, clibanaConfig ClibanaConfig) *Tailer {
	return &Tailer{
		Client:              client,
		ClibanaConfig:       clibanaConfig,
		currentPollInterval: MinPollInterval,
	}
}

func (t *Tailer) StartProducer(ctx context.Context) <-chan SearchResponseHitsHit {
	hitChan := make(chan SearchResponseHitsHit, HitChannelBuffer)

	go func() {
		defer close(hitChan)

		size := SearchRequestSize

		for {
			// Проверяем не отменён ли контекст
			select {
			case <-ctx.Done():
				return
			default:
			}

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
				select {
				case hitChan <- hit:
				case <-ctx.Done():
					return
				}
			}

			hitsReceived := len(response.Hits.Hits)

			if hitsReceived == size {
				// Получили полный batch - продолжаем без задержки
				t.currentPollInterval = MinPollInterval
			} else {
				if t.ClibanaConfig.Search.Follow {
					// Вычисляем задержку по плавной формуле: sleep = max * (1 - ratio)⁴
					// Это даёт почти нулевую задержку при 8k-10k событиях, плавно увеличивая до max при малом количестве
					ratio := float64(hitsReceived) / float64(size)
					deficit := 1.0 - ratio
					// Возводим в 4-ю степень для крутой кривой
					sleepFactor := deficit * deficit * deficit * deficit
					t.currentPollInterval = time.Duration(float64(MaxPollInterval) * sleepFactor)

					// Спим с возможностью прерывания по контексту
					select {
					case <-time.After(t.currentPollInterval):
					case <-ctx.Done():
						return
					}
				} else {
					// Single-shot режим - выходим когда получили неполный batch
					return
				}
			}
		}
	}()

	return hitChan
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
		fieldNames := make([]string, 0, len(t.ClibanaConfig.Search.Fields))
		for _, field := range t.ClibanaConfig.Search.Fields {
			fieldNames = append(fieldNames, field.Name)
		}
		query["_source"] = fieldNames
	}

	body, err := json.Marshal(query)
	if err != nil {
		FatalError(fmt.Errorf("failed to marshal search request body to JSON: %w", err))
	}

	return strings.NewReader(string(body))
}
