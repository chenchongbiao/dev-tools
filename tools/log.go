package tools

import (
	"log"

	"github.com/chenchongbiao/dev-tools/ios"
	"github.com/rivo/tview"
)

// 判断是cli还是tui，输出日志
func PrintLog(logInfo string, outCh <-chan string, errCh <-chan string, textView *tview.TextView) {
	if textView == nil {
		log.Println(logInfo)
	} else if logInfo == "" {
		ios.CommandOutput(outCh, errCh, textView)
	} else if textView != nil && logInfo != "" {
		textView.SetText(textView.GetText(false) + logInfo + "\n")
	}
}

// 判断是cli还是tui，输出 Fatal 日志
func FatalLog(logInfo string, outCh <-chan string, errCh <-chan string, textView *tview.TextView) {
	if textView != nil {
		log.Fatalln(logInfo)
	} else if logInfo == "" {
		ios.CommandOutput(outCh, errCh, textView)
	} else if textView != nil && logInfo != "" {
		textView.SetText(textView.GetText(false) + logInfo + "\n")
	}
}
