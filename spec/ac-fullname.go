package spec

import (
	"encoding/json"
	"errors"
	"github.com/appc/spec/discovery"
	"github.com/blablacar/cnt/log"
	"net/http"
	"regexp"
	"strings"
)

type ACFullname string

func (n *ACFullname) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		log.Get().Panic(err)
		return err
	}
	nn, err := NewACFullName(s)
	if err != nil {
		log.Get().Panic(err)
		return err
	}
	*n = *nn
	return nil
}

func (n ACFullname) LatestVersion() string {
	app, err := discovery.NewAppFromString(n.Name() + ":latest")
	if app.Labels["os"] == "" {
		app.Labels["os"] = "linux"
	}
	if app.Labels["arch"] == "" {
		app.Labels["arch"] = "amd64"
	}

	endpoint, _, err := discovery.DiscoverEndpoints(*app, false)
	if err != nil {
		return ""
	}

	r, _ := regexp.Compile(`^(\d+\.)?(\d+\.)?(\*|\d+)$`)

	url := getRedirectForLatest(endpoint.ACIEndpoints[0].ACI)
	log.Get().Debug("latest version url is ", url)

	for _, part := range strings.Split(url, "/") {
		if r.Match([]byte(part)) {
			return part
		}
	}
	return ""
}

func (n ACFullname) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

func (n ACFullname) String() string {
	return string(n)
}

/* example.com/yopla:1 */
func NewACFullName(s string) (*ACFullname, error) {
	n := ACFullname(s)
	return &n, nil
}

func (n ACFullname) FullyResolved() (*ACFullname, error) {
	version := n.Version()
	log.Get().Error("Version:" + version)
	if version != "" {
		return &n, nil
	}
	return NewACFullName(n.Name() + ":" +  n.LatestVersion())
}

/* 1 */
func (n ACFullname) Version() string {
	split := strings.Split(string(n), ":")
	if len(split) == 1 {
		return ""
	}
	return split[1]
}

/* yopla:1 */
func (n ACFullname) ShortNameId() string {
	return strings.Split(string(n), "/")[1]
}

/* yopla */
func (n ACFullname) ShortName() string {
	return strings.Split(n.Name(), "/")[1]
}

/* example.com/yopla */
func (n ACFullname) Name() string {
	return strings.Split(string(n), ":")[0]
}

////////////////////////////////

func getRedirectForLatest(url string) string {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}
	transport := http.DefaultTransport
	//	if insecureSkipVerify {
	//		transport = &http.Transport{
	//			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	//		}
	//	}
	client := &http.Client{Transport: transport}
	myurl := ""
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		myurl = req.URL.Path
		return errors.New("do not want to get the file")
	}
	_, err2 := client.Do(req)
	if err2 != nil {
		if myurl != "" {
			return myurl
		}
		return ""
	}
	return myurl
}
