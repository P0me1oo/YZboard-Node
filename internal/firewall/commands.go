package firewall

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type commands interface {
	Available(string) bool
	Run(context.Context, string, string, ...string) (string, error)
}

type systemCommands struct{}

func (systemCommands) Available(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

type limitedOutput struct{ bytes.Buffer }

func (b *limitedOutput) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 4<<20 {
		return 0, fmt.Errorf("防火墙命令输出超过限制")
	}
	return b.Buffer.Write(p)
}

func (systemCommands) Run(parent context.Context, name, input string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(parent, 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), "LC_ALL=C", "LANG=C")
	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}
	var output limitedOutput
	cmd.Stdout, cmd.Stderr = &output, &output
	err := cmd.Run()
	if ctx.Err() != nil {
		err = ctx.Err()
	}
	if err != nil {
		// 不把已有防火墙规则或管理员的规则注释复制到通用日志。
		return output.String(), fmt.Errorf("%s 执行失败: %w", name, err)
	}
	return output.String(), nil
}
