package common

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"github.com/appc/spec/discovery"
	"github.com/juju/errors"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
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

	endpoints, _, err := discovery.DiscoverACIEndpoints(*app, nil, discovery.InsecureTLS|discovery.InsecureHTTP, 0) //TODO support security
	if err != nil {
		return "", errors.Annotate(err, "Latest discovery fail")
	}

	if len(endpoints) == 0 {
		return "", errs.WithF(data.WithField("aci", string(n)), "Discovery does not give an endpoint to check latest version")
	}

	url := getRedirectForLatest(endpoints[0].ACI)
	logs.WithField("url", url).Debug("Latest version url")

	if version, err := checkForAcserver(url); err == nil {
		return version, nil
	}
	if version, err := checkForNexus(url); err == nil {
		return version, nil
	}
	return "", errors.New("No latest version found")
}

var acserverPathRegex = regexp.MustCompile(`(\d+(?:.\d+){0,2}(-[.\-\dA-Za-z]+){0,1})-linux-amd64\.aci$`)

func checkForAcserver(url string) (string, error) {
	res := acserverPathRegex.FindStringSubmatch(url)
	if len(res) > 0 {
		return res[1], nil
	}
	return "", errors.New("No latest version found")
}

var nexusPathRegex = regexp.MustCompile(`^\d+(.\d+){0,2}(-[.\-\dA-Za-z]+){0,1}$`)

func checkForNexus(url string) (string, error) {
	for _, part := range strings.Split(url, "/") {
		if nexusPathRegex.Match([]byte(part)) {
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

/* example.com/dgr/yopla:1 */
func NewACFullName(s string) *ACFullname {
	n := ACFullname(s)
	return &n
}

func NewACFullnameWithVersion(source ACFullname, version string) *ACFullname {
	return NewACFullnameFromNameAndVersion(source.Name(), version)
}

func NewACFullnameFromNameAndVersion(name string, version string) *ACFullname {
	return NewACFullName(name + ":" + version)
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
	return NewACFullnameWithVersion(n, version), nil
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
func (n ACFullname) TinyNameId() string {
	split := strings.Split(string(n), "/")
	return split[len(split)-1]
}

/* dgr/yopla */
func (n ACFullname) ShortName() string {
	name := n.Name()
	if !strings.Contains(name, "/") {
		return name
	}
	return strings.SplitN(n.Name(), "/", 2)[1]
}

/* yopla */
func (n ACFullname) TinyName() string {
	split := strings.Split(n.Name(), "/")
	return split[len(split)-1]
}

/* example.com */
func (n ACFullname) DomainName() string {
	return strings.Split(n.Name(), "/")[0]
}

/* example.com/dgr/yopla */
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
