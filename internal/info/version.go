package info

import (
	"github.com/goioc/di"
	"github.com/samber/lo"
)

const (
	BeanId = "Info"
)

func RegisterBean(version string, commit string) {
	_ = lo.Must(di.RegisterBeanInstance(BeanId, &Info{
		Version: version,
		Commit:  commit,
	}))
}

type Info struct {
	Version string
	Commit  string
}
