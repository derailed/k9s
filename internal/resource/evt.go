package resource

import (
	"regexp"
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

// Event tracks a kubernetes resource.
type Event struct {
	*Base
	instance *v1.Event
}

// NewEventList returns a new resource list.
func NewEventList(ns string) List {
	return NewEventListWithArgs(ns, NewEvent())
}

// NewEventListWithArgs returns a new resource list.
func NewEventListWithArgs(ns string, res Resource) List {
	return newList(ns, "event", res, ListAccess+NamespaceAccess)
}

// NewEvent instantiates a new Endpoint.
func NewEvent() *Event {
	return NewEventWithArgs(k8s.NewEvent())
}

// NewEventWithArgs instantiates a new Endpoint.
func NewEventWithArgs(r k8s.Res) *Event {
	ep := &Event{
		Base: &Base{
			caller: r,
		},
	}
	ep.creator = ep
	return ep
}

// NewInstance builds a new Endpoint instance from a k8s resource.
func (*Event) NewInstance(i interface{}) Columnar {
	cm := NewEvent()
	switch i.(type) {
	case *v1.Event:
		cm.instance = i.(*v1.Event)
	case v1.Event:
		ii := i.(v1.Event)
		cm.instance = &ii
	default:
		log.Fatal().Msgf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// Marshal resource to yaml.
func (r *Event) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
	if err != nil {
		return "", err
	}

	ev := i.(*v1.Event)
	ev.TypeMeta.APIVersion = "v1"
	ev.TypeMeta.Kind = "Event"
	return r.marshalObject(ev)
}

// // Get resource given a namespaced name.
// func (r *Event) Get(path string) (Columnar, error) {
// 	ns, n := namespaced(path)
// 	i, err := r.caller.Get(ns, n)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return r.NewInstance(i), nil
// }

// Delete a resource by name.
func (r *Event) Delete(path string) error {
	return nil
}

// Header return resource header.
func (*Event) Header(ns string) Row {
	var ff Row
	if ns == AllNamespaces {
		ff = append(ff, "NAMESPACE")
	}
	return append(ff, "NAME", "REASON", "SOURCE", "COUNT", "MESSAGE", "AGE")
}

var rx = regexp.MustCompile(`(.+)\.(.+)`)

// Fields returns display fields.
func (r *Event) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance

	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	return append(ff,
		i.Name,
		i.Reason,
		i.Source.Component,
		strconv.Itoa(int(i.Count)),
		Truncate(i.Message, 50),
		toAge(i.LastTimestamp),
	)
}

// ExtFields returns extended fields in relation to headers.
func (*Event) ExtFields() Properties {
	return Properties{}
}

// Helpers...

func (*Event) toEmoji(t, r string) string {
	switch t {
	case "Warning":
		switch r {
		case "Failed":
			return "😡"
		case "Killing":
			return "👿"
		default:
			return "😡"
		}
	default:
		switch r {
		case "Killing":
			return "👿"
		case "BackOff":
			return "👹"
		default:
			return "😮"
		}
	}
}
