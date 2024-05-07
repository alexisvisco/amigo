package testutils

import (
	"errors"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/require"
	"os"
	"path"
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

//
//func snapshotSavePgDump(db DatabaseCredentials, schema, file string) error {
//	args := []string{
//		"-d", db.DB,
//		"-h", db.Host,
//		"-U", db.User,
//		"-p", db.Port,
//		"-n", schema,
//		"-s",
//		"--no-comments",
//		"--no-owner",
//		"--no-privileges",
//		"--no-tablespaces",
//		"--no-unlogged-Table-data",
//		"--no-security-labels",
//		"--file", path.Join("testdata", file) + ".snap.sql",
//	}
//
//	if err := os.MkdirAll(path.Dir(path.Join("testdata", file)), 0755); err != nil {
//		return fmt.Errorf("unable to create directory: %w", err)
//	}
//
//	env := map[string]string{"PGPASSWORD": db.Pass}
//
//	// todo: add pg_dump to PATH
//	if _, _, err := cmdexec.Exec("/opt/homebrew/opt/libpq/bin/pg_dump", args, env); err != nil {
//		return err
//	}
//
//	return nil
//}
//
//func snapshotDiffPgDump(db DatabaseCredentials, schema, file string) error {
//	args := []string{
//		"-d", db.DB,
//		"-h", db.Host,
//		"-U", db.User,
//		"-p", db.Port,
//		"-n", schema,
//		"-s",
//		"--no-comments",
//		"--no-owner",
//		"--no-privileges",
//		"--no-tablespaces",
//		"--no-unlogged-Table-data",
//		"--no-security-labels",
//		"-f", path.Join("testdata", file) + ".out.sql",
//	}
//
//	env := map[string]string{"PGPASSWORD": db.Pass}
//
//	_, _, err := cmdexec.Exec("/opt/homebrew/opt/libpq/bin/pg_dump", args, env)
//	if err != nil {
//		return err
//	}
//
//	readFile, err := os.ReadFile(path.Join("testdata", file) + ".snap.sql")
//	if err != nil {
//		return fmt.Errorf("unable to read snap file: %w", err)
//	}
//
//	out, err := os.ReadFile(path.Join("testdata", file) + ".out.sql")
//	if err != nil {
//		return fmt.Errorf("unable to read out file: %w", err)
//	}
//
//	if string(readFile) != string(out) {
//		absOut, err := filepath.Abs(path.Join("testdata", file) + ".out.sql")
//		if err != nil {
//			return fmt.Errorf("unable to get absolute path: %w", err)
//		}
//
//		absSnap, err := filepath.Abs(path.Join("testdata", file) + ".snap.sql")
//		if err != nil {
//			return fmt.Errorf("unable to get absolute path: %w", err)
//		}
//
//		out, _, err := cmdexec.Exec("bash",
//			[]string{"-c", fmt.Sprintf("sdiff -l %s %s | cat -n | grep -v -e '($'", absSnap, absOut)}, nil)
//		if err != nil {
//			return fmt.Errorf("unable to diff files: %w", err)
//		}
//
//		return errors.New("snapshots are different, path: " + absOut + ":1\n" + out)
//	} else {
//		return nil
//	}
//}
