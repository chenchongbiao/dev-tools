package layout

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type DpBuildConfigPage struct {
	mainGrid *tview.Grid
	form     *tview.Form
	results  *tview.TextView
}

// 返回配置构建的页面
func NewDpBuildConfigPage() *DpBuildConfigPage {
	dpBuildConfigPage := &DpBuildConfigPage{
		mainGrid: tview.NewGrid(),
		form:     tview.NewForm(),
		results:  tview.NewTextView(),
	}

	dpBuildConfigPage.initUI()
	return dpBuildConfigPage
}

func (page *DpBuildConfigPage) GetMainPage() tview.Primitive {
	return page.mainGrid
}

func (page *DpBuildConfigPage) initUI() {
	page.form.SetItemPadding(1)
	page.form.SetBorder(true)
	page.form.SetTitle("build config")
	page.form.SetTitleAlign(tview.AlignLeft)
	page.form.SetLabelColor(tcell.ColorYellow)

	page.form.AddDropDown("Distro Version", []string{"beige"}, 0, nil)
	page.form.AddDropDown("Arch", []string{"amd64", "arm64", "riscv64", "loong64", "i386"}, 0, nil)
	page.form.AddDropDown("Build Target", []string{"rootfs", "WSL", "board"}, 0, func(option string, optionIndex int) {

		// if option == "rootfs" {

		// }
	})

	page.form.AddButton("build", func() {

	})

	page.results.SetBorder(true)
	page.results.SetDynamicColors(true)  // 允许动态颜色
	page.results.SetRegions(true)        //允许你为文本的特定部分定义可交互的区域
	page.results.SetChangedFunc(func() { // 设置了一个回调函数，当TextView的内容发生改变时，该函数会被触发
		app.Draw()
	})

	page.mainGrid.SetRows(25, 0)
	page.mainGrid.SetColumns(0)
	page.mainGrid.AddItem(page.form, 0, 0, 1, 1, 0, 0, false)
	page.mainGrid.AddItem(page.results, 1, 0, 1, 1, 0, 0, false)
}
