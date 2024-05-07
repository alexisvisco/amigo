package schema

type Schema interface {
	TableExist(tableName TableName) bool
	AddVersion(version string)
	RemoveVersion(version string)
	FindAppliedVersions() []string
}
