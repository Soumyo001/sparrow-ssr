# Sparrow SSR

**A Go/Beego server-side rendering engine for deployment-metadata-driven, multi-site websites.**

Sparrow SSR is an internship assessment project for **W3 Engineers**. It accepts an HTTP request, resolves the requested **hostname and URL path** against deployment metadata and page canonical URLs, loads the associated JSON documents, and renders complete HTML on the server with Beego templates.

The assessment uses **local JSON files arranged in the same relative directory structure as the publisher's S3/CDN object keys**. It does **not** fetch the JSON documents from S3 at request time. External CSS and JavaScript URLs present in the JSON are preserved in the rendered HTML.

> **Implementation scope (11 November 2026):** Sparrow SSR supports deployment-metadata-driven multi-site server-side rendering using publisher-managed JSON, alongside a client-rendered accommodation-search application powered by Neon PostgreSQL, Beego ORM and pgx. The current database seed contains one site, eight reusable locations, fifteen properties and fifteen availability windows. Property prices, guest capacities, photographs and availability are illustrative assessment data, not live booking-provider information. Two-site HTTP routing has been tested with synthetic fixtures; real PostgreSQL search has been verified against the seeded site.


## 1. Technology and requirements

| Component | Current implementation |
| --- | --- |
| Language | Go; `go.mod` specifies `go 1.25` |
| Web framework | Beego v2.1.0 (`github.com/beego/beego/v2/server/web`) |
| CLI used in local testing | Bee CLI v2.3.0 (optional; distinct from the Beego library) |
| Templates | Go/Beego HTML templates in `views/page.tpl` |
| Content source | Local JSON fixtures under `content/` |
| URL resolution | In-memory `(hostname, canonical path)` catalog |
| JSON content cache | Process-local, typed, concurrency-safe caches |
| Tests and performance tools | Go `testing`, `httptest`, race detector, benchmarks, `pprof` |

Run the commands below from the **project root**. A compatible Go toolchain is required; Bee CLI is only needed for `bee run`.

## 2. How the rendering pipeline works

```text
GET request (Host + URL path)
             |
             v
        Beego router
             |
             v
       PageController
             |
             v
Catalog.Resolve(host, path)       <-- built from deployment metadata
             |                         and head.canonical_url
             v
    PageLocation (three paths)
             |
             v
  ContentLoader.Load(location)
    |          |          |
  header     page       footer       <-- concurrent on cache misses;
    |          |          |               reused on cache hits
    +----------+----------+
               |
               v
      models.PageContent
               |
               v
      views/page.tpl
               |
               v
     HTML HTTP response
```

### Routing and deployment mapping

1. `services.NewCatalog` reads a deployment response containing `data.published_domain` and `data.uploaded_file_paths`.
2. It maps the supplied CDN JSON object paths to **local relative content paths**, checks their existence, and identifies the shared `header.json` and `footer.json`.
3. It reads each routable page's `head.canonical_url` and creates an immutable lookup keyed by normalized **hostname plus URL path**. `links.json` is not treated as a page.
4. `services.NewMultiCatalog` combines independently validated deployment catalogs and rejects duplicate `(host, path)` routes.
5. `controllers.PageController` uses the **actual request Host** and URL path; an unregistered host or path returns HTTP 404.

Storage filenames are **not inferred from public URLs**. For example, the public URL `/blog/untitled-post-wrurmp` maps to `untitled-post-wrurmp.json` directly under its site's local storage directory. This mapping comes from deployment metadata and canonical URLs, not a hardcoded hostname-to-directory formula.

### Content and HTML

The loader combines a shared `SiteHeader`, a `PageDocument`, and a shared `SiteFooter`. The controller provides the page title, metadata, canonical URL, CSS/JS references, and publisher-provided HTML fragments to `views/page.tpl`.

Shared CSS, template CSS, page CSS, shared JS, and page JS remain external URLs when present; nullable page-level CSS/JS are handled as optional fields. The HTML fragments are intentionally treated as trusted publisher-controlled content (`template.HTML`)—they must **not** be treated as safe when obtained from untrusted visitors.

