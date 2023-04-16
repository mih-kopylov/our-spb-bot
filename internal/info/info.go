package info

func NewInfo(version string, commit string) *Info {
	return &Info{
		Version: version,
		Commit:  commit,
	}
}

type Info struct {
	Version string
	Commit  string
}
