package tchart

import (
	"fmt"
	"log/slog"
	"sort"
	"time"
)

type MetricSeries map[time.Time]float64

type Times []time.Time

func (tt Times) Len() int {
	return len(tt)
}

func (tt Times) Swap(i, j int) {
	tt[i], tt[j] = tt[j], tt[i]
}

func (tt Times) Less(i, j int) bool {
	return tt[i].Sub(tt[j]) <= 0
}

func (tt Times) Includes(ti time.Time) bool {
	for _, t := range tt {
		if t.Equal(ti) {
			return true
		}
	}
	return false
}

func (mm MetricSeries) Empty() bool {
	return len(mm) == 0
}

func (mm MetricSeries) Merge(metrics MetricSeries) {
	for k, v := range metrics {
		mm[k] = v
	}
}

func (mm MetricSeries) Dump() {
	slog.Debug("METRICS")
	for _, k := range mm.Keys() {
		slog.Debug(fmt.Sprintf("%v: %f", k, mm[k]))
	}
}

func (mm MetricSeries) Add(t time.Time, f float64) {
	if _, ok := mm[t]; !ok {
		mm[t] = f
	}
}

func (mm MetricSeries) Keys() Times {
	kk := make(Times, 0, len(mm))
	for k := range mm {
		kk = append(kk, k)
	}
	sort.Sort(kk)

	return kk
}

func (mm MetricSeries) Truncate(size int) {
	kk := mm.Keys()
	kk = kk[0 : len(kk)-size]
	for t := range mm {
		if kk.Includes(t) {
			continue
		}
		delete(mm, kk[0])
	}
}
