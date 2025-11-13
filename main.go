package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	defaultPort   = "5000"
	bugsbyBaseURL = "https://bugs-service.infra.corp.arista.io"
	bugsbyV3      = bugsbyBaseURL + "/v3/bugs"
	bugsbyV1      = bugsbyBaseURL + "/v1/bugs"
	tokenFileRel  = ".local/state/artools_oauth2" // relative to home dir
)

var retryStatus = map[int]bool{
	429: true,
	500: true,
	502: true,
	503: true,
	504: true,
}

// getAuthToken tries (in order):
// 1) BUGSBY_AUTH_TOKEN env var
// 2) YAML token file ~/.local/state/artools_oauth2 with key "access_token"
func getAuthToken() (string, error) {
	if t := os.Getenv("BUGSBY_AUTH_TOKEN"); t != "" {
		return t, nil
	}

	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := filepath.Join(h, tokenFileRel)
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("no env token and failed to open token file: %w", err)
	}
	defer f.Close()

	var data map[string]interface{}
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&data); err != nil {
		return "", fmt.Errorf("failed to decode token yaml: %w", err)
	}

	if at, ok := data["access_token"].(string); ok && at != "" {
		return at, nil
	}
	return "", errors.New("access_token not found in token file")
}

// httpGetWithRetry performs an HTTP GET with simple retry/backoff for
// transient status codes and network errors.
func httpGetWithRetry(ctx context.Context, client *http.Client, url string, headers map[string]string, params map[string]string) (*http.Response, error) {
	// build URL with params
	if len(params) > 0 {
		q := make([]string, 0, len(params))
		for k, v := range params {
			q = append(q, fmt.Sprintf("%s=%s", k, urlQueryEscape(v)))
		}
		if strings.Contains(url, "?") {
			url = url + "&" + strings.Join(q, "&")
		} else {
			url = url + "?" + strings.Join(q, "&")
		}
	}

	var lastErr error
	// retry policy: up to 3 attempts, backoff factor 1s -> 1s, 2s, 4s
	backoffs := []time.Duration{1 * time.Second, 2 * time.Second, 4 * time.Second}
	for attempt := 0; attempt < len(backoffs); attempt++ {
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, err
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			// network error -> retry
			log.Printf("request error (attempt %d): %v -- will retry after %s", attempt+1, err, backoffs[attempt])
			time.Sleep(backoffs[attempt])
			continue
		}

		// If status is retryable, drain body and retry
		if retryStatus[resp.StatusCode] {
			bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
			_ = resp.Body.Close()
			lastErr = fmt.Errorf("status %d: %s", resp.StatusCode, string(bodyBytes))
			log.Printf("received retryable status %d (attempt %d) -> retrying after %s", resp.StatusCode, attempt+1, backoffs[attempt])
			time.Sleep(backoffs[attempt])
			continue
		}

		// non-retryable or success -> return
		return resp, nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("failed after retries")
	}
	return nil, lastErr
}

func urlQueryEscape(s string) string {
	// simple escape for query params
	return strings.ReplaceAll(s, " ", "+")
}

func makeHTTPClient() *http.Client {
	// default http client with reasonable timeouts
	return &http.Client{
		Timeout: 30 * time.Second,
	}
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(payload)
}

func homeHandler(token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authPresent := token != ""
		writeJSON(w, 200, map[string]any{
			"status":       "ok",
			"message":      "Bugsby Standalone Server is running",
			"authenticated": authPresent,
			"endpoints": map[string]string{
				"health":    "GET /",
				"test_bug":  "GET /test",
				"get_bug":   "GET /bug/{id}",
				"query_bugs": "GET /bugs?q=<query>",
			},
		})
	}
}

func testHandler(client *http.Client, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// reuse getBug logic for bug 1196080
		serveGetBug(w, r, client, token, 1196080)
	}
}

