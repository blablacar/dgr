package main
import (
	"fmt"
	"time"
    "gopkg.in/libgit2/git2go.v22"
	"log"
)

func GenerateVersion() string {
	return generateDate() + "-" + gitHash();
}

func generateDate() string {
	return fmt.Sprintf("%s", time.Now().Format("20060102.150405"))
}


func gitHash() string {
	repo,err := git.OpenRepository(".");
	if (err != nil) {
		log.Fatal(err)
	}

	revSpec, err := repo.Revparse("HEAD");
	if (err != nil) {
		log.Fatal(err)
	}

	return revSpec.From().Id().String()[0:6];
}
