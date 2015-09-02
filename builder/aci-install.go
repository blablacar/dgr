package builder

import "github.com/blablacar/cnt/utils"

func (cnt *Img) Install() {
	cnt.checkBuilt()
	utils.ExecCmd("rkt", "--insecure-skip-verify=true", "fetch", cnt.target+"/image.aci")
}
