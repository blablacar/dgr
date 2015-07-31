package cnt
import (
	"github.com/kardianos/osext"
	"path/filepath"
	"github.com/blablacar/cnt/log"
	"os"
	"time"
)

func BuildDate() (string, error) {
	fname, _ := osext.Executable()
	dir, err := filepath.Abs(filepath.Dir(fname))
	if err != nil {
		log.Get().Warn("Cannot get fullpath of current executable",err)
		return "", err
	}
	fi, err := os.Lstat(filepath.Join(dir, filepath.Base(fname)))
	if err != nil {
		log.Get().Warn("Cannot stat current executable",err)
		return "", err
	}
	t := fi.ModTime()
	return t.Format(time.RFC3339), nil
}