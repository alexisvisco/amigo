package schema

// Schema is the interface that need to be implemented to support migrations.
type Schema interface {
	AddVersion(version string)
	RemoveVersion(version string)
	FindAppliedVersions() []string
}
