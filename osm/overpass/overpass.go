package overpass

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
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
	Type      string
	ID        uint64
	Lat       float64
	Lon       float64
	Timestamp string
	Version   uint32
	Changeset uint64
	User      string
	UID       uint64
	TagMap    map[string]string
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
	log.Printf("Loading nodes for query: %s", query)
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
		return nil, nil
	}
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

const (
	nodeQuery = "[out:json];node[\"%s\"~\"%s\"](around:%f,%f,%f);out meta;"
	//nodeQuery = "[out:json];node[\"%s\"~\"%s\"](around:%f,%f,%f);out meta;"
	//nodeQuery = "[out:json];node[\"%s\"~\"%s\"](around:%f,%f,%f);out meta;"
)

func loadNodes(p OSMPOI, dist float64) ([]poi.POI, error) {
	k, v := p.OSMFilter()
	query := fmt.Sprintf(nodeQuery, k, v, dist, p.Latitude(), p.Longitude())
	resp, err := loadNode(query)
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
	return pois, nil
}
