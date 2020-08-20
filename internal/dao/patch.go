package dao

import (
	"encoding/json"
	v1 "k8s.io/api/core/v1"
)

type JsonPatch struct {
	Spec Spec `json:"spec"`
}

type Spec struct {
	Template PodSpec `json:"template"`
}

type PodSpec struct {
	Spec ImagesSpec `json:"spec"`
}

type ImagesSpec struct {
	SetElementOrderContainers     []Element `json:"$setElementOrder/containers,omitempty"`
	SetElementOrderInitContainers []Element `json:"$setElementOrder/initContainers,omitempty"`
	Containers                    []Element `json:"containers,omitempty"`
	InitContainers                []Element `json:"initContainers,omitempty"`
}

type Element struct {
	Image string `json:"image,omitempty"`
	Name  string `json:"name"`
}

// Build a json patch string to update PodSpec images
func GetTemplateJsonPatch(spec v1.PodSpec) (string, error) {
	jsonPatch := JsonPatch{
		Spec: Spec{
			Template: getPatchPodSpec(spec),
		},
	}
	bytes, err := json.Marshal(jsonPatch)
	return string(bytes), err
}

func GetJsonPatch(spec v1.PodSpec) (string, error) {
	podSpec := getPatchPodSpec(spec)
	bytes, err := json.Marshal(podSpec)
	return string(bytes), err
}

func getPatchPodSpec(spec v1.PodSpec) PodSpec {
	podSpec := PodSpec{
		Spec: ImagesSpec{
			SetElementOrderContainers:     extractElements(spec.Containers, false),
			Containers:                    extractElements(spec.Containers, true),
			SetElementOrderInitContainers: extractElements(spec.InitContainers, false),
			InitContainers:                extractElements(spec.InitContainers, true),
		},
	}

	return podSpec
}

func extractElements(containers []v1.Container, withImage bool) []Element {
	elements := make([]Element, 0)
	for _, c := range containers {
		if withImage {
			elements = append(elements, Element{Name: c.Name, Image: c.Image})
		} else {
			elements = append(elements, Element{Name: c.Name})
		}
	}
	return elements
}
