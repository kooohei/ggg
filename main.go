package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/marcusolsson/tui-go"
)

var (
	ui        tui.UI
	baseDir   string
	entry     *tui.Entry
	list      *tui.Table
	statusBar *tui.StatusBar
	gitPath   string
)

func isValidPath(param string) bool {
	_, err := os.Stat(param)
	return err == nil
}

func setBaseDir() (err error) {
	if len(os.Args) == 1 {
		baseDir = "/"
		return
	} else if len(os.Args) == 2 {
		if !isValidPath(os.Args[1]) {
			err = errors.New("ベースディレクトリが見つかりません。")
			return
		}
		baseDir = os.Args[1]
	} else {
		err = errors.New("usage: ggg [base directory]")
		return
	}
	return
}

func hasGit() (err error) {
	var p = ""
	p, err = exec.LookPath("git")
	if err != nil {
		return
	}
	gitPath = p
	return
}

func main() {
	err := hasGit()
	if err != nil {
		log.Fatalf("gitコマンドが見つかりませんでした。")
	}
	err = setBaseDir()

	if err = initLayout(); err != nil {
		log.Fatalf("レイアウトの初期化でエラーが発生しました。%s¥n", err)
	}

	attachEvent()

	onInit()
	if err = ui.Run(); err != nil {
		log.Fatalf("アプリケーション実行時エラー: %s¥n", err)
	}
}

func onInit() {
	setStatus("---")
}

func getRow(name, number, path string) (row *tui.Box) {
	pad := tui.NewPadder
	l := tui.NewLabel
	row = tui.NewHBox(
		pad(2, 0, l(name)),
		pad(2, 0, l(number)),
		pad(2, 0, l(path)),
	)
	return
}

func initLayout() (err error) {
	// set application title
	ttl := tui.NewLabel("ggg v1.0")
	pad1 := tui.NewPadder(1, 0, ttl)
	header := tui.NewHBox(pad1)
	header.SetBorder(true)
	header.SetSizePolicy(tui.Expanding, tui.Preferred)

	// part of build Seach Form Block
	lbl1 := tui.NewLabel("KEYWORD:")
	lbl1Padder := tui.NewPadder(2, 1, lbl1)

	entry = tui.NewEntry()
	entry.SetFocused(true)
	entry.OnSubmit(onSubmit)

	entryBorder := tui.NewHBox(entry)
	entryBorder.SetSizePolicy(tui.Expanding, tui.Preferred)
	entryBorder.SetBorder(true)
	formBox := tui.NewHBox(lbl1Padder, entryBorder)

	// part of build List for found files.

	list = tui.NewTable(1, 1)
	list.SetSizePolicy(tui.Expanding, tui.Expanding)

	list.AppendRow(getRow("File Name", "Line Number", "Path"))

	scrollBox := tui.NewScrollArea(list)
	scrollBox.SetSizePolicy(tui.Expanding, tui.Expanding)

	// part of status line at most bottom into app container.

	statusBar = tui.NewStatusBar("")
	statusBar.SetSizePolicy(tui.Expanding, tui.Preferred)
	txt := fmt.Sprintf("Press Esc 'Quit' / Walk from [%s]", baseDir)
	sLabel := tui.NewLabel(txt)
	statusBox := tui.NewVBox(sLabel, statusBar)

	// part of applicatio container.
	base := tui.NewVBox(
		header,
		formBox,
		scrollBox,
		statusBox,
	)

	// create tui.UI then start app loop.
	ui, err = tui.New(base)
	if err != nil {
	}
	return
}

func setStatus(mes string) {
	statusBar.SetText(mes)
}

func attachEvent() {
	setKeyBinding()
}

func setKeyBinding() {
	ui.SetKeybinding("Esc", func() {
		ui.Quit()
	})

}

func onSubmit(ent *tui.Entry) {
	parameter := ent.Text()
	if strings.Trim(parameter, " ") == "" {
		return
	}
	parameter = strings.Replace(parameter, "-n ", "", -1)
	cmdtxt := fmt.Sprintf("%s grep -n %s %s", gitPath, parameter, baseDir)

	result, err := exec.Command(cmdtxt).Output()
	if err != nil {
		setStatus(err.Error())
		return
	}
	lines := strings.Split(string(result), " ")
	setList(lines)
}

func setList(lines []string) {
	cnt := 0
	res := "Not Found"
	list.RemoveRows()
	for _, row := range lines {
		f, n, p, ignore := getColumn(row)
		if ignore {
			continue
		}
		list.AppendRow(getRow(f, n, p))
		cnt++
	}

	if cnt != 0 {
		res = fmt.Sprintf("Matched %d lines", cnt)
	}
	setStatus(res)
}

func getColumn(row string) (f, n, p string, ignore bool) {
	chunks := strings.Split(row, ":")
	if len(chunks) != 3 {
		ignore = true
		return
	}
	f = chunks[0]
	n = chunks[1]
	p = strings.Join(chunks[2:], "")

	return
}
