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
		// var response jsonrpc.Response
		// json.Unmarshal([]byte(), &response)
		t.Log(sErr.String())
		t.Fatalf("command run failed: %v", err)
	}

}

func SInvoke(t *testing.T, in string, out *strings.Builder, sErr *strings.Builder, env *[]string, ignoreErr bool) {
	cmd := exec.Command("zion")
	if env != nil {
		cmd.Env = *env
	}

	cmd.Stdin = strings.NewReader(in)
	cmd.Stdout = out
	cmd.Stderr = sErr
	err := cmd.Run()
	if !ignoreErr && err != nil {
		t.Fatalf("command run failed: %v", err)
	}

}
