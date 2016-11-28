package common

func Tar(destination string, source ...string) error { // remove zip
	params := []string{"--sort=name", "--numeric-owner", "-cpf", destination}
	params = append(params, source...)
	return ExecCmd("tar", params...)
}
