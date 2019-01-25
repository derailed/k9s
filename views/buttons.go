package views

// type buttonView struct {
// 	*tview.Grid
// }

// func newButtonView() *buttonView {
// 	v := buttonView{Grid: tview.NewGrid()}
// 	v.SetBorder(true)
// 	v.SetTitle("Buttons")
// 	v.SetRows(1, 1, 1, 1)
// 	v.SetColumns(5, 5, 5, 5)
// 	v.SetGap(1, 1)

// 	for r := 0; r < 4; r++ {
// 		for c := 0; c < 4; c++ {
// 			b := tview.NewButton(fmt.Sprintf("%d:%d", r, c))
// 			b.SetBackgroundColor(tcell.ColorGray)
// 			v.AddItem(b, r, c, 1, 1, 1, 1, false)
// 		}
// 	}

// 	return &v
// }

// func (b *buttonView) init(context.Context) {
// }