func serveGetBug(w http.ResponseWriter, r *http.Request, client *http.Client, token string, bugID int) {
	ctx, cancel := context.WithTimeout(r.Context(), 35*time.Second)
	defer cancel()

	url := bugsbyV3
	params := map[string]string{
		"q":     fmt.Sprintf("id==%d", bugID),
		"limit": "1",
	}
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}

	log.Printf("Making request to: %s", url)
	log.Printf("Params: %+v", params)

	resp, err := httpGetWithRetry(ctx, client, url, headers, params)
	if err != nil {
		log.Printf("Error calling Bugsby v3: %v", err)
		writeJSON(w, 500, map[string]any{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024)) // limit 2MB

	log.Printf("Response status: %d", resp.StatusCode)

	if resp.StatusCode == http.StatusOK {
		var data map[string]any
		if err := json.Unmarshal(bodyBytes, &data); err != nil {
			log.Printf("JSON parse error: %v", err)
			writeJSON(w, 500, map[string]any{
				"success":      false,
				"error":        fmt.Sprintf("failed to parse JSON response: %v", err),
				"raw_response": string(bodyBytes)[:minInt(len(bodyBytes), 1000)],
			})
			return
		}
		// expect top-level "bugs" list (mimic Python behaviour)
		if bugs, ok := data["bugs"].([]any); ok && len(bugs) > 0 {
			writeJSON(w, 200, map[string]any{
				"success": true,
				"bug":     bugs[0],
			})
			return
		}
		writeJSON(w, 404, map[string]any{
			"success": false,
			"error":   fmt.Sprintf("Bug %d not found", bugID),
		})
		return
	}

	// other status codes
	writeJSON(w, resp.StatusCode, map[string]any{
		"success": false,
		"error":   fmt.Sprintf("API returned status %d", resp.StatusCode),
		"details": string(bodyBytes),
	})
}

func getBugHandler(client *http.Client, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// path like /bug/1196080
		parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
		if len(parts) < 2 {
			writeJSON(w, 400, map[string]any{"success": false, "error": "bad request"})
			return
		}
		idStr := parts[1]
		id, err := strconv.Atoi(idStr)
		if err != nil {
			writeJSON(w, 400, map[string]any{"success": false, "error": "invalid bug id"})
			return
		}
		serveGetBug(w, r, client, token, id)
	}
}

func getBugsHandler(client *http.Client, token string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		limit := r.URL.Query().Get("limit")
		if q == "" {
			writeJSON(w, 400, map[string]any{
				"success": false,
				"error":   "Query parameter 'q' is required",
				"example": "/bugs?q=id==1196080",
			})
			return
		}
		if limit == "" {
			limit = "100"
		}

		ctx, cancel := context.WithTimeout(r.Context(), 35*time.Second)
		defer cancel()

		url := bugsbyV3
		params := map[string]string{
			"q":     q,
			"limit": limit,
		}
		headers := map[string]string{
			"Content-Type": "application/json",
		}
		if token != "" {
			headers["Authorization"] = "Bearer " + token
		}

		log.Printf("Making request to: %s", url)
		log.Printf("Params: %+v", params)

		resp, err := httpGetWithRetry(ctx, client, url, headers, params)
		if err != nil {
			log.Printf("Error calling Bugsby v3: %v", err)
			writeJSON(w, 500, map[string]any{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(io.LimitReader(resp.Body, 5*1024*1024))

		if resp.StatusCode == http.StatusOK {
			var data map[string]any
			if err := json.Unmarshal(bodyBytes, &data); err != nil {
				log.Printf("JSON parse error: %v", err)
				writeJSON(w, 500, map[string]any{
					"success":      false,
					"error":        fmt.Sprintf("failed to parse JSON response: %v", err),
					"raw_response": string(bodyBytes)[:minInt(len(bodyBytes), 1000)],
				})
				return
			}
			bugs := []any{}
			if b, ok := data["bugs"].([]any); ok {
				bugs = b
			}
			writeJSON(w, 200, map[string]any{
				"success": true,
				"count":   len(bugs),
				"bugs":    bugs,
			})
			return
		}

		writeJSON(w, resp.StatusCode, map[string]any{
			"success": false,
			"error":   fmt.Sprintf("API returned status %d", resp.StatusCode),
			"details": string(bodyBytes),
		})
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	port := flag.String("port", defaultPort, "port to run server on")
	flag.Parse()

	token, err := getAuthToken()
	if err != nil {
		// Not fatal - server still runs but reports unauthenticated in /.
		log.Printf("warning: couldn't load auth token: %v", err)
		token = ""
	} else {
		log.Printf("auth token loaded (length=%d)", len(token))
	}

	client := makeHTTPClient()

	http.HandleFunc("/", homeHandler(token))
	http.HandleFunc("/test", testHandler(client, token))
	http.HandleFunc("/bug/", getBugHandler(client, token)) // note: expects /bug/<id>
	http.HandleFunc("/bugs", getBugsHandler(client, token))

	addr := ":" + *port
	log.Printf("🚀 Starting Bugsby Standalone Server on %s ...", addr)
	log.Printf("Available endpoints: GET / , GET /test , GET /bug/{id} , GET /bugs?q=...")

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}