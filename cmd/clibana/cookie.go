package main

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/opensearch-project/opensearch-go/v2"
)

// dashboardProxyTransport wraps an http.RoundTripper to route requests through
// the OpenSearch Dashboards or Kibana (ElasticSearch) console proxy, adding required headers and cookies.
type dashboardProxyTransport struct {
	base       http.RoundTripper
	cookies    []*http.Cookie
	baseURL    string
	serverType string
}

func (t *dashboardProxyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Extract original path and query string (strip leading slash for proxy path param)
	origPath := strings.TrimPrefix(req.URL.Path, "/")
	if req.URL.RawQuery != "" {
		origPath += "?" + req.URL.RawQuery
	}
	origMethod := req.Method

	// Determine proxy path and XSRF header based on server type
	var proxyPath, xsrfHeader string
	if t.serverType == ServerTypeElasticSearch {
		proxyPath = "/api/console/proxy"
		xsrfHeader = "kbn-xsrf"
	} else {
		proxyPath = "/_dashboards/api/console/proxy"
		xsrfHeader = "osd-xsrf"
	}

	// Build console proxy URL
	proxyURL, err := url.Parse(t.baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}
	proxyURL.Path = proxyPath
	proxyURL.RawQuery = url.Values{
		"path":   {origPath},
		"method": {origMethod},
	}.Encode()

	// Clone request with new URL
	proxyReq := req.Clone(req.Context())
	proxyReq.URL = proxyURL
	proxyReq.Host = proxyURL.Host
	proxyReq.Method = http.MethodPost // Console proxy requires POST

	// Add required XSRF header for dashboard API
	proxyReq.Header.Set(xsrfHeader, "true")

	// Add cookies
	for _, cookie := range t.cookies {
		proxyReq.AddCookie(cookie)
	}

	DebugLogger.Printf("Proxy request: %s %s (server type: %s)", proxyReq.Method, proxyReq.URL.String(), t.serverType)

	return t.base.RoundTrip(proxyReq)
}

// parseCookieFile parses a Netscape/curl format cookie file.
// Format: domain<TAB>flag<TAB>path<TAB>secure<TAB>expiry<TAB>name<TAB>value.
func parseCookieFile(path string) ([]*http.Cookie, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open cookie file: %w", err)
	}
	defer f.Close()

	var cookies []*http.Cookie
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) < 7 {
			continue
		}
		expiry, _ := strconv.ParseInt(fields[4], 10, 64)
		cookies = append(cookies, &http.Cookie{
			Domain:  fields[0],
			Path:    fields[2],
			Secure:  fields[3] == "TRUE",
			Expires: time.Unix(expiry, 0),
			Name:    fields[5],
			Value:   fields[6],
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read cookie file: %w", err)
	}

	if len(cookies) == 0 {
		return nil, fmt.Errorf("no cookies found in file: %s", path)
	}

	return cookies, nil
}

// detectServerTypeFromCookies detects the server type based on cookie paths.
// OpenSearch Dashboards uses "/_dashboards" path, while Kibana (ElasticSearch) uses "/".
func detectServerTypeFromCookies(cookies []*http.Cookie) string {
	for _, c := range cookies {
		if c.Name == "security_authentication" || strings.HasPrefix(c.Name, "security_authentication") {
			if strings.HasPrefix(c.Path, "/_dashboards") {
				return ServerTypeOpenSearch
			}
		}
	}

	return ServerTypeElasticSearch
}

func buildCookieAuthClientConfig(config ClibanaConfig, baseTransport http.RoundTripper) opensearch.Config {
	cookies, err := parseCookieFile(config.CookieFile)
	if err != nil {
		FatalError(err)
	}

	DebugLogger.Printf("Loaded %d cookies from %s", len(cookies), config.CookieFile)

	// Determine server type: use explicit config or auto-detect from cookies
	serverType := config.ServerType
	if serverType == "" {
		serverType = detectServerTypeFromCookies(cookies)
		DebugLogger.Printf("Auto-detected server type: %s", serverType)
	} else {
		DebugLogger.Printf("Using configured server type: %s", serverType)
	}

	proxyTransport := &dashboardProxyTransport{
		base:       baseTransport,
		cookies:    cookies,
		baseURL:    config.URL,
		serverType: serverType,
	}

	return opensearch.Config{
		Transport: proxyTransport,
	}
}
