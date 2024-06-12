package tools

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/AlecAivazis/survey/v2/terminal"
	"github.com/spf13/cobra"
)

// 标记标记正在处理引发的错误
type FlagError struct {
	Err error
}

func (fe FlagError) Error() string {
	return fe.Err.Error()
}

func (fe FlagError) Unwrap() error {
	return fe.Err
}

// 没有任何错误消息的情况下触发退出代码1的错误
var SilentError = errors.New("SilentError")

// 用户启动的取消
var CancelError = errors.New("CancelError")

// 判断用户取消
func IsUserCancellation(err error) bool {
	return errors.Is(err, CancelError) || errors.Is(err, terminal.InterruptErr)
}

func PrintError(out io.Writer, err error, cmd *cobra.Command, debug bool) {
	var dnsError *net.DNSError

	if errors.As(err, &dnsError) {
		fmt.Fprintf(out, "error connecting to %s\n", dnsError.Name)

		if debug {
			fmt.Fprintln(out, dnsError)
		}

		return
	}

	fmt.Fprintln(out, err)

	var flagError *FlagError
	if errors.As(err, &flagError) || strings.HasPrefix(err.Error(), "unknown command ") {
		if !strings.HasSuffix(err.Error(), "\n") {
			fmt.Fprintln(out)
		}

		fmt.Fprintln(out, cmd.UsageString())
	}
}
