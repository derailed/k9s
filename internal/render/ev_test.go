// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

// BOZO!!
// func TestEventRender(t *testing.T) {
// 	c := render.Event{}
// 	r := model1.NewRow(7)
// 	c.Render(load(t, "ev"), "", &r)

// 	assert.Equal(t, "default/hello-1567197780-mn4mv.15bfce150bd764dd", r.ID)
// 	assert.Equal(t, model1.Fields{"default", "pod:hello-1567197780-mn4mv", "Normal", "Pulled", "kubelet", "1", `Successfully pulled image "blang/busybox-bash"`}, r.Fields[:7])
// }

// func BenchmarkEventRender(b *testing.B) {
// 	ev := load(b, "ev")
// 	var re render.Event
// 	r := model1.NewRow(7)

// 	b.ResetTimer()
// 	b.ReportAllocs()
// 	for i := 0; i < b.N; i++ {
// 		_ = re.Render(&ev, "", &r)
// 	}
// }
