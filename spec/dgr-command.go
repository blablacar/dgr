package spec

type DgrCommand interface {
	Build() error
	Clean()
	Push()
	Install()
	Test()
	Graph()
	Update() error
}
