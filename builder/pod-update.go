package builder

import "github.com/n0rad/go-erlog/logs"

func (p *Pod) Update() error {
	logs.WithF(p.fields).Fatal("Update is not implemented on pod") // TODO update for pod
	return nil
}
