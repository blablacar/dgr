package spec

import (
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/appc/spec/discovery"
	"github.com/juju/errors"
	"net/http"
	"regexp"
	"strings"
)

type ACFullname string

func (n *ACFullname) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	nn := NewACFullName(s)
	*n = *nn
	return nil
}

func (n ACFullname) LatestVersion() (string, error) {
	app, err := discovery.NewAppFromString(n.Name() + ":latest")
	if app.Labels["os"] == "" {
		app.Labels["os"] = "linux"
	}
	if app.Labels["arch"] == "" {
		app.Labels["arch"] = "amd64"
	}

	endpoint, _, err := discovery.DiscoverEndpoints(*app, false)
	if err != nil {
		return "", errors.Annotate(err, "Latest discovery fail")
	}

	r, _ := regexp.Compile(`^(\d+\.)?(\d+\.)?(\*|\d+)(\-[\dA-Za-z]+){0,1}$`)

	url := getRedirectForLatest(endpoint.ACIEndpoints[0].ACI)
	log.Debug("latest version url is ", url)

	for _, part := range strings.Split(url, "/") {
		if r.Match([]byte(part)) {
			return part, nil
		}
	}
	return "", errors.New("No latest version found")
}

func (n ACFullname) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.String())
}

func (n ACFullname) String() string {
	return string(n)
}

/* example.com/yopla:1 */
func NewACFullName(s string) *ACFullname {
	n := ACFullname(s)
	return &n
}

func (n ACFullname) FullyResolved() (*ACFullname, error) {
	version := n.Version()
	if version != "" {
		return &n, nil
	}
	version, err := n.LatestVersion()
	if err != nil {
		return nil, errors.Annotate(err, "Cannot fully resolve AcFullname")
	}
	return NewACFullName(n.Name() + ":" + version), nil
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

/* example.com */
func (n ACFullname) DomainName() string {
	return strings.Split(n.Name(), "/")[0]
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