## 3. Project layout

```text
sparrow-ssr/
├── main.go
├── go.mod
├── conf/
│   ├── app.conf
│   └── deployment.json
├── models/
│   ├── deployment.go
│   ├── content.go
│   ├── route.go
│   └── render.go
├── services/
│   ├── catalog.go
│   ├── multi_catalog.go
│   ├── content_loader.go
│   ├── file_cache.go
│   ├── catalog_test.go
│   ├── multi_catalog_test.go
│   ├── content_loader_test.go
│   ├── content_cache_test.go
│   └── content_bench_test.go
├── controllers/
│   └── page.go
├── routers/
│   └── router.go
├── views/
│   └── page.tpl
├── tests/
│   ├── ssr_integration_test.go
│   └── ssr_bench_test.go
└── content/
    └── beta/preview/test-travel-templates/
        ├── header.json
        ├── footer.json
        ├── home.json
        ├── privacy-policy.json
        ├── site-terms.json
        ├── untitled-post-wrurmp.json
        └── links.json
```

`services/` contains application-specific logic; controllers handle HTTP/template data; `models/` represent observed JSON fields. The test files under `tests/` exercise the Beego HTTP path, not just isolated services.

## 4. Running the application

### PostgreSQL initialization

Configure `.env` with a valid Neon PostgreSQL connection string requiring TLS.

For a newly created database, run:

```bash
go run ./cmd/dbsetup
go run ./cmd/dbseed
```

The setup command creates the normalized `sites`, `locations`, `properties`, and `property_availability` tables, indexes and foreign-key constraints.

The seed command inserts the application's initial sample property and availability records. It does not read property JSON files.

Re-running the seed command does not overwrite existing property records. The original verification inserted 1 site, 8 locations, 15 properties and 15 availability records; a second execution inserted zero new records.

### Running the application

```bash
go run .
```

The application initializes PostgreSQL before accepting requests. The property API uses PostgreSQL as its runtime property and availability source. Publisher-managed page JSON under `content/` remains necessary for server-side rendering.

### Single-deployment demo

```bash
go mod download
go run .
# Or, if Bee CLI is installed:
bee run
```

The latest submitted `main.go` explicitly runs Beego on **`127.0.0.1:8080`**, so the development server listens on the local machine only. By default, it loads `conf/deployment.json` and serves four indexed pages.

Because routing requires the published hostname, make requests with a matching `Host` header:

```bash
curl -i \
  -H 'Host: test-travel-templates.beta.123presto.com' \
  http://127.0.0.1:8080/
```

The current deployment's known paths are `/`, `/privacy-policy`, `/site-terms`, and `/blog/untitled-post-wrurmp`. An unknown hostname or unknown canonical path should receive HTTP 404. A plain request to `127.0.0.1:8080` without the expected Host is also 404; no development host alias is configured.

### Multiple deployments at startup

`main.go` reads an **optional, comma-separated** `SPARROW_DEPLOYMENTS` environment variable. For example, **after you supply a valid second deployment metadata file and its local content**:

```bash
SPARROW_DEPLOYMENTS='conf/deployment.json,conf/second-deployment.json' go run .
```

`conf/second-deployment.json` is illustrative; **it is not bundled with this project**. The existing real demo consists of one deployment. The multi-site tests create disposable synthetic sites instead.


The application uses `godotenv.Load()` to load an optional `.env` file. Existing process environment variables take precedence over values in that file.

The web application prefers `DATABASE_URL_POOLED` when it is nonempty, otherwise it uses `DATABASE_URL`. This is selection based on whether the variable is present, not automatic connection-failure recovery.

The database setup command currently uses `DATABASE_URL`. The seed command prefers `DATABASE_URL_DIRECT`, falling back to `DATABASE_URL`. Database setup and seeding should use direct PostgreSQL connections.

The `.env` file must not be committed to version control. PostgreSQL credentials are required to run the application, while ordinary unit and HTTP tests do not require Neon. The Neon integration tests are explicitly enabled using `SPARROW_TEST_NEON=1`.

