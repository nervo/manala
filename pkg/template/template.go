package template

type Interface interface {
	GetDir() string
}

type template struct {
	dir string
}

func (tpl *template) GetDir() string {
	return tpl.dir
}
