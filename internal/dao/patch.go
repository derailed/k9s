package dao

import (
	"encoding/json"
)

type ImageSpec struct {
	Index             int
	Name, DockerImage string
	Init              bool
}

type ImageSpecs []ImageSpec

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
func GetTemplateJsonPatch(imageSpecs ImageSpecs) ([]byte, error) {
	jsonPatch := JsonPatch{
		Spec: Spec{
			Template: getPatchPodSpec(imageSpecs),
		},
	}
	return json.Marshal(jsonPatch)
}

func GetJsonPatch(imageSpecs ImageSpecs) ([]byte, error) {
	podSpec := getPatchPodSpec(imageSpecs)
	return json.Marshal(podSpec)
}

func getPatchPodSpec(imageSpecs ImageSpecs) PodSpec {
	initElementsOrders, initElements, elementsOrders, elements := extractElements(imageSpecs)
	podSpec := PodSpec{
		Spec: ImagesSpec{
			SetElementOrderInitContainers: initElementsOrders,
			InitContainers:                initElements,
			SetElementOrderContainers:     elementsOrders,
			Containers:                    elements,
		},
	}
	return podSpec
}

func extractElements(imageSpecs ImageSpecs) (initElementsOrders []Element, initElements []Element, elementsOrders []Element, elements []Element) {
	for _, spec := range imageSpecs {
		if spec.Init {
			initElementsOrders = append(initElementsOrders, Element{Name: spec.Name})
			initElements = append(initElements, Element{Name: spec.Name, Image: spec.DockerImage})
		} else {
			elementsOrders = append(elementsOrders, Element{Name: spec.Name})
			elements = append(elements, Element{Name: spec.Name, Image: spec.DockerImage})
		}
	}
	return initElementsOrders, initElements, elementsOrders, elements
}
