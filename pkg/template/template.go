package template

type Interface interface {
	GetDir() string
}

type template struct {
	dir string
}

func (t *template) GetDir() string {
	return t.dir
}
