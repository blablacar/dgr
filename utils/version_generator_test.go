package utils

import (
	"testing"
)

func TestVersionGenerator(t *testing.T) {
	v := GenerateVersion()
	println(v)
}