The application defaults to `127.0.0.1:8080`. When `PORT` is configured, it binds to `0.0.0.0:<PORT>`. Render deployment has not yet been verified.

## 5. Concurrent content loading and caching

### Selected design

`services.ContentLoader` caches **decoded JSON documents individually**, not entire rendered HTML pages:

- `fileCache[models.SiteHeader]` for shared headers;
- `fileCache[models.PageDocument]` for individual pages;
- `fileCache[models.SiteFooter]` for shared footers.

Each cache key is a **storage-relative file path** (for example, `beta/preview/site-a/header.json`), so two sites with identically named files do not collide. Each `ContentLoader` owns its caches; they are **not shared across server processes**.

On a cold or partially cached request, the loader can start three independent goroutines and wait with `sync.WaitGroup`. A generic mutex-protected cache stores in-progress entries with a `ready` channel. Concurrent requests for the **same missing file** wait for the first load, avoiding redundant reads and decodes. File I/O takes place outside the cache map mutex. On a fully warm request, the loader returns cached decoded content **without launching new goroutines**.

`ContentLoader.Stats()` reports attempted file reads and typed cache entry counts. `ContentLoader.Invalidate(relativePaths...)` explicitly evicts selected files. A failed file load is not retained as a successful cache entry. In-flight requests may still complete using previous values after invalidation; invalidation is **not an atomic site-wide publication mechanism**.

The cache currently has **no TTL, eviction policy, or memory-size bound**, and there is no automatic deployment-triggered invalidation. JSON changes require a process restart or an explicit call to `Invalidate` (which is a service method, **not a public endpoint**).

### Tested behaviors

Tests submitted on 5 November 2026 passed for shared header/footer reuse, 50 simultaneous cold-homepage loads producing **three attempted file reads**, targeted invalidation, rejection of ordinary lexical path traversal, multi-site cache separation, and the race detector. The two-site HTTP integration fixture exercises two hosts with identical paths; four pages across those two sites required **eight unique document reads**, and repeated requests caused no additional reads. These results apply to the tested fixture and conditions.

## 6. Optimization alternatives and selection rationale

**Instructor question:** *“List the available options of optimization and decide which approach is better and why.”*

There is no finite list of every possible optimization; the table evaluates the **relevant options for this implementation**. Only the selected content-cache design has been directly compared with uncached loading in the submitted benchmarks. Other choices remain candidates, not measured winners or losers.

| Optimization option | Benefits | Costs / limits | Decision |
| --- | --- | --- | --- |
| Scan JSON directories or routes on every request | Simple initial implementation | Repeated file discovery and parsing | **Not selected**; build a catalog at startup |
| Startup host+canonical-path lookup index | Request-time map lookup; deployment metadata remains authoritative | Requires catalog rebuild for deployment changes | **Selected** |
| Sequential header/page/footer reads | Simple control flow | Independent cache misses cannot overlap | **Not selected** for uncached loading |
| Concurrent reads with goroutines + `WaitGroup` | Overlaps independent file I/O | Goroutine and synchronization overhead | **Selected only for cold/partial misses** |
| No content cache | Always reads current local files | Repeated I/O, decode and allocations | **Benchmark baseline**, not selected |
| Cache decoded JSON in Go process memory | Reuses headers/footers, avoids I/O and JSON decode on hits, no network round trip | Process-local, potentially stale, unbounded memory | **Selected and measured** |
| Cache complete rendered HTML | Potentially avoids repeated template execution | More complicated page/site invalidation; higher duplicated HTML memory | **Deferred**, not benchmarked |
| Redis/distributed cache | Shared cache across server instances; centralized invalidation is possible with additional design | Separate service/network traffic, serialization and invalidation complexity | **Not selected for the current single-process demo**; not directly benchmarked |
| Local + Redis hybrid cache | Local hits plus distributed coordination | Two-layer invalidation and operational complexity | **Deferred** |
| `sync.Map` or `sync.RWMutex` instead of mutex map | Might help particular contention patterns | Requires dedicated benchmark and correctness review | **Not changed** |
| Bounded cache, TTL, LRU or explicit deployment versioning | Controls memory growth/staleness | Requires lifetime, eviction and publishing semantics | **Future work** |
| Preallocate buffers, stream template output or cache templates/HTML | Might reduce response allocations | Must preserve exact HTML and correct Beego response behavior | **Investigate only with A/B tests** |
| HTTP compression, browser caching, CDN optimization | Can reduce transferred bytes or asset delivery latency | Separate from JSON decode; depends on proxy, assets and clients | **Out of scope for current service benchmark** |
| Additional goroutines on warm-cache hits | No independent filesystem I/O remains | Adds scheduling/synchronization overhead | **Not selected** |

