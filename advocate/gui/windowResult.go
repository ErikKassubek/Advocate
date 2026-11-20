package gui

import (
	"advocate/trace"
	"advocate/utils/flags"
	"bufio"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
)

type windowResults struct {
	window fyne.Window

	closeButton  *componentButton
	resultsViwer *componentResultViewer
	pathSelector *componentPathSelector

	path  string
	trace *trace.Trace
}

func (this *window) openWindowResults() {
	tv := windowResults{}

	tv.window = this.app.NewWindow("Results")

	tv.closeButton = createButton("Close", tv.window.Close)
	tv.pathSelector = createPathSelector("Trace", &tv.path, tv.updateResults, tv.window)

	tv.resultsViwer = createResultsViewer()

	content := container.NewBorder(
		tv.pathSelector.Container,
		tv.closeButton.Container,
		nil,
		nil,
		container.NewStack(tv.resultsViwer.Scroll),
	)

	tv.window.SetContent(content)

	tv.window.Resize(fyne.NewSize(1728, 972))

	if flags.ProgPath != "" {
		tv.pathSelector.setPath(flags.ProgPath)
		tv.updateResults(flags.ProgPath)
	}

	tv.window.Show()
}

func (this *windowResults) updateResults(dir string) {
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		name := d.Name()
		if strings.HasPrefix(name, "bug_") && strings.HasSuffix(name, ".md") {
			this.addBug(path)
		}

		return nil
	})

}

func (this *windowResults) addBug(file string) {
	name := filepath.Base(file)
	name = strings.TrimPrefix(name, "bug_")
	name = strings.TrimSuffix(name, ".md")

	id := this.resultsViwer.AddTab(name)

	this.resultsViwer.AddLine(id, file)
	this.resultsViwer.AddLine(id, "")

	f, err := os.Open(file)
	if err != nil {
		return
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		this.resultsViwer.AddLine(id, line)
	}

}
