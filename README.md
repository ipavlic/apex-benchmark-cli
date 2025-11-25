[![Test and Coverage](https://github.com/ipavlic/apex-benchmark-cli/actions/workflows/test.yml/badge.svg)](https://github.com/ipavlic/apex-benchmark-cli/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/ipavlic/apex-benchmark-cli/graph/badge.svg?token=1VTMOZE5FF)](https://codecov.io/gh/ipavlic/apex-benchmark-cli)
[![Go Report Card](https://goreportcard.com/badge/github.com/ipavlic/apex-benchmark-cli)](https://goreportcard.com/report/github.com/ipavlic/peak)


# Apex Benchmark CLI

Benchmark Salesforce Apex code without deployment. Get performance metrics instantly via Salesforce CLI.

## Installation

**Requirements:** [Salesforce CLI](https://developer.salesforce.com/tools/salesforcecli) installed and authenticated

```bash
go install github.com/ipavlic/apex-benchmark-cli/cmd/apex-bench@latest
```

## Quick Start

**Benchmark code:**
```bash
apex-bench run --code "String s = 'Hello' + ' ' + 'World';"
```

**Compare approaches:**
```bash
apex-bench compare \
  --bench "Plus:String s = 'a' + 'b' + 'c';" \
  --bench "Format:String.format('{0}{1}{2}', new List<String>{'a','b','c'})" \
  --output table
```

**Output:**
```
┌────────┬──────────┬──────────┬──────────┬──────────┐
│  NAME  │ AVG CPU  │ MIN CPU  │ MAX CPU  │ RELATIVE │
├────────┼──────────┼──────────┼──────────┼──────────┤
│ Plus   │ 0.180 ms │ 0.170 ms │ 0.190 ms │ 1.00x ⭐ │
│ Format │ 0.350 ms │ 0.340 ms │ 0.360 ms │ 1.94x    │
└────────┴──────────┴──────────┴──────────┴──────────┘
```

## Usage

### `run` - Single benchmark

```bash
apex-bench run [--code "..." | --file path.apex] [flags]
```

**Flags:**
- `--iterations <n>` - Measurement iterations (default: 100)
- `--warmup <n>` - Warmup iterations (default: 10)
- `--runs <n>` - Complete runs for statistics (default: 1)
- `--parallel <n>` - Max concurrent `sf apex run` executions (default: 1)
  - When `--runs > 1`, executes multiple runs simultaneously for faster results
  - Example: `--runs 10 --parallel 3` runs 10 benchmarks, 3 at a time
  - Start with 3-5 to avoid overwhelming your org's API limits
- `--output json|table` - Output format (default: json)
- `--track-heap` - Track heap usage
- `--track-db` - Track DML/SOQL

**Examples:**
```bash
# From file with multiple runs
apex-bench run --file query.apex --runs 10 --parallel 3 --output table

# Track database operations
apex-bench run --code "[SELECT Id FROM Account LIMIT 1]" --track-db
```

### `compare` - Compare multiple approaches

```bash
apex-bench compare --bench "Name:code" --bench "Name:file.apex" [flags]
```

All `run` flags are supported.

**Example:**
```bash
apex-bench compare \
  --bench "Map:Map<Id,Account> m = new Map<Id,Account>([SELECT Id FROM Account]);" \
  --bench "List:List<Account> a = [SELECT Id FROM Account];" \
  --iterations 200 --runs 5
```

## Output

**JSON** (default):
```json
{
  "name": "Benchmark",
  "runs": 5,
  "iterations": 100,
  "avgCpuMs": 0.182,
  "stdDevCpuMs": 0.008,
  "minCpuMs": 0.170,
  "maxCpuMs": 0.195
}
```

**Table** - formatted output with relative performance in compare mode.

## How It Works

1. Wraps your code in measurement logic (warmup + timed iterations)
2. Executes via `sf apex run`
3. Extracts metrics from debug logs
4. Aggregates multiple runs with statistics

## Best Practices

- **Use CPU time** for stable comparisons (not wall time)
- **Warmup** helps stabilize JIT-optimized code
- **Multiple runs** (`--runs 5-10`) provide statistical confidence
- **Parallel wisely** to avoid API limits (start with `--parallel 3`)

## Troubleshooting

| Issue | Fix |
|-------|-----|
| `sf: command not found` | Install [Salesforce CLI](https://developer.salesforce.com/tools/salesforcecli) |
| No org authenticated | Run `sf org login web` |
| High variability | Increase warmup (`--warmup 100`) and runs (`--runs 10`) |

## License

MIT
