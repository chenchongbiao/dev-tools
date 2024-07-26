package tools

import (
	"log"

	"github.com/chenchongbiao/dev-tools/ios"
	"github.com/rivo/tview"
)

// 判断是cli还是tui，输出日志
func PrintLog(logInfo string, outCh <-chan string, errCh <-chan string, textView *tview.TextView) {
	if outCh != nil || errCh != nil {
		ios.CommandOutput(outCh, errCh, textView)
	}

	if textView == nil && logInfo != "" {
		log.Println(logInfo)
	}

	if textView != nil && logInfo != "" {
		textView.SetText(textView.GetText(false) + logInfo + "\n")
	}
}

// 判断是cli还是tui，输出 Fatal 日志
func FatalLog(logInfo string, outCh <-chan string, errCh <-chan string, textView *tview.TextView) {
	if outCh != nil || errCh != nil {
		ios.CommandOutput(outCh, errCh, textView)
	}

	if textView == nil && logInfo != "" {
		log.Fatalf(logInfo)
	}

	if textView != nil && logInfo != "" {
		textView.SetText(textView.GetText(false) + logInfo + "\n")
	}
}
