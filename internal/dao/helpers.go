package dao

import (
	"math"

	"github.com/derailed/tview"
	runewidth "github.com/mattn/go-runewidth"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func toPerc(v1, v2 float64) float64 {
	if v2 == 0 {
		return 0
	}
	return math.Round((v1 / v2) * 100)
}

// Truncate a string to the given l and suffix ellipsis if needed.
func Truncate(str string, width int) string {
	return runewidth.Truncate(str, width, string(tview.SemigraphicsHorizontalEllipsis))
}

// FetchNodes returns a collection of nodes.
func FetchNodes(f Factory) (*v1.NodeList, error) {
	auth, err := f.Client().CanI("", "v1/nodes", []string{"list"})
	if !auth || err != nil {
		return nil, err
	}

	return f.Client().DialOrDie().CoreV1().Nodes().List(metav1.ListOptions{})
}
