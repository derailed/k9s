package dao

import "encoding/json"

type JsonPatch struct {
	Spec Spec `json:"spec"`
}

type Spec struct {
	Template Template `json:"template"`
}

type Template struct {
	Spec ImagesSpec `json:"spec"`
}

type ImagesSpec struct {
	SetElementOrders []Element `json:"$setElementOrder/containers"`
	Containers       []Element `json:"containers"`
}

type Element struct {
	Image string `json:"image,omitempty"`
	Name  string `json:"name"`
}

// Build a json patch string to update PodSpec images
func SetImageJsonPatch(images map[string]string) (string, error) {
	elementOrders := make([]Element, 0)
	containers := make([]Element, 0)
	for key, value := range images {
		elementOrders = append(elementOrders, Element{Name: key})
		containers = append(containers, Element{Name: key, Image: value})
	}
	jsonPatch := JsonPatch{
		Spec: Spec{
			Template: Template{
				Spec: ImagesSpec{
					SetElementOrders: elementOrders,
					Containers:       containers,
				},
			},
		},
	}
	bytes, err := json.Marshal(jsonPatch)
	return string(bytes), err
}
