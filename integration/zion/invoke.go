package zion

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func Invoke(t *testing.T, in []byte, out *strings.Builder, sErr *strings.Builder, env *[]string) {
	cmd := exec.Command("zion")
	if env != nil {
		cmd.Env = *env
	}

	cmd.Stdin = bytes.NewReader(in)
	cmd.Stdout = out
	cmd.Stderr = sErr
	err := cmd.Run()
	if err != nil {
		t.Fatalf("command run failed: %v", err)
	}
}
