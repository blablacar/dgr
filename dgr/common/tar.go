package common

func Tar(destination string, source ...string) error { // remove zip
	source = append(source, "")
	source = append(source, "")
	copy(source[2:], source[0:])
	source[0] = "cpf"
	source[1] = destination

	return ExecCmd("tar", source...)
}
