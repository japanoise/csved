package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/japanoise/termbox-util"
	"github.com/nsf/termbox-go"
)

var (
	cbuf    *buffer
	message string
	debug   bool
)

func refresh(sx, sy int) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	// update scroll
	if cbuf.xsel < cbuf.xoffset {
		cbuf.xoffset = cbuf.xsel
	} else if cbuf.xsel > 0 {
		x := 0
		for xoff, col := range cbuf.cols[cbuf.xoffset:] {
			if xoff+cbuf.xoffset == cbuf.xsel {
				if x > sx {
					cbuf.xoffset = cbuf.xsel
				}
				break
			}
			x += col.maxWidth + 1
		}
	}

	if cbuf.ysel < cbuf.yoffset {
		cbuf.yoffset = cbuf.ysel
	} else if cbuf.ysel > (sy-3)+cbuf.yoffset {
		cbuf.yoffset = cbuf.ysel
	}
	if cbuf.ysel != 0 && cbuf.titles && cbuf.ysel == cbuf.yoffset {
		cbuf.yoffset = cbuf.ysel - 1
	}

	// row data
	for y := 0; y < sy-2 && (y+cbuf.yoffset) < cbuf.nrows; y++ {
		x := 0
		for xoff, col := range cbuf.cols[cbuf.xoffset:] {
			if cbuf.titles && y == 0 {
				if cbuf.ysel == 0 && xoff+cbuf.xoffset == cbuf.xsel {
					for i := 0; i < col.maxWidth; i++ {
						termutil.PrintRune(x+i, y, ' ', termbox.AttrReverse|termbox.ColorRed)
					}
					termutil.PrintstringColored(termbox.AttrReverse|termbox.ColorRed, col.data[0], x, y)
				} else {
					termutil.PrintstringColored(termbox.AttrReverse, col.data[0], x, y)
				}
			} else if y+cbuf.yoffset == cbuf.ysel && xoff+cbuf.xoffset == cbuf.xsel {
				for i := 0; i < col.maxWidth; i++ {
					termutil.PrintRune(x+i, y, ' ', termbox.AttrReverse)
				}
				termutil.PrintstringColored(termbox.AttrReverse, col.data[y+cbuf.yoffset], x, y)
			} else {
				termutil.Printstring(col.data[y+cbuf.yoffset], x, y)
			}
			x += col.maxWidth + 1
		}
	}

	// status bar
	for i := 0; i < sx; i++ {
		termutil.PrintRune(i, sy-2, ' ', termbox.AttrReverse)
	}
	sbx := 0
	if debug {
		termutil.PrintstringColored(termbox.AttrReverse, fmt.Sprint(cbuf.xsel, cbuf.ysel, cbuf.xoffset, cbuf.yoffset, cbuf.cols[cbuf.xsel].maxWidth), sbx, sy-2)
		sbx += 20
	}
	termutil.PrintstringColored(termbox.AttrReverse, cbuf.filename, sbx, sy-2)
	termutil.Printstring(message, 0, sy-1)

	termbox.Flush()
}

func loadFile(filename string) error {
	if strings.HasSuffix(strings.ToLower(filename), ".tsv") {
		var err error
		cbuf, err = createBufferFromFile(filename, '\t')
		return err
	} else if strings.HasSuffix(strings.ToLower(filename), ".csv") {
		var err error
		cbuf, err = createBufferFromFile(filename, ',')
		return err
	}
	choice := termutil.ChoiceIndex("What delimiter does the file use?", []string{"Commas (',')", "Tabs ('        ')", "Spaces (' ')", "Other"}, 0)
	switch choice {
	case 0:
		var err error
		cbuf, err = createBufferFromFile(filename, ',')
		return err
	case 1:
		var err error
		cbuf, err = createBufferFromFile(filename, '\t')
		return err
	case 2:
		var err error
		cbuf, err = createBufferFromFile(filename, ' ')
		return err
	default:
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
		var delim string
		for delim == "" {
			delim = termutil.Prompt("What delimiter does the file use?", nil)
		}
		ru, _ := utf8.DecodeRuneInString(delim)
		var err error
		cbuf, err = createBufferFromFile(filename, ru)
		return err
	}
}

func main() {
	// argparse
	flag.BoolVar(&debug, "d", false, "print debug information")
	flag.Parse()

	if flag.NArg() > 0 {
		err := loadFile(flag.Args()[0])
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	// start termbox
	termbox.Init()
	defer termbox.Close()
	sx, sy := termbox.Size()

	// main loop
	for {
		refresh(sx, sy)
		message = ""
		ev := termbox.PollEvent()
		switch ev.Type {
		case termbox.EventResize:
			termbox.Sync()
			sx, sy = termbox.Size()
			refresh(sx, sy)
		default:
			pev := termutil.ParseTermboxEvent(ev)
			switch pev {
			case "LEFT", "C-b", "h":
				if cbuf.xsel > 0 {
					cbuf.xsel--
				}
			case "RIGHT", "C-f", "l":
				if cbuf.xsel < cbuf.ncols-1 {
					cbuf.xsel++
				}
			case "UP", "C-p", "k":
				if cbuf.ysel > 0 {
					cbuf.ysel--
				}
			case "DOWN", "C-n", "j":
				if cbuf.ysel < cbuf.nrows-1 {
					cbuf.ysel++
				}
			case "RET":
				val := termutil.Prompt("Value for this cell?", refresh)
				cbuf.setSel(val)
			case "C-t":
				cbuf.titles = !cbuf.titles
			case "C-s":
				err := cbuf.save()
				if err == nil {
					message = fmt.Sprintf("Saved %s", cbuf.filename)
				} else {
					message = err.Error()
				}
			case "C-c":
				return
			}
		}
	}
}
