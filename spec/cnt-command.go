package spec

type CntCommand interface {
	Build() error
	Clean()
	Push()
	Install()
	Test()
	Graph()
	Update() error
}
