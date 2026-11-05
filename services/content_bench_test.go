package services

import "testing"

func benchmarkContentLoad(
	b *testing.B,
	cached bool,
) {
	catalog, err := NewCatalog(
		"../content",
		"../conf/deployment.json",
	)
	if err != nil {
		b.Fatal(err)
	}

	loader, err := NewContentLoader("../content")
	if err != nil {
		b.Fatal(err)
	}

	location, ok := catalog.Resolve(testSiteHost, "/")
	if !ok {
		b.Fatal("home route not found")
	}

	// Warm up the cache before starting the cached benchmark.
	if cached {
		if _, err := loader.Load(location); err != nil {
			b.Fatal(err)
		}
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if cached {
			_, err = loader.Load(location)
		} else {
			_, err = loader.loadUncached(location)
		}

		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkContentLoadUncached(b *testing.B) {
	benchmarkContentLoad(b, false)
}

func BenchmarkContentLoadCached(b *testing.B) {
	benchmarkContentLoad(b, true)
}
