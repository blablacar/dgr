package attributes

import (
	"io/ioutil"
)

type Inputs struct {
	Directory string
	Files     []string
}

// constructor
func NewInputs(d string) *Inputs {
	in := new(Inputs)
	in.Directory = d + "/"
	return in
}

// list input files
func (in *Inputs) ListFiles() error {
	list_l1, err := ioutil.ReadDir(in.Directory)
	if err != nil {
		return err
	}
	for _, f_l1 := range list_l1 {
		if f_l1.IsDir() {
			list_l2, err := ioutil.ReadDir(in.Directory + "/" + f_l1.Name())
			if err != nil {
				return err
			}
			for _, f_l2 := range list_l2 {
				in.Files = append(in.Files, f_l1.Name()+"/"+f_l2.Name())
			}
		} else {
			in.Files = append(in.Files, f_l1.Name())
		}
	}
	return nil
}
