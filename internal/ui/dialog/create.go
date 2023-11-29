package dialog

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	appsv1 "k8s.io/api/apps/v1"
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
	confirm := tview.NewModalForm(fmt.Sprintf("<Create %s>", strings.Title(gvr.R())), f)
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
	case "namespaces":
		return []string{"Name"}, nil
	case "pods":
		return []string{"Name", "Image", "Command"}, nil
	case "configmaps", "secrets":
		return []string{"Name", "Key", "Value"}, nil
	case "deployments":
		return []string{"Name", "Image", "Command"}, nil
	case "statefulsets":
		return []string{"Name", "Image", "Command"}, nil
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
	case "namespaces":
		return &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: data[0],
			},
			Spec: corev1.NamespaceSpec{
				Finalizers: []corev1.FinalizerName{
					corev1.FinalizerKubernetes,
				},
			},
		}, nil
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
	case "configmaps":
		configmap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      data[0],
				Namespace: ns,
			},
			Data: map[string]string{data[1]: data[2]},
		}
		return configmap, nil
	case "secrets":
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      data[0],
				Namespace: ns,
			},
			Data: map[string][]byte{data[1]: []byte(data[2])},
		}
		return secret, nil
	case "deployments":
		deployment := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      data[0],
				Namespace: ns,
			},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": data[0],
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": data[0],
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  data[0],
								Image: data[1],
							},
						},
					},
				},
			},
		}
		if len(data[2]) != 0 {
			deployment.Spec.Template.Spec.Containers[0].Command = strings.Split(data[2], " ")
		}
		return deployment, nil
	case "statefulsets":
		statefulset := &appsv1.StatefulSet{
			ObjectMeta: metav1.ObjectMeta{
				Name:      data[0],
				Namespace: ns,
			},
			Spec: appsv1.StatefulSetSpec{
				ServiceName: data[0],
				Selector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"app": data[0],
					},
				},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							"app": data[0],
						},
					},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  data[0],
								Image: data[1],
							},
						},
					},
				},
			},
		}
		if len(data[2]) != 0 {
			statefulset.Spec.Template.Spec.Containers[0].Command = strings.Split(data[2], " ")
		}
		return statefulset, nil
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
