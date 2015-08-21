package utils
import (
	"os"
	"crypto/sha512"
	"io"
	"fmt"
)

const StdinFileName = "-"

func Sha512sum(filePath string) (res string, err error) {
	// Open file.
	var fr *os.File
	if filePath == StdinFileName {
		fr = os.Stdin
	} else {
		fr, err = os.Open(filePath)
		if err != nil {
			return "", err
		}
		defer fr.Close()
	}

	h := sha512.New()
	_, err = io.Copy(h, fr)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}