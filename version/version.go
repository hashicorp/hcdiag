package version

import (
	"fmt"
)

const (
	slug = "hcdiag v"
)

var (
	// version is the main version number that is being run at the moment.
	//
	// version must be of the format <MAJOR>.<MINOR>.<PATCH>, as described in the semantic versioning specification.
	version = "0.3.0"

	// prerelease is a pre-release marker for the version. If this is "" (empty string) then it means that
	// it is a final release. Otherwise, this is a pre-release such as "dev" (in development),
	// "beta", "rc1", etc.
	prerelease = ""

	// metadata is any additional (optional) information regarding the build, as described by the semantic
	// versioning specification.
	metadata string

	// gitCommit is the commit associated with the build. It is set by the build process and should not require
	// human updates under normal circumstances.
	gitCommit string

	// buildDate is the date/time when the build was created. It is set by the build process and should not require
	// human updates under normal circumstances.
	buildDate string
)

// Version is a container for version information.
type Version struct {
	Version    string `json:"version,omitempty"`
	Prerelease string `json:"prerelease,omitempty"`
	Metadata   string `json:"build_metadata,omitempty"`
	Revision   string `json:"revision,omitempty"`
	BuildDate  string `json:"build_date,omitempty"`
}

// GetVersion produces a Version that includes fields set based on version package variables.
func GetVersion() Version {
	return Version{
		Version:    version,
		Prerelease: prerelease,
		Metadata:   metadata,
		Revision:   gitCommit,
		BuildDate:  buildDate,
	}
}

// SemanticVersion produces a semantic version number from a Version object.
func (v Version) SemanticVersion() string {
	sv := v.Version

	if v.Prerelease != "" {
		sv = fmt.Sprintf("%s-%s", sv, v.Prerelease)
	}

	if v.Metadata != "" {
		sv = fmt.Sprintf("%s+%s", sv, v.Metadata)
	}

	return sv
}

// FullVersionNumber produces a human-readable string representation of the Version object.
//
// In addition to a short slug that includes the product name (hcdiag) and the semantic version,
// the Revision will be included if the optional argument, `rev`, is true. Further, if a buildDate
// is set, it is also included in the output.
func (v Version) FullVersionNumber(rev bool) string {
	versionString := slug
	versionString += v.SemanticVersion()

	if rev && v.Revision != "" {
		versionString += fmt.Sprintf(" (%s)", v.Revision)
	}

	if v.BuildDate != "" {
		versionString += fmt.Sprintf(", built %s", v.BuildDate)
	}

	return versionString
}
