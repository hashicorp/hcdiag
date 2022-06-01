package version

import (
	"fmt"
)

var (
	// Version is the main version number that is being run at the moment.
	//
	// Version must be of the format <MAJOR>.<MINOR>.<PATCH>, as described in the semantic versioning specification.
	Version = "0.4.0"

	// Prerelease is a pre-release marker for the version. If this is "" (empty string) then it means that
	//it is a final release. Otherwise, this is a pre-release such as "dev" (in development),
	// "beta", "rc1", etc.
	Prerelease = "dev"

	// Metadata is any additional (optional) information regarding the build, as described by the semantic
	// versioning specification.
	Metadata string

	// GitCommit is the commit associated with the build. It is set by the build process and should not require
	// human updates under normal circumstances.
	GitCommit string

	// BuildDate is the date/time when the build was created. It is set by the build process and should not require
	// human updates under normal circumstances.
	BuildDate string
)

// VersionInfo is a container for version information.
type VersionInfo struct {
	Version    string `json:"version,omitempty"`
	Prerelease string `json:"prerelease,omitempty"`
	Metadata   string `json:"build_metadata,omitempty"`
	BuildDate  string `json:"build_date,omitempty"`
	Revision   string `json:"revision,omitempty"`
}

// GetVersion produces a VersionInfo that includes fields set based on version package variables.
func GetVersion() VersionInfo {
	return VersionInfo{
		Version:    Version,
		Prerelease: Prerelease,
		Metadata:   Metadata,
		BuildDate:  BuildDate,
		Revision:   GitCommit,
	}
}

// VersionNumber produces a semantic version number from a VersionInfo object.
func (c VersionInfo) VersionNumber() string {
	version := c.Version

	if c.Prerelease != "" {
		version = fmt.Sprintf("%s-%s", version, c.Prerelease)
	}

	if c.Metadata != "" {
		version = fmt.Sprintf("%s+%s", version, c.Metadata)
	}

	return version
}

// FullVersionNumber produces a human-readable string representation of the VersionInfo object.
//
// In addition to a short slug that includes the product name (hcdiag) and the semantic version,
// the Revision will be included if the optional argument, `rev`, is true. Further, if a BuildDate
// is set, it is also included in the output.
func (c VersionInfo) FullVersionNumber(rev bool) string {
	const slug = "hcdiag v"
	versionString := slug
	versionString += c.VersionNumber()

	if rev && c.Revision != "" {
		versionString += fmt.Sprintf(" (%s)", c.Revision)
	}

	if c.BuildDate != "" {
		versionString += fmt.Sprintf(", built %s", c.BuildDate)
	}

	return versionString
}
