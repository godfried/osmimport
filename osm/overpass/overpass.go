package overpass

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/godfried/osmimport/poi"
)

var endpoints = []string{
	"https://overpass.kumi.systems/api/interpreter",
	"https://lz4.overpass-api.de/api/interpreter",
	"https://z.overpass-api.de/api/interpreter",
}

type Result struct {
	Elements []*Element
}

type Element struct {
	Type      string            `json:"type"`
	ID        uint64            `json:"id"`
	Lat       float64           `json:"lat"`
	Lon       float64           `json:"lon"`
	Timestamp string            `json:"timestamp"`
	Version   uint32            `json:"version"`
	Changeset uint64            `json:"changeset"`
	User      string            `json:"user"`
	UID       uint64            `json:"uid"`
	TagMap    map[string]string `json:"tags"`
}

func (e Element) String() string {
	return fmt.Sprintf("%#v", e)
}

func (e Element) Names() []poi.Name {
	names := make([]poi.Name, 0, len(e.TagMap))
	for key, val := range e.TagMap {
		if strings.Contains(key, "name") {
			names = append(names, poi.Name{Key: poi.NameKey(key), Value: val})
		}
	}
	return names
}

func (e *Element) Tags() map[string]string {
	return e.TagMap
}

func (e *Element) Latitude() float64 {
	return e.Lat
}

func (e *Element) Longitude() float64 {
	return e.Lon
}

func loadNode(query string) (*http.Response, error) {
	var resp *http.Response
	var err error
	vals := url.Values{"data": []string{query}}
	c := &http.Client{Timeout: 20 * time.Second}
	for _, endpoint := range endpoints {
		resp, err = c.PostForm(endpoint, vals)
		if err != nil {
			log.Printf("Could not load node: %s", err)
			continue
		}
		if resp.StatusCode != 200 {
			resp.Body.Close()
			err = fmt.Errorf("bad status code: %s", resp.Status)
			continue
		}
		return resp, nil
	}
	return nil, err
}

type NodeLoader func(poi OSMPOI, dist float64) (*Element, error)

type OSMPOI interface {
	poi.POI
	OSMFilter() (key, value string)
}

func LoadNearestNode(p OSMPOI, dist float64) (*Element, error) {
	nodes, err := loadNodes(p, dist)
	if err != nil {
		return nil, err
	}
	nearest := poi.SelectNearest(nodes, p, dist)
	if nearest == nil {
		return nil, nil
	}
	return nearest.(*Element), nil

}

func LoadMatchingNode(p OSMPOI, dist float64) (*Element, error) {
	nodes, err := loadNodes(p, dist)
	if err != nil {
		return nil, err
	}
	match := poi.SelectMatch(nodes, p, dist)
	if match == nil {
		log.Printf("No match found for %s", p.Names()[0].Value)
		return nil, nil
	}
	log.Printf("match found %#v", match)
	return match.(*Element), nil
}

func HasMatches(p OSMPOI, dist float64) bool {
	m, err := LoadMatchingNode(p, dist)
	if err != nil {
		log.Printf("error loading matches: %s", err)
		return false
	}
	return m != nil
}

const query = `
[out:json][timeout:{{.Timeout}}];
(
	node["{{.Key}}"="{{.Value}}"](around:{{.Radius}},{{.Latitude}},{{.Longitude}});
	way["{{.Key}}"="{{.Value}}"](around:{{.Radius}},{{.Latitude}},{{.Longitude}});
	relation["{{.Key}}"="{{.Value}}"](around:{{.Radius}},{{.Latitude}},{{.Longitude}});
);
out meta;
`

const timeoutSeconds = 20

type queryParam struct {
	Timeout   int
	Key       string
	Value     string
	Radius    float64
	Latitude  float64
	Longitude float64
}

func buildQuery(p OSMPOI, dist float64) (string, error) {
	k, v := p.OSMFilter()
	arg := queryParam{
		Timeout:   timeoutSeconds,
		Key:       k,
		Value:     v,
		Radius:    dist,
		Latitude:  p.Latitude(),
		Longitude: p.Longitude(),
	}
	var b strings.Builder
	err := template.Must(template.New("query").Parse(query)).Execute(&b, arg)
	return b.String(), err
}

func loadNodes(p OSMPOI, dist float64) ([]poi.POI, error) {
	q, err := buildQuery(p, dist)
	if err != nil {
		return nil, err
	}
	name := p.Names()[0].Value
	log.Printf("Loading nodes for %s using query: %s", name, q)
	resp, err := loadNode(q)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	result := new(Result)
	err = json.Unmarshal(body, result)
	if err != nil {
		return nil, err
	}
	if len(result.Elements) == 0 {
		return nil, nil
	}
	pois := make([]poi.POI, 0, len(result.Elements))
	for _, e := range result.Elements {
		pois = append(pois, e)
	}
	log.Printf("Got %d matches for %s", len(pois), name)
	return pois, nil
}
