// Package flow implements the core logic for the wave-flow plugin.
package flow

import (
	"io"
	"os"
	"os/exec"
)

func buildEnv(extraEnv map[string]string) []string {
	env := os.Environ()
	for k, v := range extraEnv {
		env = append(env, k+"="+v)
	}
	return env
}

func newShellCmd(cmdStr string, extraEnv map[string]string, stdout, stderr io.Writer) *exec.Cmd {
	c := exec.Command("sh", "-c", cmdStr)
	c.Stdout = stdout
	c.Stderr = stderr
	c.Env = buildEnv(extraEnv)
	return c
}
