package types

// CodeSpec defines the input for code generation
type CodeSpec struct {
	Name       string
	UserCode   string
	Setup      string
	Teardown   string
	Iterations int
	Warmup     int
	TrackHeap  bool
	TrackDB    bool
}

// Result represents the output of a single benchmark run
type Result struct {
	Name          string   `json:"name"`
	Iterations    int      `json:"iterations"`
	AvgWallMs     float64  `json:"avgWallMs"`
	AvgCpuMs      float64  `json:"avgCpuMs"`
	MinWallMs     float64  `json:"minWallMs"`
	MaxWallMs     float64  `json:"maxWallMs"`
	MinCpuMs      float64  `json:"minCpuMs"`
	MaxCpuMs      float64  `json:"maxCpuMs"`
	AvgHeapKb     *float64 `json:"avgHeapKb,omitempty"`
	MinHeapKb     *float64 `json:"minHeapKb,omitempty"`
	MaxHeapKb     *float64 `json:"maxHeapKb,omitempty"`
	DmlStatements *int     `json:"dmlStatements,omitempty"`
	SoqlQueries   *int     `json:"soqlQueries,omitempty"`
}

// AggregatedResult combines multiple Results with statistics
type AggregatedResult struct {
	Name         string   `json:"name"`
	Runs         int      `json:"runs"`
	Iterations   int      `json:"iterations"`
	Warmup       int      `json:"warmup"`
	AvgCpuMs     float64  `json:"avgCpuMs"`
	StdDevCpuMs  float64  `json:"stdDevCpuMs"`
	MinCpuMs     float64  `json:"minCpuMs"`
	MaxCpuMs     float64  `json:"maxCpuMs"`
	AvgWallMs    float64  `json:"avgWallMs"`
	StdDevWallMs float64  `json:"stdDevWallMs"`
	MinWallMs    float64  `json:"minWallMs"`
	MaxWallMs    float64  `json:"maxWallMs"`
	RawResults   []Result `json:"raw,omitempty"`
}

// BenchmarkConfig represents configuration loaded from file
type BenchmarkConfig struct {
	Benchmarks []BenchmarkSpec `yaml:"benchmarks"`
	Iterations int             `yaml:"iterations"`
	Warmup     int             `yaml:"warmup"`
	Runs       int             `yaml:"runs"`
	Parallel   int             `yaml:"parallel"`
	TrackHeap  bool            `yaml:"trackHeap"`
	TrackDB    bool            `yaml:"trackDB"`
	Org        string          `yaml:"org"`
	Output     string          `yaml:"output"`
}

// BenchmarkSpec defines a single benchmark in config file
type BenchmarkSpec struct {
	Name     string `yaml:"name"`
	File     string `yaml:"file,omitempty"`
	Code     string `yaml:"code,omitempty"`
	Setup    string `yaml:"setup,omitempty"`
	Teardown string `yaml:"teardown,omitempty"`
}
