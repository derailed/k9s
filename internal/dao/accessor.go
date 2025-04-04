package dao

import (
	"log/slog"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/slogs"
)

var accessors = Accessors{
	*client.WkGVR:  new(Workload),
	*client.CtGVR:  new(Context),
	*client.CoGVR:  new(Container),
	*client.ScnGVR: new(ImageScan),
	*client.SdGVR:  new(ScreenDump),
	*client.BeGVR:  new(Benchmark),
	*client.PfGVR:  new(PortForward),
	*client.DirGVR: new(Dir),

	*client.SvcGVR:  new(Service),
	*client.PodGVR:  new(Pod),
	*client.NodeGVR: new(Node),
	*client.NsGVR:   new(Namespace),
	*client.CmGVR:   new(ConfigMap),
	*client.SecGVR:  new(Secret),

	*client.DpGVR:  new(Deployment),
	*client.DsGVR:  new(DaemonSet),
	*client.StsGVR: new(StatefulSet),
	*client.RsGVR:  new(ReplicaSet),

	*client.CjGVR:  new(CronJob),
	*client.JobGVR: new(Job),

	*client.HmGVR:  new(HelmChart),
	*client.HmhGVR: new(HelmHistory),

	*client.CrdGVR: new(CustomResourceDefinition),
}

// Accessors represents a collection of dao accessors.
type Accessors map[client.GVR]Accessor

// AccessorFor returns a client accessor for a resource if registered.
// Otherwise it returns a generic accessor.
// Customize here for non resource types or types with metrics or logs.
func AccessorFor(f Factory, gvr *client.GVR) (Accessor, error) {
	r, ok := accessors[*gvr]
	if !ok {
		r = new(Scaler)
		slog.Debug("No DAO registry entry. Using generics!", slogs.GVR, gvr)
	}
	r.Init(f, gvr)

	return r, nil
}