### Why in-process decoded JSON caching was selected over Redis

The assessment's deployed server is currently a **single Beego process reading local files**. Cached Go structs can be reused directly in the same process, with no Redis service or network request. Individual file caching also shares one header and footer among multiple pages while retaining distinct storage-path keys across sites.

Redis becomes valuable when **multiple application instances** need a shared cache, coordinated updates, or more extensive operational controls. It is **not claimed to be intrinsically slower in every deployment**; Redis and this in-process cache have **not been directly benchmarked** here. Adding Redis today would introduce components that do not address the remaining observed template/output-buffer allocation costs. Redis's documentation itself discusses the latency advantage of keeping frequently used data in an application-local cache.

**Selection:** retain the startup catalog, concurrent loads for cache misses, deduplicated process-local decoded JSON cache, and a no-goroutine warm-cache path. Consider Redis, bounded eviction, or HTML caching only if deployment requirements and new measurements justify them.

## 7. Tests and reproducible commands

### Functional and race tests

```bash
go test ./... -count=1
go test -race ./... -count=1

go test ./services -run 'TestMultiCatalog' -v -count=1
go test ./tests -run '^TestMultiSiteSSRHTTP$' -v -count=1
```

The user-reported functional and race suites passed on 5 November 2026, including the multi-site Beego integration test. A passing `-race` run means **no race was detected under those test executions**, not that all possible schedules are proven race-free.

### Loader-only benchmarks

```bash
go test ./services -run '^$' \
  -bench 'BenchmarkContentLoad' -benchmem -count=5
```

### End-to-end loopback HTTP benchmarks

```bash
go test ./tests -run '^$' \
  -bench '^BenchmarkSSRHTTP$' -benchmem -benchtime=3s

go test ./tests -run '^$' \
  -bench '^BenchmarkSSRHTTP$' -benchmem -count=5

go test -race ./tests -run '^$' \
  -bench '^BenchmarkSSRHTTP/ParallelWarm$' -benchtime=100x
```

The HTTP benchmarks use `httptest.NewServer` with an actual loopback HTTP client and Beego handler. They include client **and** server work in the same Go process. `ColdWithInvalidation` includes a cache invalidation operation before **every** timed HTTP request; it is **not** a cold process-start benchmark. `Warm` primes the cache first. `ParallelWarm` reports aggregate operations/time rather than a per-visitor latency percentile. Browser layout, remote assets, and internet latency are excluded.

## 8. Measured performance (submitted 5 November 2026)

**Measurement environment:** Linux `amd64`; Intel Core i3-4130 @ 3.40 GHz; benchmarks displayed `-4` (GOMAXPROCS = 4). Values are from the user's submitted command output, not independently measured on other hardware.

### Decoded-content loader only

| Benchmark | First run | 5-run median |
| --- | ---: | ---: |
| Uncached concurrent JSON load | 454,484 ns/op | **439,985 ns/op** |
| Cached warm load | 193.7 ns/op | **200.0 ns/op** |

The five-run median comparison is approximately **2,200×** in favor of the warm cache **for this loader-only operation**. It is **not** the whole-page improvement.

### Full loopback HTTP processing

| Scenario | First run | Median of five subsequent runs | Five-run range |
| --- | ---: | ---: | ---: |
| Cold with invalidation | 1,028,646 ns/op | **845,968 ns/op** | 813,265–1,136,400 ns/op |
| Warm | 315,757 ns/op | **369,861 ns/op** | 344,092–593,251 ns/op |
| Parallel warm | 195,989 ns/op | **221,758 ns/op** | 212,743–295,980 ns/op |

