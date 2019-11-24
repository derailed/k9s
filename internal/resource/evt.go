package resource

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
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
func (r *Event) New(i interface{}) (Columnar, error) {
	c := NewEvent(r.Connection)
	switch instance := i.(type) {
	case *v1.Event:
		c.instance = instance
	case v1.Event:
		c.instance = &instance
	default:
		return nil, fmt.Errorf("Expecting Event but got %T", instance)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c, nil
}

// Marshal resource to yaml.
func (r *Event) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	ev, ok := i.(*v1.Event)
	if !ok {
		return "", errors.New("expecting evt resource")
	}
	ev.TypeMeta.APIVersion = "v1"
	ev.TypeMeta.Kind = "Event"

	return r.marshalObject(ev)
}

// Delete a resource by name.
func (r *Event) Delete(path string, cascade, force bool) error {
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
		Truncate(i.Message, 80),
		toAge(i.LastTimestamp),
	)
}
