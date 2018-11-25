package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/ruggi/gdiff/parser"
)

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		panic(err)
	}
	defer g.Close()

	g.Cursor = true

	g.SetManagerFunc(layout(g))

	if err := keybindings(g); err != nil {
		panic(err)
	}
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		panic(err)
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

var (
	curDiff = 0
	diffs   []*parser.Diff
	cursorY = 0
)

func layout(g *gocui.Gui) func(*gocui.Gui) error {
	var err error
	diffs, err = parser.LoadDiffs()
	if err != nil {
		panic(err) // TODO move this away
	}

	return func(g *gocui.Gui) error {
		w, h := g.Size()

		if v, err := g.SetView("left", 0, 0, w/2-1, h-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}

			v.Editable = false
			v.Wrap = true

			fmt.Fprintf(v, "%s", diffs[curDiff].Left)
		}
		if v, err := g.SetView("right", w/2, 0, w-1, h-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}

			v.Editable = false
			v.Wrap = true

			fmt.Fprintf(v, "%s", diffs[curDiff].Right)
		}

		return nil
	}
}

func moveCursor(direction int) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, _ *gocui.View) error {
		cursorY += direction
		if cursorY < 0 {
			cursorY = 0
		}
		// TODO prevent going after max height
		if v, err := g.View("left"); err == nil {
			v.SetOrigin(0, cursorY)
		}
		if v, err := g.View("right"); err == nil {
			v.SetOrigin(0, cursorY)
		}
		return nil
	}
}

func nextFile(g *gocui.Gui, v *gocui.View) error {
	curDiff++
	if curDiff >= len(diffs) {
		curDiff = 0
	}
	g.Update(func(g *gocui.Gui) error {
		if v, err := g.View("left"); err == nil {
			v.Clear()
			fmt.Fprintf(v, "%s", diffs[curDiff].Left)
		}
		if v, err := g.View("right"); err == nil {
			v.Clear()
			fmt.Fprintf(v, "%s", diffs[curDiff].Right)
		}
		return nil
	}) // TODO this SUCKS. it's just for experimenting.
	return nil
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowDown, gocui.ModNone, moveCursor(1)); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowUp, gocui.ModNone, moveCursor(-1)); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextFile); err != nil {
		return err
	}
	return nil
}
