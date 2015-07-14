package utils

func Tar(zip bool, destination string, source ...string) error {
	zipFlag := ""
	if zip {
		zipFlag = "z"
	}

	source = append(source, "")
	source = append(source, "")
	copy(source[2:], source[0:])
	source[0] = "cf" + zipFlag
	source[1] = destination

	return ExecCmd("tar", source...)
}
