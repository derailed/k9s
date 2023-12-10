package config

type Labels map[string][]string

func (l Labels) exclude(k, val string) bool {
	vv, ok := l[k]
	if !ok {
		return false
	}

	for _, v := range vv {
		if v == val {
			return true
		}
	}

	return false
}

type BlackList struct {
	Labels Labels `yaml:"labels"`
}

func newBlackList() BlackList {
	return BlackList{
		Labels: make(Labels),
	}
}

func (b BlackList) exclude(ll map[string]string) bool {
	for k, v := range ll {
		if b.Labels.exclude(k, v) {
			return true
		}
	}

	return false
}

type ImageScans struct {
	Enable    bool      `yaml:"enable"`
	BlackList BlackList `yaml:"blackList"`
}

func NewImageScans() *ImageScans {
	return &ImageScans{
		BlackList: newBlackList(),
	}
}

func (i *ImageScans) ShouldExclude(ll map[string]string) bool {
	if !i.Enable {
		return false
	}

	return i.BlackList.exclude(ll)
}
