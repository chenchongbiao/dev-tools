package layout

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	app               *tview.Application
	stage             *tview.Flex
	dpBuildConfigPage *DpBuildConfigPage // 配置构建的页面
	leftMenuPage      *tview.Form        // 左侧菜单页面
)

func DpBuildLayout() {
	app = tview.NewApplication().EnableMouse(true)

	grid := tview.NewGrid().
		SetRows(0).       // 第一行为0高度, 根据内容自适应大小
		SetColumns(26, 0) // 设置了两列，第一列宽度为26个单元格，第二列宽度为0
	grid.SetBorder(true)
	grid.SetBorderColor(tcell.ColorDodgerBlue)

	// 添加菜单
	leftMenuPage = newLeftMenuPage()
	grid.AddItem(leftMenuPage, 0, 0, 1, 1, 0, 0, true)

	dpBuildConfigPage = NewDpBuildConfigPage()
	stage = tview.NewFlex()
	setStage(dpBuildConfigPage.GetMainPage(), false)
	grid.AddItem(stage, 0, 1, 1, 1, 0, 0, false)

	if err := app.SetRoot(grid, true).SetFocus(grid).Run(); err != nil {
		panic(err)
	}
}

// 设置界面
func setStage(tp tview.Primitive, focus bool) {
	stage.Clear()
	stage.AddItem(tp, 0, 1, focus)
}

// 左侧菜单
func newLeftMenuPage() *tview.Form {
	menu := tview.NewForm()
	menu.SetBorder(true)
	menu.SetFieldBackgroundColor(tcell.ColorBlack)
	menu.SetFieldTextColor(tcell.ColorBlack)
	menu.AddButton("build config", func() {
		if dpBuildConfigPage == nil {
			dpBuildConfigPage = NewDpBuildConfigPage()
		}

		setStage(dpBuildConfigPage.GetMainPage(), false)
		app.SetFocus(dpBuildConfigPage.form)
	})
	return menu
}
