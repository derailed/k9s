package k8s

type base struct {
}

func (b *base) Kill(ns, n string) error {
	return nil
}
