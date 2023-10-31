package names

// node plugins
const (
	// filter plugins
	// filter by node name
	NodeName = "NodeName"
	// filter by node name regex
	NodeRegex = "NodeRegex"
	// filter by cpu overcommit ratio
	CPUOvercommit = "CPUOvercommit"
	// filter by memory overcommit ratio
	MemoryOvercommit = "MemoryOvercommit"

	// score plugins
	// random score
	Random = "Random"
	// resource utilization score
	NodeResource = "NodeResource"

	// vmid plugins
	// select by range
	Range = "Range"
	// select by regex
	Regex = "Regex"
)
