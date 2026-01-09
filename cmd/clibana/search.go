package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/opensearch-project/opensearch-go/v2"
)

func search(client *opensearch.Client, clibanaConfig ClibanaConfig) {
	// Создаем контекст с обработкой сигналов для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обрабатываем Ctrl+C для остановки
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	tailer := NewTailer(client, clibanaConfig)
	hitChan := tailer.StartProducer(ctx)

	for hit := range hitChan {
		var output string

		if clibanaConfig.Search.Fields != nil {
			var values []string

			for _, field := range clibanaConfig.Search.Fields {
				if value, ok := getNestedField(hit.Source, field.Name); ok {
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
