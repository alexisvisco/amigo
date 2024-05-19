package testutils

import (
	"errors"
	"fmt"
	"github.com/alexisvisco/amigo/pkg/schema"
	"github.com/alexisvisco/amigo/pkg/utils/cmdexec"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// EnableSnapshot is a map to enable snapshot for a specific test.
// The key can be either:
// - the name of the test (t.Name())
// - "all" to enable all snapshots
var EnableSnapshot = map[string]struct{}{}

type TestingT interface {
	Name() string
	require.TestingT
	Logf(format string, args ...interface{})
}

func EnableSnapshotFor(t TestingT) {
	EnableSnapshot[t.Name()] = struct{}{}
}

func EnableSnapshotForAll() {
	EnableSnapshot["all"] = struct{}{}
}

// TestSnapshotSaveText is a helper function to save the output of a test as a snapshot.
// To activate the save on your top test add EnableSnapshot[t.Name()] = struct{}
// Or EnableSnapshot["all"] = struct{} to save all snapshots.
func maySnapshotSaveText(t TestingT, content string, enable ...bool) {
	continueSave := false
	if _, ok := EnableSnapshot["all"]; ok {
		continueSave = true
	}

	for k := range EnableSnapshot {
		if strings.HasPrefix(t.Name(), k) {
			continueSave = true
			break
		}
	}

	if len(enable) > 0 {
		continueSave = enable[0]
	}

	if !continueSave {
		return
	}

	f := path.Join("testdata", t.Name()) + ".snap.txt"

	err := os.MkdirAll(path.Dir(f), 0755)
	require.NoError(t, err)

	err = os.WriteFile(f, []byte(content), 0644)
	require.NoError(t, err)

	t.Logf("snapshot saved: %s", f)
}

func AssertSnapshotDiff(t TestingT, content string, save ...bool) {
	maySnapshotSaveText(t, content, save...)

	fSnap := path.Join("testdata", t.Name()) + ".snap.txt"
	err := os.MkdirAll(path.Dir(fSnap), 0755)
	require.NoError(t, err)

	f := path.Join("testdata", t.Name()) + ".out.txt"
	os.WriteFile(f, []byte(content), 0644)
	require.NoError(t, err)

	snap, err := os.ReadFile(fSnap)
	require.NoError(t, err)

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(string(snap), content, false)
	allEqual := true
	for _, d := range diffs {
		if d.Type != diffmatchpatch.DiffEqual {
			allEqual = false
		}
	}

	if !allEqual {
		err := errors.New("snapshots are different\n" + dmp.DiffPrettyText(diffs))
		require.NoError(t, err)
	}
}

func MaySnapshotSavePgDump(t TestingT, schemaName string, db schema.DatabaseCredentials, id string, enable ...bool) {
	continueSave := false
	if _, ok := EnableSnapshot["all"]; ok {
		continueSave = true
	}

	for k := range EnableSnapshot {
		if strings.HasPrefix(t.Name(), k) {
			continueSave = true
			break
		}
	}

	if len(enable) > 0 {
		continueSave = enable[0]
	}

	if !continueSave {
		return
	}

	file := path.Join("testdata", t.Name()+"_"+id) + ".snap.sql"

	args := []string{
		"-d", db.DB,
		"-h", db.Host,
		"-U", db.User,
		"-p", db.Port,
		"-n", schemaName,
		"-s",
		"--no-comments",
		"--no-owner",
		"--no-privileges",
		"--no-tablespaces",
		"--no-security-labels",
		"--file", file,
	}

	err := os.MkdirAll(path.Dir(file), 0755)
	require.NoError(t, err)

	env := map[string]string{"PGPASSWORD": db.Pass}

	_, _, err = cmdexec.Exec(getPgDumpPath(), args, env)
	require.NoError(t, err)

	return
}

func AssertSnapshotPgDumpDiff(t TestingT, schemaName string, db schema.DatabaseCredentials, id string, enable ...bool) {
	MaySnapshotSavePgDump(t, schemaName, db, id, enable...)
	fileOut := path.Join("testdata", t.Name()+"_"+id) + ".out.sql"
	fileSnap := path.Join("testdata", t.Name()+"_"+id) + ".snap.sql"

	args := []string{
		"-d", db.DB,
		"-h", db.Host,
		"-U", db.User,
		"-p", db.Port,
		"-n", schemaName,
		"-s",
		"--no-comments",
		"--no-owner",
		"--no-privileges",
		"--no-tablespaces",
		"--no-security-labels",
		"-f", fileOut,
	}

	env := map[string]string{"PGPASSWORD": db.Pass}

	_, _, err := cmdexec.Exec(getPgDumpPath(), args, env)
	require.NoError(t, err)

	snap, err := os.ReadFile(fileSnap)
	require.NoError(t, err)

	out, err := os.ReadFile(fileOut)
	require.NoError(t, err)

	if string(snap) != string(out) {
		absOut, err := filepath.Abs(fileOut)
		require.NoError(t, err)

		absSnap, err := filepath.Abs(fileSnap)
		require.NoError(t, err)

		out, _, err := cmdexec.Exec("bash",
			[]string{"-c", fmt.Sprintf("sdiff -l %s %s | cat -n | grep -v -e '($'", absSnap, absOut)}, nil)
		require.NoError(t, err)

		t.Errorf("snapshots are different between %s and %s:\n%s", fileSnap, fileOut, out)
	}
}

func getPgDumpPath() string {
	def := "/opt/homebrew/opt/libpq/bin/pg_dump"

	// If the PG_DUMP_PATH is set, use it
	if os.Getenv("PG_DUMP_PATH") != "" {
		def = os.Getenv("PG_DUMP_PATH")
	}

	return def
}
