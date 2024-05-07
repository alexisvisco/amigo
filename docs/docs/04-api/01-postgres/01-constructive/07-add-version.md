# Add version

The `AddVersion` function allows you to add a version to the database schema. 

A version is a unique identifier that represents the state of the schema at a particular point in time. The function accepts the version name and optional version options.

#### Basic Usage

The function follows this format:

```go
p.AddVersion(versionName)
```

