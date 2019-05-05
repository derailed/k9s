package watch

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	wv1 "k8s.io/client-go/informers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubernetes/pkg/util/node"
)

const (
	// PodIndex marker for stored pods.
	PodIndex string = "po"
	podCols         = 11
)

// Pod tracks pod activities.
type Pod struct {
	cache.SharedIndexInformer

	client   k8s.Connection
	data     RowEvents
	mxData   k8s.PodsMetrics
	ns       string
	listener TableListenerFn
	activeNS *string
}

// NewPod returns a new pod.
func NewPod(client k8s.Connection, ns string) *Pod {
	po := Pod{
		ns:     ns,
		client: client,
		data:   RowEvents{},
		mxData: k8s.PodsMetrics{},
	}

	if client == nil {
		return &po
	}

	po.SharedIndexInformer = wv1.NewPodInformer(
		client.DialOrDie(),
		ns,
		0,
		cache.Indexers{},
	)
	po.AddEventHandler(&po)

	return &po
}

// SetListener registers event recipient.
func (p *Pod) SetListener(ns string, cb TableListenerFn) {
	p.listener, p.activeNS = cb, &ns
	p.fireChanged(*p.activeNS)
}

// UnsetListener unregister event recipient.
func (p *Pod) UnsetListener(_ string) {
	p.listener, p.activeNS = nil, nil
}

// Data return current data.
func (p *Pod) tableData(ns string) TableData {
	// Filter list based on active namespace.
	data := RowEvents{}
	for k := range p.data {
		pns, _ := namespaced(k)
		if ns == AllNamespaces || pns == ns {
			data[k] = p.data[k]
		}
	}

	return TableData{
		Header:    p.header(),
		Rows:      data,
		Namespace: ns,
	}
}

// List all pods from store in the given namespace.
func (p *Pod) List(ns string) (k8s.Collection, error) {
	var res k8s.Collection
	for _, o := range p.GetStore().List() {
		pod := o.(*v1.Pod)
		if ns == "" || pod.Namespace == ns {
			res = append(res, pod)
		}
	}
	return res, nil
}

// Get retrieves a given pod from store.
func (p *Pod) Get(fqn string) (interface{}, error) {
	o, ok, err := p.GetStore().GetByKey(fqn)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("Pod %s not found", fqn)
	}

	return o, nil
}

func (p *Pod) fireChanged(ns string) {
	if cb := p.listener; cb != nil {
		cb(p.tableData(*p.activeNS))
	}
}

// OnAdd notify pod added.
func (p *Pod) OnAdd(obj interface{}) {
	po := obj.(*v1.Pod)
	ff := make(Row, podCols)
	p.fields(p.ns, po, ff)
	fqn := MetaFQN(po.ObjectMeta)
	// log.Debug().Msgf("Pod Added %s", fqn)

	p.data[fqn] = &RowEvent{
		Action: watch.Added,
		Fields: ff,
		Deltas: make(Row, len(ff)),
	}

	if p.HasSynced() {
		p.fireChanged(po.Namespace)
	}
}

// OnUpdate notify pod updated.
func (p *Pod) OnUpdate(oldObj, newObj interface{}) {
	opo, npo := oldObj.(*v1.Pod), newObj.(*v1.Pod)

	k1 := MetaFQN(opo.ObjectMeta)
	k2 := MetaFQN(npo.ObjectMeta)
	// log.Debug().Msgf("Pod Updated %#v -- %#v", opo.Name, npo.Name)

	p.deltas(opo, npo)
	ff := make(Row, podCols)
	p.fields(p.ns, npo, ff)
	if re, ok := p.data[k1]; ok {
		re.Action = watch.Modified
		re.Deltas = re.Fields
		re.Fields = ff
	}
	p.data[k2] = &RowEvent{
		Action: watch.Added,
		Fields: ff,
		Deltas: make(Row, len(ff)),
	}
	p.fireChanged(npo.Namespace)
}

func (p *Pod) deltas(p1, p2 *v1.Pod) {
	f1 := make(Row, podCols)
	p.fields(p.ns, p1, f1)
	f2 := make(Row, podCols)
	p.fields(p.ns, p2, f2)
	for i := 0; i < len(f1); i++ {
		if f1[i] != f2[i] {
			log.Debug().Msgf("Pod changed %s - %s", f1[i], f2[i])
		}
	}
}

// OnDelete notify pod was deleted.
func (p *Pod) OnDelete(obj interface{}) {
	po := obj.(*v1.Pod)
	key := MetaFQN(po.ObjectMeta)
	// log.Debug().Msgf("Pod Delete %s", key)
	delete(p.data, key)
	p.fireChanged(po.Namespace)
}

// header return resource header.
func (*Pod) header() Row {
	var hh Row
	return append(hh,
		"NAMESPACE",
		"NAME",
		"READY",
		"STATUS",
		"RS",
		"CPU",
		"MEM",
		"IP",
		"NODE",
		"QOS",
		"AGE",
	)
}

