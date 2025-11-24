package generator

const apexTemplate = `// Apex Benchmark - Generated Code
// Benchmark: {{.Name}}
// Iterations: {{.Iterations}}
// Warmup: {{.Warmup}}

{{if .Setup}}
// Setup code
{{.Setup}}
{{end}}

Integer warmupIterations = {{.Warmup}};
Integer measurementIterations = {{.Iterations}};

// Warmup phase - JIT optimization
for (Integer i = 0; i < warmupIterations; i++) {
    {{.UserCode}}
}

// Measurement phase
Long totalWallTime = 0;
Long totalCpuTime = 0;
Long minWallTime = Long.MAX_VALUE;
Long maxWallTime = 0;
Integer minCpuTime = Integer.MAX_VALUE;
Integer maxCpuTime = 0;

{{if .TrackHeap}}
Long totalHeapUsed = 0;
Long minHeapUsed = Long.MAX_VALUE;
Long maxHeapUsed = 0;
{{end}}

{{if .TrackDB}}
Integer dmlStatementsBefore = Limits.getDmlStatements();
Integer soqlQueriesBefore = Limits.getQueries();
{{end}}

for (Integer i = 0; i < measurementIterations; i++) {
    {{if .TrackHeap}}
    Long heapBefore = Limits.getHeapSize();
    {{end}}

    Long wallStart = System.now().getTime();
    Integer cpuStart = Limits.getCpuTime();

    {{.UserCode}}

    Long wallEnd = System.now().getTime();
    Integer cpuEnd = Limits.getCpuTime();

    {{if .TrackHeap}}
    Long heapAfter = Limits.getHeapSize();
    Long heapDelta = heapAfter - heapBefore;
    totalHeapUsed += heapDelta;
    if (heapDelta < minHeapUsed) minHeapUsed = heapDelta;
    if (heapDelta > maxHeapUsed) maxHeapUsed = heapDelta;
    {{end}}

    Long wallDelta = wallEnd - wallStart;
    Integer cpuDelta = cpuEnd - cpuStart;

    totalWallTime += wallDelta;
    totalCpuTime += cpuDelta;

    if (wallDelta < minWallTime) minWallTime = wallDelta;
    if (wallDelta > maxWallTime) maxWallTime = wallDelta;
    if (cpuDelta < minCpuTime) minCpuTime = cpuDelta;
    if (cpuDelta > maxCpuTime) maxCpuTime = cpuDelta;
}

{{if .TrackDB}}
Integer dmlStatementsAfter = Limits.getDmlStatements();
Integer soqlQueriesAfter = Limits.getQueries();
Integer dmlStatementsDelta = dmlStatementsAfter - dmlStatementsBefore;
Integer soqlQueriesDelta = soqlQueriesAfter - soqlQueriesBefore;
{{end}}

{{if .Teardown}}
// Teardown code
{{.Teardown}}
{{end}}

// Calculate averages (convert to milliseconds with decimals)
Decimal avgWallMs = Decimal.valueOf(totalWallTime) / measurementIterations;
Decimal avgCpuMs = Decimal.valueOf(totalCpuTime) / measurementIterations;
Decimal minWallMs = Decimal.valueOf(minWallTime);
Decimal maxWallMs = Decimal.valueOf(maxWallTime);
Decimal minCpuMs = Decimal.valueOf(minCpuTime);
Decimal maxCpuMs = Decimal.valueOf(maxCpuTime);

{{if .TrackHeap}}
Decimal avgHeapKb = Decimal.valueOf(totalHeapUsed) / measurementIterations / 1024;
Decimal minHeapKb = Decimal.valueOf(minHeapUsed) / 1024;
Decimal maxHeapKb = Decimal.valueOf(maxHeapUsed) / 1024;
{{end}}

// Build result JSON
String resultJson = '{' +
    '"name":"{{.Name}}",' +
    '"iterations":' + measurementIterations + ',' +
    '"avgWallMs":' + avgWallMs.format() + ',' +
    '"avgCpuMs":' + avgCpuMs.format() + ',' +
    '"minWallMs":' + minWallMs.format() + ',' +
    '"maxWallMs":' + maxWallMs.format() + ',' +
    '"minCpuMs":' + minCpuMs.format() + ',' +
    '"maxCpuMs":' + maxCpuMs.format() +
    {{if .TrackHeap}}
    ',"avgHeapKb":' + avgHeapKb.format() +
    ',"minHeapKb":' + minHeapKb.format() +
    ',"maxHeapKb":' + maxHeapKb.format() +
    {{end}}
    {{if .TrackDB}}
    ',"dmlStatements":' + dmlStatementsDelta +
    ',"soqlQueries":' + soqlQueriesDelta +
    {{end}}
    '}';

// Output result with marker for parsing
System.debug('BENCH_RESULT:' + resultJson);
`
