package render

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var (
	totalRx = regexp.MustCompile(`Total:\s+([0-9.]+)\ssecs`)
	reqRx   = regexp.MustCompile(`Requests/sec:\s+([0-9.]+)`)
	okRx    = regexp.MustCompile(`\[2\d{2}\]\s+(\d+)\s+responses`)
	errRx   = regexp.MustCompile(`\[[4-5]\d{2}\]\s+(\d+)\s+responses`)
	toastRx = regexp.MustCompile(`Error distribution`)
)

// BenchInfo represents benchmark run info.
type BenchInfo struct {
	File os.FileInfo
	Path string
}

// Bench renders a benchmarks to screen.
type Bench struct{}

// ColorerFunc colors a resource row.
func (Bench) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		c := tcell.ColorPaleGreen
		statusCol := 2
		if strings.TrimSpace(re.Row.Fields[statusCol]) != "pass" {
			c = ErrColor
		}
		return c
	}
}

// Header returns a header row.
func (Bench) Header(ns string) HeaderRow {
	return HeaderRow{
		Header{Name: "NAMESPACE", Align: tview.AlignLeft},
		Header{Name: "NAME", Align: tview.AlignLeft},
		Header{Name: "STATUS", Align: tview.AlignLeft},
		Header{Name: "TIME", Align: tview.AlignLeft},
		Header{Name: "REQ/S", Align: tview.AlignRight},
		Header{Name: "2XX", Align: tview.AlignRight},
		Header{Name: "4XX/5XX", Align: tview.AlignRight},
		Header{Name: "REPORT", Align: tview.AlignLeft},
		Header{Name: "AGE", Align: tview.AlignLeft},
	}
}

// Render renders a K8s resource to screen.
func (b Bench) Render(o interface{}, ns string, r *Row) error {
	bench, ok := o.(BenchInfo)
	if !ok {
		return fmt.Errorf("Expected string, but got %T", o)
	}

	data, err := b.readFile(bench.Path)
	if err != nil {
		return fmt.Errorf("Unable to load bench file %s", bench.Path)
	}

	r.Fields = make(Fields, len(b.Header(ns)))
	if err := b.initRow(r.Fields, bench.File); err != nil {
		return err
	}
	b.augmentRow(r.Fields, data)
	r.ID = bench.Path

	return nil
}

// Helpers...

func (Bench) readFile(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (Bench) initRow(row Fields, f os.FileInfo) error {
	tokens := strings.Split(f.Name(), "_")
	if len(tokens) < 2 {
		return fmt.Errorf("Invalid file name %s", f.Name())
	}
	row[0] = tokens[0]
	row[1] = tokens[1]
	row[7] = f.Name()
	row[8] = time.Since(f.ModTime()).String()

	return nil
}

func (b Bench) augmentRow(fields Fields, data string) {
	if len(data) == 0 {
		return
	}

	col := 2
	fields[col] = "pass"
	mf := toastRx.FindAllStringSubmatch(data, 1)
	if len(mf) > 0 {
		fields[col] = "fail"
	}
	col++

	mt := totalRx.FindAllStringSubmatch(data, 1)
	if len(mt) > 0 {
		fields[col] = mt[0][1]
	}
	col++

	mr := reqRx.FindAllStringSubmatch(data, 1)
	if len(mr) > 0 {
		fields[col] = mr[0][1]
	}
	col++

	ms := okRx.FindAllStringSubmatch(data, -1)
	fields[col] = b.countReq(ms)
	col++

	me := errRx.FindAllStringSubmatch(data, -1)
	fields[col] = b.countReq(me)
}

func (Bench) countReq(rr [][]string) string {
	if len(rr) == 0 {
		return "0"
	}

	var sum int
	for _, m := range rr {
		if m, err := strconv.Atoi(string(m[1])); err == nil {
			sum += m
		}
	}
	return asNum(sum)
}

// AsNumb prints a number with thousand separator.
func asNum(n int) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", n)
}