// fields retrieves displayable fields.
func (p *Pod) fields(ns string, pod *v1.Pod, ff Row) {
	var col int
	ff[col] = pod.ObjectMeta.Namespace
	col++
	ff[col] = pod.ObjectMeta.Name
	col++
	ss := pod.Status.ContainerStatuses
	cr, _, rc := p.statuses(ss)
	ff[col] = strconv.Itoa(cr) + "/" + strconv.Itoa(len(ss))
	col++
	ff[col] = p.phase(pod)
	col++
	ff[col] = strconv.Itoa(rc)
	col++
	fqn := MetaFQN(pod.ObjectMeta)
	p.fetchMetrics()
	mx := p.mxData[fqn]
	ff[col] = ToMillicore(mx.CurrentCPU)
	col++
	ff[col] = ToMi(mx.CurrentMEM)
	col++
	ff[col] = pod.Status.PodIP
	col++
	ff[col] = pod.Spec.NodeName
	col++
	ff[col] = p.mapQOS(pod.Status.QOSClass)
	col++
	ff[col] = toAge(pod.ObjectMeta.CreationTimestamp)
	col++
}

func (p *Pod) fetchMetrics() {
	if p.client == nil {
		return
	}

	client := k8s.NewMetricsServer(p.client)
	mx, err := client.FetchPodsMetrics(p.ns)
	if err != nil {
		log.Error().Err(err).Msg("Pod metrics failed")
		return
	}
	client.PodsMetrics(mx, p.mxData)
}

// ----------------------------------------------------------------------------
// Helpers...

func (*Pod) mapQOS(class v1.PodQOSClass) string {
	switch class {
	case v1.PodQOSGuaranteed:
		return "GA"
	case v1.PodQOSBurstable:
		return "BU"
	default:
		return "BE"
	}
}

func (*Pod) statuses(ss []v1.ContainerStatus) (cr, ct, rc int) {
	for _, c := range ss {
		if c.State.Terminated != nil {
			ct++
		}
		if c.Ready {
			cr = cr + 1
		}
		rc += int(c.RestartCount)
	}

	return
}

func isSet(s *string) bool {
	return s != nil && *s != ""
}

func (p *Pod) phase(po *v1.Pod) string {
	status := string(po.Status.Phase)
	if po.Status.Reason != "" {
		if po.DeletionTimestamp != nil && po.Status.Reason == node.NodeUnreachablePodReason {
			return "Unknown"
		}
		status = po.Status.Reason
	}

	var init bool
	init, status = p.initContainerPhase(po.Status, len(po.Spec.InitContainers), status)
	if init {
		return status
	}

	var running bool
	running, status = p.containerPhase(po.Status, status)
	if status == "Completed" && running {
		status = "Running"
	}

	if po.DeletionTimestamp == nil {
		return status
	}

	return "Terminated"
}

func (*Pod) containerPhase(st v1.PodStatus, status string) (bool, string) {
	var running bool
	for i := len(st.ContainerStatuses) - 1; i >= 0; i-- {
		cs := st.ContainerStatuses[i]
		switch {
		case cs.State.Waiting != nil && cs.State.Waiting.Reason != "":
			status = cs.State.Waiting.Reason
		case cs.State.Terminated != nil && cs.State.Terminated.Reason != "":
			status = cs.State.Terminated.Reason
		case cs.State.Terminated != nil:
			if cs.State.Terminated.Signal != 0 {
				status = "Signal:" + strconv.Itoa(int(cs.State.Terminated.Signal))
			} else {
				status = "ExitCode:" + strconv.Itoa(int(cs.State.Terminated.ExitCode))
			}
		case cs.Ready && cs.State.Running != nil:
			running = true
		}
	}

	return running, status
}

func (*Pod) initContainerPhase(st v1.PodStatus, initCount int, status string) (bool, string) {
	var init bool
	for i, cs := range st.InitContainerStatuses {
		switch {
		case cs.State.Terminated != nil:
			if cs.State.Terminated.ExitCode == 0 {
				continue
			}
			if cs.State.Terminated.Reason != "" {
				status = "Init:" + cs.State.Terminated.Reason
				break
			}
			if cs.State.Terminated.Signal != 0 {
				status = "Init:Signal:" + strconv.Itoa(int(cs.State.Terminated.Signal))
			} else {
				status = "Init:ExitCode:" + strconv.Itoa(int(cs.State.Terminated.ExitCode))
			}
		case cs.State.Waiting != nil && cs.State.Waiting.Reason != "" && cs.State.Waiting.Reason != "PodInitializing":
			status = "Init:" + cs.State.Waiting.Reason
		default:
			status = "Init:" + strconv.Itoa(i) + "/" + strconv.Itoa(initCount)
		}
		init = true
		break
	}

	return init, status
}
