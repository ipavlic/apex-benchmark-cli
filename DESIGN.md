# Apex Benchmark CLI - Design

## Overview

Go-based CLI for benchmarking Apex code snippets without deployment. Generates self-contained code with measurement logic, executes via `sf apex run`, and reports results.

## Key Decisions

### Self-Contained Measurement
Inline all timing logic in generated Apex code. No dependencies, no deployment required, works in any org.

### Language: Go
- Single binary distribution
- Built-in concurrency for parallel execution
- Fast startup, excellent CLI libraries
- Cross-platform compilation

### Separate from apex-benchmark
Different languages, deployment models, and use cases warrant separate project

## Architecture

**Flow:** CLI Args → Generator → Executor → Parser → Aggregator → Reporter

**Components:**
- **CLI** (`cmd/`) - Cobra commands, flag parsing
- **Generator** (`pkg/generator/`) - Template-based Apex code generation
- **Executor** (`pkg/executor/`) - Spawns `sf apex run`, parallel execution with goroutines
- **Parser** (`pkg/parser/`) - Extracts `BENCH_RESULT:<json>` from debug logs
- **Aggregator** (`pkg/stats/`) - Calculates mean, min, max, stddev
- **Reporter** (`pkg/reporter/`) - JSON/table output, comparison mode

## Generated Code Pattern

The tool generates self-contained Apex that:
1. Runs warmup iterations (not measured)
2. Measures wall time, CPU time per iteration
3. Optionally tracks heap usage and DB operations
4. Outputs `BENCH_RESULT:<json>` for parsing

Example output marker:
```apex
System.debug('BENCH_RESULT:' + JSON.serialize(new Map<String,Object>{
    'name' => 'MyBenchmark',
    'iterations' => 100,
    'avgCpuMs' => 0.25,
    'minCpuMs' => 0.20,
    'maxCpuMs' => 0.30
}));
```

## CLI Usage

```bash
# Basic benchmark
apex-bench run --code "String s = 'a' + 'b';"

# From file with options
apex-bench run --file snippet.apex --iterations 500 --runs 10 --parallel 3

# Compare multiple
apex-bench compare --bench "Plus:plus.apex" --bench "Format:format.apex"
```

**Key Flags:**
- `--code` / `--file` - Code to benchmark
- `--iterations` - Measurement iterations (default: 100)
- `--runs` - Number of runs for aggregation (default: 1)
- `--parallel` - Concurrent executions (default: 1)
- `--output` - Format: `json` or `table` (default: json)
- `--track-heap` / `--track-db` - Optional metrics

## Project Structure

```
apex-benchmark-cli/
├── cmd/apex-bench/      # CLI entry point (cobra)
├── pkg/
│   ├── generator/       # Apex code generation
│   ├── executor/        # sf apex run execution
│   ├── parser/          # Result extraction
│   ├── stats/           # Statistical aggregation
│   ├── reporter/        # Output formatting
│   └── types/           # Shared data structures
├── testdata/            # Example snippets and configs
└── Makefile             # Build tasks
```

## Data Structures

**CodeSpec** - Input to generator (user code, iterations, warmup, tracking options)
**Result** - Single run output (avg/min/max for CPU, wall time, heap, DB)
**AggregatedResult** - Multi-run stats (mean, stddev, raw results)

## Dependencies

- `cobra` - CLI framework
- `viper` - Config file handling
- `tablewriter` - Table output
- `golang.org/x/sync` - Semaphore for rate limiting

## Build & Distribution

```bash
make build        # Current platform
make build-all    # Cross-compile
make test         # Run tests
```

Distribution via GitHub Releases (Linux, macOS, Windows binaries) and `go install`
