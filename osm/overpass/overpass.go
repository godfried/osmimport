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
	//	err := enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "modify"}})
	//	if err != nil {
	//		return err
	//	}
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
	//if err != nil {
	//	return err
	//}

	//return enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "modify"}})
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

func runQuery(query string) (*http.Response, error) {
	var resp *http.Response
	var err error
	vals := url.Values{"data": []string{query}}
	c := &http.Client{Timeout: 180 * time.Second}
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

func LoadNearestNode(p OSMPOI, dist float64) (*Element, error) {
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

func LoadMatchingNode(p OSMPOI, dist float64) (*Element, error) {
	nodes, err := loadMatches(p, dist)
	if err != nil {
		return nil, err
	}
	match := poi.SelectMatch(nodes, p)
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

func buildQuery(p OSMPOI, dist float64) (string, error) {
	filters := p.OSMFilter()
	arg := queryParam{
		Timeout:   timeoutSeconds,
		Filters:   filters,
		Radius:    dist,
		Latitude:  p.Latitude(),
		Longitude: p.Longitude(),
	}
	var b strings.Builder
	err := template.Must(template.New("query").Parse(query)).Execute(&b, arg)
	return b.String(), err
}

func loadMatches(p OSMPOI, dist float64) ([]poi.POI, error) {
	q, err := buildQuery(p, dist)
	if err != nil {
		return nil, err
	}
	name := p.Names()[0].Value
	log.Printf("Loading nodes for %s using query: %s", name, q)
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
