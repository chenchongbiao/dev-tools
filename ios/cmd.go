package ios

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
	"sync"

	"github.com/rivo/tview"
)

// 直接传入格式化后命令给sh执行
func Run(cmd string) error {
	return exec.Command("sh", "-c", cmd).Run()
}

// 执行命令并捕获输出，持续输入到 chan 中
func CommandExecutor(cmd string) (<-chan string, <-chan string) {
	outCh := make(chan string)
	errCh := make(chan string)

	go func() {
		defer close(outCh)
		defer close(errCh)

		cmd := exec.Command("sh", "-c", cmd)
		stdOutPipe, err := cmd.StdoutPipe()
		if err != nil {
			errCh <- fmt.Sprintf("Failed to open stdout pipe: %v", err)
			return
		}
		stdErrPipe, err := cmd.StderrPipe()
		if err != nil {
			errCh <- fmt.Sprintf("Failed to open stderr pipe: %v", err)
			return
		}

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdOutPipe)
			for scanner.Scan() {
				outCh <- scanner.Text()
			}
			if err := scanner.Err(); err != nil && err != io.EOF {
				errCh <- fmt.Sprintf("Stderr scanner error: %v", err)
			}
		}()

		go func() {
			defer wg.Done()
			scanner := bufio.NewScanner(stdErrPipe)
			for scanner.Scan() {
				outCh <- scanner.Text()
			}
			if err := scanner.Err(); err != nil && err != io.EOF {
				errCh <- fmt.Sprintf("Stderr scanner error: %v", err)
			}
		}()

		if err := cmd.Start(); err != nil {
			errCh <- fmt.Sprintf("Failed to start command: %v", err)
			return
		}

		wg.Wait()

		if err := cmd.Wait(); err != nil {
			errCh <- fmt.Sprintf("Command failed: %v", err)
		}
	}()

	return outCh, errCh
}

// 从 chan 中打印命令过程的输出，根据传入的参数执行输出到命令行或者 tview.TextView
func CommandOutput(outCh <-chan string, errCh <-chan string, textView *tview.TextView) {
	if outCh == nil && errCh == nil {
		return
	}

	if textView != nil {
		go func() {
			for {
				select {
				case out, ok := <-outCh:
					if !ok {
						outCh = nil
					} else {
						textView.SetText(textView.GetText(false) + out + "\n")
					}
				case err, ok := <-errCh:
					if !ok {
						errCh = nil
					} else {
						textView.SetText(textView.GetText(false) + err + "\n")
					}
				}
				if outCh == nil && errCh == nil {
					break
				}
			}
		}()
	} else {
		for {
			select {
			case out, ok := <-outCh:
				if !ok {
					outCh = nil
				} else {
					log.Println(out)
				}
			case err, ok := <-errCh:
				if !ok {
					errCh = nil
				} else {
					log.Println(err)
				}
			}

			if outCh == nil && errCh == nil {
				break
			}
		}
	}
}

// 运行命令，获取命令执行结果
func RunCommandOutResult(command string) string {
	cmd := exec.Command("sh", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	// 检查命令是否成功执行
	if err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}
