package cmdexec

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func Exec(cmd string, args []string, env map[string]string) (string, string, error) {
	co := exec.Command(cmd, args...)

	for k, v := range env {
		co.Env = append(co.Env, k+"="+v)
	}

	co.Env = append(co.Env, os.Environ()...)

	addToPath := []string{"/opt/homebrew/opt/libpq/bin", "/usr/local/opt/libpq/bin"}
	for i, key := range co.Env {
		if strings.HasPrefix(key, "PATH=") {
			co.Env[i] = key + ":" + strings.Join(addToPath, ":")
			break
		}
	}

	bufferStdout := new(strings.Builder)
	bufferStderr := new(strings.Builder)

	co.Stdout = bufferStdout
	co.Stderr = bufferStderr
	err := co.Run()
	if err != nil {
		return bufferStdout.String(), bufferStderr.String(), fmt.Errorf("unable to execute command: %w", err)
	}

	return bufferStdout.String(), bufferStderr.String(), nil
}

func ExecToWriter(cmd string, args []string, env map[string]string, stdout io.Writer, stderr io.Writer) error {
	co := exec.Command(cmd, args...)

	for k, v := range env {
		co.Env = append(co.Env, k+"="+v)
	}

	co.Env = append(co.Env, os.Environ()...)

	addToPath := []string{"/opt/homebrew/opt/libpq/bin", "/usr/local/opt/libpq/bin"}
	for i, key := range co.Env {
		if strings.HasPrefix(key, "PATH=") {
			co.Env[i] = key + ":" + strings.Join(addToPath, ":")
			break
		}
	}

	co.Stdout = stdout
	co.Stderr = stderr
	err := co.Run()
	if err != nil {
		return fmt.Errorf("unable to execute command: %w", err)
	}

	return nil
}
