package dialog

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const createKey = "create"

type (
	createFunc func(obj runtime.Object) error
)

// ShowCreate show create form
func ShowCreate(styles config.Dialog, pages *ui.Pages, log *model.Flash, gvr client.GVR, ns string, ok createFunc, cancel cancelFunc) error {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color()).
		SetFieldBackgroundColor(styles.BgColor.Color())
	menus, err := createMenu(gvr)
	if err != nil {
		return err
	}
	for _, menu := range menus {
		f.AddInputField(buildMenuKey(menu), "", 30, nil, nil)
	}
	f.AddButton("OK", func() {
		dismissCreate(pages)
		obj, err := buildResource(gvr, ns, getFormData(f, menus))
		if err != nil {
			log.Err(err)
		}
		err = ok(obj)
		if err != nil {
			log.Err(err)
		}
	})
	f.AddButton("Cancel", func() {
		dismissCreate(pages)
		cancel()
	})
	confirm := tview.NewModalForm(fmt.Sprintf("<Create %s>", gvr.R()), f)
	confirm.SetDoneFunc(func(int, string) {
		dismissCreate(pages)
		cancel()
	})
	pages.AddPage(createKey, confirm, false, false)
	pages.ShowPage(createKey)
	return nil
}

func createMenu(gvr client.GVR) ([]string, error) {
	switch gvr.R() {
	case "pods":
		return []string{"Pod name", "Image", "Command"}, nil
	case "deployment":
		return []string{"deployment name", "Image", "Command"}, nil
	default:
		return nil, fmt.Errorf("Not support this resource: %s", gvr.String())
	}
}

func getFormData(f *tview.Form, menus []string) []string {
	var data []string
	for _, menu := range menus {
		data = append(data, f.GetFormItemByLabel(buildMenuKey(menu)).(*tview.InputField).GetText())
	}
	return data
}

func buildResource(gvr client.GVR, ns string, data []string) (runtime.Object, error) {
	switch gvr.R() {
	case "pods":
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      data[0],
				Namespace: ns,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  data[0],
						Image: data[1],
					},
				},
			},
		}
		if len(data[2]) != 0 {
			pod.Spec.Containers[0].Command = strings.Split(data[2], " ")
		}
		return pod, nil
	default:
		return nil, fmt.Errorf("Not support: %s", gvr.String())
	}
}

func buildMenuKey(key string) string {
	return fmt.Sprintf("%s:", key)
}

func dismissCreate(pages *ui.Pages) {
	pages.RemovePage(createKey)
}