Compared with `ColdWithInvalidation`, the **sequential warm median** is approximately **2.29× faster**, corresponding to **56.3% lower ns/op** for these defined benchmark workloads. Parallel warm `ns/op` represents aggregate throughput, **not** individual request latency.

In the initial HTTP benchmark, cold versus warm allocations were **280,773 versus 152,819 B/op** (about **45.6% less**) and **364 versus 267 allocs/op** (about **26.6% fewer**). The allocations reported by the HTTP benchmark include both client and server components.

### CPU and allocation profiling

The corrected **sequential-only warm** profile used:

```bash
go test ./tests -run '^$' \
  -bench '^BenchmarkSSRHTTP$/^Warm$' -benchtime=5s -benchmem \
  -cpuprofile=cpu_warm_sequential.out \
  -memprofile=mem_warm_sequential.out

go tool pprof -top cpu_warm_sequential.out
go tool pprof -top -cum cpu_warm_sequential.out
go tool pprof -alloc_space -top mem_warm_sequential.out
go tool pprof -alloc_objects -top mem_warm_sequential.out
```

This run measured **306,784 ns/op**, **150,711 B/op**, **266 allocs/op**, and **112.33 MB/s** of response-body throughput. In the total-allocated-bytes profile, `bytes.growSlice` accounted for approximately **90.60%** of sampled/estimated allocation volume. Template execution and Beego response writing contributed substantial **overlapping cumulative call paths**. These percentages represent allocations **over the profiling run**, not the application's resident memory or a direct percentage of HTTP response latency.

**Interpretation:** Warm content loading is not the observed main allocation source. Template evaluation and output buffering are candidates for further study, but the existing data do **not** prove that a different renderer or buffer policy is better. No production rendering rewrite was made on this evidence.

## 9. Known limitations and follow-up work

- **Deployment lifecycle:** multiple metadata files can be registered at startup, but no dynamic catalog updates, authenticated deployment callbacks, or coordinated invalidation across instances exist.
- **Cache lifecycle:** in-memory only, unbounded, without automatic TTL/eviction or deployment-driven refresh. Explicit invalidation can leave some already in-flight requests using old content.
- **Storage scope:** JSON is local; remote S3 retrieval and deployment-system integration have not been implemented.
- **Input and trust boundaries:** lexical relative-path validation is present but is not a full symlink-containment policy. Publisher-supplied HTML is rendered as trusted; do not feed untrusted visitor HTML into those fields.
- **Asset status:** some local `/static/img/...` URLs returned 404 and certain external image URLs returned 403 during browser testing. Root cause for external 403 has not been established. These are intentionally deferred; CSS/JS CDN links remain unchanged.
- **Benchmark scope:** loopback HTTP benchmarks are not browser page-load or production traffic measurements. End-to-end network/CDN timings, peak load tests, and latency percentiles have not been measured.
- **Hosting:** the latest submitted `main.go` binds to `127.0.0.1:8080` for local development. Public deployment configuration and reverse-proxy behavior are not covered here.

Future optimizations should start with **measured requirements**, then isolated tests that preserve the same HTML and multi-site routing behavior. Profiling indicates template/output-buffer allocation paths are logical candidates for an experiment, not an automatic rewrite.

## 10. Reference documentation

- [Go: `sync` package (`Mutex`, `WaitGroup`, `Map`)](https://pkg.go.dev/sync)
- [Go: diagnostics, profiling, and tracing](https://go.dev/doc/diagnostics)
- [Redis: client-side caching and network-round-trip trade-offs](https://redis.io/docs/latest/develop/clients/client-side-caching/)

---

**Project status:** URL resolution, server-side HTML, concurrent JSON loading and caching, multi-site startup catalog, multi-site HTTP tests, benchmarks, and profiling have been implemented/tested to the extent described above. A future feature is not represented as implemented solely because it appears in this README.