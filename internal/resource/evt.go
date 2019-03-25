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
func NewEventList(c Connection, ns string) List {
	return NewList(
		ns,
		"ev",
		NewEvent(c),
		ListAccess+NamespaceAccess,
	)
}

// NewEvent instantiates a new Event.
func NewEvent(c Connection) *Event {
	ev := &Event{&Base{Connection: c, Resource: k8s.NewEvent(c)}, nil}
	ev.Factory = ev

	return ev
}

// New builds a new Event instance from a k8s resource.
func (r *Event) New(i interface{}) Columnar {
	c := NewEvent(r.Connection)
	switch instance := i.(type) {
	case *v1.Event:
		c.instance = instance
	case v1.Event:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown Event type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *Event) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	ev := i.(*v1.Event)
	ev.TypeMeta.APIVersion = "v1"
	ev.TypeMeta.Kind = "Event"

	return r.marshalObject(ev)
}

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

// ----------------------------------------------------------------------------
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
