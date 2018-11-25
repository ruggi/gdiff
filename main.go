package main

import (
	"fmt"

	"github.com/jroimartin/gocui"
	"github.com/ruggi/vgit/parser"
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

func layout(g *gocui.Gui) func(*gocui.Gui) error {
	diffs, err := parser.LoadDiffs()
	if err != nil {
		panic(err)
	}

	diff := diffs[0] // TODO files picker

	return func(g *gocui.Gui) error {
		w, h := g.Size()

		if v, err := g.SetView("left", 0, 0, w/2-1, h-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}

			fmt.Fprintf(v, "%s", diff.Left)

			v.Editable = false
			v.Wrap = true
		}
		if v, err := g.SetView("right", w/2, 0, w-1, h-1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}

			fmt.Fprintf(v, "%s", diff.Right)

			v.Editable = false
			v.Wrap = true
		}

		return nil
	}
}

var (
	y = 0
)

func moveCursor(direction int) func(*gocui.Gui, *gocui.View) error {
	return func(g *gocui.Gui, _ *gocui.View) error {
		y += direction
		if y < 0 {
			y = 0
		}
		// TODO prevent going after max height
		if v, err := g.View("left"); err == nil {
			v.SetOrigin(0, y)
		}
		if v, err := g.View("right"); err == nil {
			v.SetOrigin(0, y)
		}
		return nil
	}
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
	return nil
}
