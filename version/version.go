package version

var (
	Version = "0.19.0"
	// PriorVersion - prior version
	PriorVersion = "0.18.0"
)

const (
	// LatestVersion product version supported
	LatestVersion        = "2.16.0"
	CompactLatestVersion = "2160"
	// LastMicroVersion product version supported
	LastMicroVersion = "2.15.0"
	// LastMinorVersion product version supported
	LastMinorVersion = "2.13.0"
)

// SupportedVersions - product versions this operator supports
var SupportedVersions = []string{LatestVersion, LastMicroVersion, LastMinorVersion}
var SupportedMicroVersions = []string{LatestVersion, LastMicroVersion}
var OperandVersionFromOperatorVersion map[string]string = map[string]string{
	"0.17.0": "2.13.0",
	"0.18.0": "2.15.0",
	"0.19.0": "2.16.0",
}
var FullVersionFromMinorVersion map[string]string = map[string]string{
	"130": "2.13.0",
	"150": "2.15.0",
	"160": "2.16.0",
}

var CompactFullVersionFromMinorVersion map[string]string = map[string]string{
	"130": "2130",
	"150": "2150",
	"160": "2160",
}

var CompactVersionFromVersion map[string]string = map[string]string{
	"2.13.0": "2130",
	"2.15.0": "2150",
	"2.16.0": "2160",
}

var FullVersionFromCompactVersion map[string]string = map[string]string{
	"2130": "2.13.0",
	"2150": "2.15.0",
	"2160": "2.16.0",
}

var MinorVersionFromFullVersion map[string]string = map[string]string{
	"2.13.0": "130",
	"2.15.0": "150",
	"2.16.0": "160",
}
