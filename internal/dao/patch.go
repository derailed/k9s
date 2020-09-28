package dao

import (
	"encoding/json"
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
func GetTemplateJsonPatch(containersPatch map[string]string, initContainersPatch map[string]string) (string, error) {
	jsonPatch := JsonPatch{
		Spec: Spec{
			Template: getPatchPodSpec(containersPatch, initContainersPatch),
		},
	}
	bytes, err := json.Marshal(jsonPatch)
	return string(bytes), err
}

func GetJsonPatch(containersPatch map[string]string, initContainersPatch map[string]string) (string, error) {
	podSpec := getPatchPodSpec(containersPatch, initContainersPatch)
	bytes, err := json.Marshal(podSpec)
	return string(bytes), err
}

func getPatchPodSpec(containersPatch map[string]string, initContainersPatch map[string]string) PodSpec {
	elementsOrders, elements := extractElements(containersPatch)
	initElementsOrders, initElements := extractElements(initContainersPatch)
	podSpec := PodSpec{
		Spec: ImagesSpec{
			SetElementOrderContainers:     elementsOrders,
			Containers:                    elements,
			SetElementOrderInitContainers: initElementsOrders,
			InitContainers:                initElements,
		},
	}
	return podSpec
}

func extractElements(containers map[string]string) (elementsOrders []Element, elements []Element) {
	for name, image := range containers {
		elementsOrders = append(elementsOrders, Element{Name: name})
		elements = append(elements, Element{Name: name, Image: image})
	}
	return elementsOrders, elements
}
