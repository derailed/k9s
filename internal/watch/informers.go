package watch

// BOZO!!
// import (
// 	"fmt"

// 	"github.com/derailed/k9s/internal/k8s"
// 	"github.com/rs/zerolog/log"
// )

// type Informers struct {
// 	informers map[string]*Informer
// 	stopChan  chan struct{}
// 	client    k8s.Connection
// 	activeNS  string
// }

// func NewInformers(client k8s.Connection) *Informers {
// 	return &Informers{
// 		informers: make(map[string]*Informer),
// 		stopChan:  make(chan struct{}),
// 		client:    client,
// 	}
// }

// func (i *Informers) Dump() {
// 	log.Debug().Msgf("----------- INFORMERS -------------")
// 	for k, inf := range i.informers {
// 		if k == i.activeNS {
// 			log.Debug().Msgf("(*) %q", k)
// 		} else {
// 			log.Debug().Msgf("    %q", k)
// 			for n, v := range inf.informers {
// 				log.Debug().Msgf("      %s", n)
// 				for _, key := range v.GetStore().ListKeys() {
// 					log.Debug().Msgf("        Key: %q", key)
// 				}
// 			}
// 		}
// 	}
// }

// func (i *Informers) HasAllNamespace() bool {
// 	_, ok := i.informers[""]
// 	return ok
// }

// func (i *Informers) InformerFor(ns string) (*Informer, error) {
// 	inf, ok := i.informers[ns]
// 	if !ok {
// 		return nil, fmt.Errorf("No informer found for ns `%s", ns)
// 	}

// 	return inf, nil
// }

// func (i *Informers) SetActive(ns string) error {
// 	_, ok := i.informers[ns]
// 	if ok {
// 		i.activeNS = ns
// 		return nil
// 	}

// 	if err := i.add(ns); err != nil {
// 		return err
// 	}
// 	i.activeNS = ns
// 	i.Dump()

// 	return nil
// }

// func (i *Informers) ActiveInformer() *Informer {
// 	inf, ok := i.informers[i.activeNS]
// 	if !ok {
// 		log.Fatal().Msgf("No active informer found for %q", i.activeNS)
// 		return nil
// 	}

// 	return inf
// }

// func (i *Informers) add(ns string) error {
// 	if err := i.register(ns); err != nil {
// 		return err
// 	}
// 	i.informers[ns].Run(i.stopChan)
// 	i.Dump()

// 	return nil
// }

// func (i *Informers) register(ns string) error {
// 	_, ok := i.informers[ns]
// 	if ok {
// 		return nil
// 	}

// 	inf, err := NewInformer(i.client, ns)
// 	if err != nil {
// 		return err
// 	}
// 	i.informers[ns] = inf

// 	return nil
// }

// func (i *Informers) Restart(ns string) error {
// 	i.Stop()
// 	if err := i.register(ns); err != nil {
// 		return err
// 	}
// 	i.Start()

// 	return nil
// }

// func (i *Informers) Start() {
// 	i.Stop()
// 	i.stopChan = make(chan struct{})
// 	for k := range i.informers {
// 		i.informers[k].Run(i.stopChan)
// 	}
// }

// // Stop stops and delete all informers.
// func (i *Informers) Stop() {
// 	if i.stopChan != nil {
// 		close(i.stopChan)
// 		i.stopChan = nil
// 	}

// 	i.Clear()
// }

// // Clear stops and delete all informers.
// func (i *Informers) Clear() {
// 	for k := range i.informers {
// 		delete(i.informers, k)
// 	}
// }
