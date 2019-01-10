package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"github.com/japanoise/termbox-util"
)

type column struct {
	maxWidth int
	data     []string
}

func (c *column) calcWidth() {
	c.maxWidth = 0
	for _, datum := range c.data {
		w := termutil.RunewidthStr(datum)
		if w > c.maxWidth {
			c.maxWidth = w
		}
	}
}

type buffer struct {
	titles   bool
	ncols    int
	nrows    int
	xoffset  int
	yoffset  int
	xsel     int
	ysel     int
	delim    rune
	filename string
	cols     []column
}

func (buf *buffer) addColumn() {
	col := column{}
	col.data = make([]string, buf.nrows)
	buf.cols = append(buf.cols, col)
	buf.ncols++
}

func (buf *buffer) addRow() {
	for i := range buf.cols {
		buf.cols[i].data = append(buf.cols[i].data, "")
	}
	buf.nrows++
}

func (buf *buffer) delCurColumn() {
	if buf.ncols <= 1 {
		return
	}
	copy(buf.cols[buf.xsel:], buf.cols[buf.xsel+1:])
	buf.cols[len(buf.cols)-1] = column{}
	buf.cols = buf.cols[:len(buf.cols)-1]
	buf.ncols--
	if buf.xsel >= buf.ncols {
		buf.xsel = buf.ncols - 1
	}
}

func (buf *buffer) delCurRow() {
	if buf.nrows <= 1 {
		return
	}
	index := buf.ysel
	for i := range buf.cols {
		copy(buf.cols[i].data[index:], buf.cols[i].data[index+1:])
		buf.cols[i].data[len(buf.cols[i].data)-1] = ""
		buf.cols[i].data = buf.cols[i].data[:len(buf.cols[i].data)-1]
	}
	buf.nrows--
	if index >= buf.nrows {
		buf.ysel = buf.nrows - 1
	}
}

func (buf *buffer) save() error {
	f, err := os.Create(buf.filename)
	if err != nil {
		return err
	}
	defer f.Close()

	for y := 0; y < buf.nrows; y++ {
		for i, col := range buf.cols {
			if i != 0 {
				fmt.Fprintf(f, "%s", string(buf.delim))
			}
			fmt.Fprintf(f, "%s", col.data[y])
		}
		fmt.Fprintln(f)
	}
	return nil
}

func (buf *buffer) setSel(val string) {
	buf.cols[buf.xsel].data[buf.ysel] = val
	w := termutil.RunewidthStr(val)
	if w > buf.cols[buf.xsel].maxWidth {
		buf.cols[buf.xsel].maxWidth = w
	}
}

func (buf *buffer) getSel() string {
	return buf.cols[buf.xsel].data[buf.ysel]
}

func (buf *buffer) nextCell() {
	buf.xsel++
	if buf.xsel >= buf.ncols {
		buf.xsel = 0
		buf.ysel++
		if buf.ysel >= buf.nrows {
			buf.addRow()
		}
	}
}

func (buf *buffer) BOL() {
	buf.xsel = 0
}

func (buf *buffer) EOL() {
	buf.xsel = buf.ncols - 1
}

func createBlankBuffer() *buffer {
	buf := buffer{}
	buf.addColumn()
	buf.addRow()
	return &buf
}

func createBufferFromFile(filename string, delim rune) (*buffer, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	nbuf := buffer{}
	nbuf.addColumn()
	nbuf.delim = delim
	nbuf.filename = filename
	row := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		nbuf.addRow()
		col := 0
		buf := bytes.Buffer{}
		instring := false
		line := scanner.Text()
		if debug {
			fmt.Println(line)
		}
		for _, ru := range line {
			if ru == delim && !instring {
				nbuf.cols[col].data[row] = buf.String()
				w := termutil.RunewidthStr(nbuf.cols[col].data[row])
				if w > nbuf.cols[col].maxWidth {
					nbuf.cols[col].maxWidth = w
				}
				buf = bytes.Buffer{}
				col++
				if col >= nbuf.ncols {
					nbuf.addColumn()
				}
			} else {
				if ru == '"' {
					instring = !instring
				}
				buf.WriteRune(ru)
			}
		}
		str := buf.String()
		if str != "" {
			nbuf.cols[col].data[row] = str
			w := termutil.RunewidthStr(str)
			if w > nbuf.cols[col].maxWidth {
				nbuf.cols[col].maxWidth = w
			}
		}
		row++
	}

	return &nbuf, nil
}
