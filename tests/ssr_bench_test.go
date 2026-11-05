package tests

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"sparrow-ssr/models"
	"sparrow-ssr/routers"
	"sparrow-ssr/services"

	beego "github.com/beego/beego/v2/server/web"
)

const ssrBenchmarkHost = "test-travel-templates.beta.123presto.com"

var (
	ssrBenchmarkOnce     sync.Once
	ssrBenchmarkErr      error
	ssrBenchmarkLoader   *services.ContentLoader
	ssrBenchmarkLocation models.PageLocation
)

// Initialize Beego and register the real application routes only once.
// This also supports repeated runs with -count=5.
func initializeSSRBenchmark() error {
	ssrBenchmarkOnce.Do(func() {
		projectRoot, err := filepath.Abs("..")
		if err != nil {
			ssrBenchmarkErr = err
			return
		}

		contentRoot := filepath.Join(projectRoot, "content")
		deploymentFile := filepath.Join(projectRoot, "conf", "deployment.json")

		catalog, err := services.NewMultiCatalog(
			contentRoot,
			[]string{deploymentFile},
		)
		if err != nil {
			ssrBenchmarkErr = fmt.Errorf("initialize catalog: %w", err)
			return
		}

		location, found := catalog.Resolve(ssrBenchmarkHost, "/")
		if !found {
			ssrBenchmarkErr = fmt.Errorf("published homepage route not found")
			return
		}

		loader, err := services.NewContentLoader(contentRoot)
		if err != nil {
			ssrBenchmarkErr = fmt.Errorf("initialize content loader: %w", err)
			return
		}

		// Match the initialization strategy from the passing Step 5 test.
		beego.TestBeegoInit(projectRoot)
		routers.Register(catalog, loader)

		ssrBenchmarkLoader = loader
		ssrBenchmarkLocation = location
	})

	return ssrBenchmarkErr
}

// requestSSRBenchmarkPage makes one real HTTP request over loopback.
// The Host header is required by the application's strict routing.
func requestSSRBenchmarkPage(client *http.Client, targetURL string) (int, int64, error) {
	request, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		return 0, 0, err
	}
	request.Host = ssrBenchmarkHost

	response, err := client.Do(request)
	if err != nil {
		return 0, 0, err
	}

	bytesReceived, readErr := io.Copy(io.Discard, response.Body)
	closeErr := response.Body.Close()
	if readErr != nil {
		return response.StatusCode, bytesReceived, readErr
	}
	if closeErr != nil {
		return response.StatusCode, bytesReceived, closeErr
	}

	return response.StatusCode, bytesReceived, nil
}

func BenchmarkSSRHTTP(b *testing.B) {
	if err := initializeSSRBenchmark(); err != nil {
		b.Fatal(err)
	}

	// A real loopback listener, unlike directly calling ServeHTTP.
	server := httptest.NewServer(beego.BeeApp.Handlers)
	defer server.Close()

	client := server.Client()
	targetURL := server.URL + "/"
	loader := ssrBenchmarkLoader
	location := ssrBenchmarkLocation

	// Verify a complete HTML response before timing any requests.
	request, err := http.NewRequest(http.MethodGet, targetURL, nil)
	if err != nil {
		b.Fatal(err)
	}
	request.Host = ssrBenchmarkHost

	response, err := client.Do(request)
	if err != nil {
		b.Fatal(err)
	}
	body, readErr := io.ReadAll(response.Body)
	closeErr := response.Body.Close()
	if readErr != nil {
		b.Fatal(readErr)
	}
	if closeErr != nil {
		b.Fatal(closeErr)
	}
	if response.StatusCode != http.StatusOK ||
		!strings.HasPrefix(response.Header.Get("Content-Type"), "text/html") ||
		!strings.Contains(string(body), "Experience Mexico") {
		b.Fatalf("unexpected homepage response: status=%d, content-type=%q, bytes=%d",
			response.StatusCode, response.Header.Get("Content-Type"), len(body))
	}

	b.Run("ColdWithInvalidation", func(b *testing.B) {
		before := loader.Stats().FileReads
		b.SetBytes(int64(len(body)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			// Simulate a cold content cache for each request.
			// Invalidation is included in this benchmark's timing.
			if err := loader.Invalidate(
				location.HeaderJSON, location.PageJSON, location.FooterJSON,
			); err != nil {
				b.Fatal(err)
			}

			status, size, err := requestSSRBenchmarkPage(client, targetURL)
			if err != nil || status != http.StatusOK || size == 0 {
				b.Fatalf("cold request failed: status=%d, bytes=%d, err=%v",
					status, size, err)
			}
		}

		b.StopTimer()
		if got := loader.Stats().FileReads - before; got != uint64(3*b.N) {
			b.Fatalf("expected %d cold file reads, got %d", 3*b.N, got)
		}
	})

	b.Run("Warm", func(b *testing.B) {
		// Prime the cache outside the timed region.
		status, size, err := requestSSRBenchmarkPage(client, targetURL)
		if err != nil || status != http.StatusOK || size == 0 {
			b.Fatalf("warmup failed: status=%d, bytes=%d, err=%v", status, size, err)
		}

		before := loader.Stats().FileReads
		b.SetBytes(int64(len(body)))
		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			status, size, err := requestSSRBenchmarkPage(client, targetURL)
			if err != nil || status != http.StatusOK || size == 0 {
				b.Fatalf("warm request failed: status=%d, bytes=%d, err=%v", status, size, err)
			}
		}

		b.StopTimer()
		if got := loader.Stats().FileReads - before; got != 0 {
			b.Fatalf("warm requests unexpectedly read %d files", got)
		}
	})

	b.Run("ParallelWarm", func(b *testing.B) {
		status, size, err := requestSSRBenchmarkPage(client, targetURL)
		if err != nil || status != http.StatusOK || size == 0 {
			b.Fatalf("parallel warmup failed: status=%d, bytes=%d, err=%v", status, size, err)
		}

		before := loader.Stats().FileReads
		failures := make(chan string, 1)
		b.SetBytes(int64(len(body)))
		b.ReportAllocs()
		b.ResetTimer()

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				status, size, err := requestSSRBenchmarkPage(client, targetURL)
				if err != nil || status != http.StatusOK || size == 0 {
					select {
					case failures <- fmt.Sprintf("status=%d, bytes=%d, err=%v", status, size, err):
					default:
					}
				}
			}
		})

		b.StopTimer()
		select {
		case failure := <-failures:
			b.Fatalf("concurrent HTTP request failed: %s", failure)
		default:
		}
		if got := loader.Stats().FileReads - before; got != 0 {
			b.Fatalf("parallel warm requests unexpectedly read %d files", got)
		}
	})
}
