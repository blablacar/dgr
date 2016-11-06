// Copyright 2015 The appc Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package discovery

import (
	"fmt"
	"github.com/appc/spec/discovery"
	"io"
	"net/http"
	"net/url"
)

func httpsOrHTTP(name string, hostHeaders map[string]http.Header, insecure discovery.InsecureOption) (urlStr string, body io.ReadCloser, err error) {
	fetch := func(scheme string) (urlStr string, res *http.Response, err error) {
		u, err := url.Parse(scheme + "://" + name)
		if err != nil {
			return "", nil, err
		}
		u.RawQuery = "ac-discovery=1"
		urlStr = u.String()
		req, err := http.NewRequest("GET", urlStr, nil)
		if err != nil {
			return "", nil, err
		}
		if hostHeader, ok := hostHeaders[u.Host]; ok {
			req.Header = hostHeader
		}
		if insecure&discovery.InsecureTLS != 0 {
			res, err = discovery.ClientInsecureTLS.Do(req)
			return
		}
		res, err = discovery.Client.Do(req)
		return
	}
	closeBody := func(res *http.Response) {
		if res != nil {
			res.Body.Close()
		}
	}
	urlStr, res, err := fetch("https")
	if err != nil || res.StatusCode != http.StatusOK {
		if insecure&discovery.InsecureHTTP != 0 {
			closeBody(res)
			urlStr, res, err = fetch("http")
		}
	}

	if res != nil && res.StatusCode != http.StatusOK {
		err = fmt.Errorf("expected a 200 OK got %d", res.StatusCode)
	}

	if err != nil {
		closeBody(res)
		return "", nil, err
	}
	return urlStr, res.Body, nil
}
