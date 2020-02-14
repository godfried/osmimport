package overpass

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/godfried/osmimport/poi"
)

var endpoints = []string{
	"https://lz4.overpass-api.de/api/interpreter",
	"https://z.overpass-api.de/api/interpreter",
	"https://overpass.kumi.systems/api/interpreter",
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
	Nodes     []uint64          `json:"nodes"`
}

const elementXML = `
<{{.Type}} id="{{.ID}}">
{{- range .Nodes -}}
<nd ref="{{.}}"/>
{{- end -}}
{{- range $k, $v := .TagMap -}}
<tag k="{{$k}}" v="{{$v}}"/>
{{- end -}}
</{{.Type}}>
`

func (e Element) MarshalXML(enc *xml.Encoder, start xml.StartElement) error {
	attrs := []xml.Attr{
		{Name: xml.Name{Local: "id"}, Value: strconv.FormatUint(e.ID, 10)},
		{Name: xml.Name{Local: "version"}, Value: strconv.FormatUint(uint64(e.Version), 10)},
	}
	if e.Type == "node" {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "lat"}, Value: strconv.FormatFloat(e.Lat, 'f', -1, 64)})
		attrs = append(attrs, xml.Attr{Name: xml.Name{Local: "lon"}, Value: strconv.FormatFloat(e.Lon, 'f', -1, 64)})
	}
	err := enc.EncodeToken(xml.StartElement{
		Name: xml.Name{Local: e.Type},
		Attr: attrs,
	})
	if err != nil {
		return err
	}
	for _, n := range e.Nodes {
		err := enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "nd"}, Attr: []xml.Attr{{Name: xml.Name{Local: "ref"}, Value: strconv.FormatUint(n, 10)}}})
		if err != nil {
			return err
		}
		err = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "nd"}})
		if err != nil {
			return err
		}
	}
	for k, v := range e.TagMap {
		err := enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "tag"}, Attr: []xml.Attr{{Name: xml.Name{Local: "k"}, Value: k}, {Name: xml.Name{Local: "v"}, Value: v}}})
		if err != nil {
			return err
		}
		err = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "tag"}})
		if err != nil {
			return err
		}
	}
	return enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: e.Type}})
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

func (e *Element) AddTag(key, value string) {
	e.TagMap[key] = value
}

func runQuery(query string) (*http.Response, error) {
	var resp *http.Response
	var err error
	vals := url.Values{"data": []string{query}}
	c := &http.Client{Timeout: 180 * time.Second}
	retries := 50
	for i := 0; i < retries; i++ {
		for _, endpoint := range endpoints {
			resp, err = c.PostForm(endpoint, vals)
			if err != nil {
				log.Printf("Could not load elements: %s", err)
				continue
			}
			if resp.StatusCode != 200 {
				resp.Body.Close()
				err = fmt.Errorf("bad status code: %s", resp.Status)
				continue
			}
			return resp, nil
		}
		time.Sleep(10 * time.Second)
	}
	return nil, err
}

func loadElements(query string) ([]*Element, error) {
	resp, err := runQuery(query)
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
	return result.Elements, nil
}

type NodeLoader func(poi OSMPOI, dist float64) (*Element, error)

type OSMPOI interface {
	poi.POI
	OSMFilter() []poi.Attribute
}

func LoadNearestElement(p OSMPOI, dist float64) (*Element, error) {
	nodes, err := loadMatches(p, dist)
	if err != nil {
		return nil, err
	}
	nearest := poi.SelectNearest(nodes, p, dist)
	if nearest == nil {
		return nil, nil
	}
	return nearest.(*Element), nil

}

func LoadMatchingElement(p OSMPOI, dist float64) (*Element, error) {
	matches, err := loadMatches(p, dist)
	if err != nil {
		return nil, err
	}
	match := poi.SelectMatch(matches, p)
	if match == nil {
		log.Printf("No match found for %s", p.Names()[0].Value)
		return nil, nil
	}
	return match.(*Element), nil
}

func HasMatches(p OSMPOI, dist float64) bool {
	m, err := LoadMatchingElement(p, dist)
	if err != nil {
		log.Printf("error loading matches: %s", err)
		return false
	}
	return m != nil
}

func RunQuery(query string) ([]*Element, error) {
	return loadElements(query)
}

const query = `
[out:json][timeout:{{.Timeout}}];
(
{{- $radius := .Radius -}}
{{- $latitude := .Latitude -}}
{{- $longitude := .Longitude -}}
{{range .Filters}}
	node["{{.Key}}"="{{.Value}}"](around:{{$radius}},{{$latitude}},{{$longitude}});
	way["{{.Key}}"="{{.Value}}"](around:{{$radius}},{{$latitude}},{{$longitude}});
	relation["{{.Key}}"="{{.Value}}"](around:{{$radius}},{{$latitude}},{{$longitude}});
{{- end -}}
);
out meta;
`

const timeoutSeconds = 20

type queryParam struct {
	Timeout   int
	Filters   []poi.Attribute
	Value     string
	Radius    float64
	Latitude  float64
	Longitude float64
}

func buildQuery(p OSMPOI, dist float64) string {
	return BuildQuery(p.OSMFilter(), dist, p.Latitude(), p.Longitude())
}

func BuildQuery(filters []poi.Attribute, radius, lat, lon float64) string {
	arg := queryParam{
		Timeout:   timeoutSeconds,
		Filters:   filters,
		Radius:    radius,
		Latitude:  lat,
		Longitude: lon,
	}
	var b strings.Builder
	err := template.Must(template.New("query").Parse(query)).Execute(&b, arg)
	if err != nil {
		panic(err)
	}
	return b.String()
}

func loadMatches(p OSMPOI, dist float64) ([]poi.POI, error) {
	q := buildQuery(p, dist)
	es, err := loadElements(q)
	if err != nil {
		return nil, err
	}
	pois := make([]poi.POI, 0, len(es))
	for _, e := range es {
		pois = append(pois, e)
	}
	return pois, nil
}
