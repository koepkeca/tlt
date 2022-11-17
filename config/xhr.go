package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type XhrFile struct {
	Log struct {
		Version string `json:"version"`
		Creator struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"creator"`
		Browser struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		} `json:"browser"`
		Pages []struct {
			StartedDateTime string `json:"startedDateTime"`
			ID              string `json:"id"`
			Title           string `json:"title"`
			PageTimings     struct {
				OnContentLoad int `json:"onContentLoad"`
				OnLoad        int `json:"onLoad"`
			} `json:"pageTimings"`
		} `json:"pages"`
		Entries []struct {
			Pageref         string `json:"pageref"`
			StartedDateTime string `json:"startedDateTime"`
			Request         struct {
				BodySize    int    `json:"bodySize"`
				Method      string `json:"method"`
				URL         string `json:"url"`
				HTTPVersion string `json:"httpVersion"`
				Headers     []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"headers"`
				Cookies     []interface{} `json:"cookies"`
				QueryString []interface{} `json:"queryString"`
				HeadersSize int           `json:"headersSize"`
			} `json:"request"`
			Response struct {
				Status      int    `json:"status"`
				StatusText  string `json:"statusText"`
				HTTPVersion string `json:"httpVersion"`
				Headers     []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"headers"`
				Cookies []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"cookies"`
				Content struct {
					MimeType string `json:"mimeType"`
					Size     int    `json:"size"`
					Comment  string `json:"comment"`
				} `json:"content"`
				RedirectURL string `json:"redirectURL"`
				HeadersSize int    `json:"headersSize"`
				BodySize    int    `json:"bodySize"`
			} `json:"response"`
			Cache struct {
			} `json:"cache"`
			Timings struct {
				Blocked int `json:"blocked"`
				DNS     int `json:"dns"`
				Connect int `json:"connect"`
				Ssl     int `json:"ssl"`
				Send    int `json:"send"`
				Wait    int `json:"wait"`
				Receive int `json:"receive"`
			} `json:"timings"`
			Time            int    `json:"time"`
			SecurityState   string `json:"_securityState"`
			ServerIPAddress string `json:"serverIPAddress"`
			Connection      string `json:"connection"`
		} `json:"entries"`
	} `json:"log"`
}

func NewXhrFromReader(i io.Reader) (xhr XhrFile, e error) {
	buf, e := io.ReadAll(i)
	if e != nil {
		e = errors.New(fmt.Sprintf("config.NewXhrFromReader: error reading file [%s]", e))
		return
	}
	e = json.Unmarshal(buf, &xhr)
	if e != nil {
		e = errors.New(fmt.Sprintf("config.NewXhrFromReader: json unmarshal error (check file contents) [%s]", e))
	}
	return
}
