package osm

import (
	"encoding/xml"

	"math"

	"io/ioutil"
	"os"

	"github.com/godfried/osmimport/poi"
)

type Node struct {
	XMLName xml.Name `xml:"node"`
	Lat     float64  `xml:"lat,attr"`
	Lon     float64  `xml:"lon,attr"`
	ID      int      `xml:"id,attr"`
	Visible bool     `xml:"visible,attr"`
	Tag     []Tag
}

type Bounds struct {
	MinLat float64 `xml:"minlat,attr"`
	MinLon float64 `xml:"minlon,attr"`
	MaxLat float64 `xml:"maxlat,attr"`
	MaxLon float64 `xml:"maxlon,attr"`
	Origin string  `xml:"origin,attr"`
}

func NewBounds() *Bounds {
	return &Bounds{
		MinLon: math.MaxFloat64,
		MinLat: math.MaxFloat64,
		MaxLat: -math.MaxFloat64,
		MaxLon: -math.MaxFloat64,
		Origin: "OpenStreetMap server",
	}
}

func (b *Bounds) Expand(p poi.POI) {
	if b.MinLon > p.Longitude() {
		b.MinLon = p.Longitude()
	}
	if b.MinLat > p.Latitude() {
		b.MinLat = p.Latitude()
	}
	if b.MaxLon < p.Longitude() {
		b.MaxLon = p.Longitude()
	}
	if b.MaxLat < p.Latitude() {
		b.MaxLat = p.Latitude()
	}
}

type OSM struct {
	XMLName   xml.Name `xml:"osm"`
	Version   string   `xml:"version,attr"`
	Generator string   `xml:"generator,attr"`
	Node      []*Node
	Bounds    *Bounds `xml:"bounds"`
}

func NewOSM(size int) *OSM {
	return &OSM{
		Version:   "0.6",
		Generator: "JOSM",
		Bounds:    NewBounds(),
		Node:      make([]*Node, 0, size),
	}
}

func NewNode(p poi.POI, id int) *Node {
	poiTags := p.Tags()
	tags := make([]Tag, 0, len(poiTags))
	for k, v := range poiTags {
		tags = append(tags, Tag{Key: k, Value: v})
	}
	return &Node{Lat: p.Latitude(), Lon: p.Longitude(), Tag: tags, Visible: true, ID: id}
}

type Tag struct {
	XMLName xml.Name `xml:"tag"`
	Key     string   `xml:"k,attr"`
	Value   string   `xml:"v,attr"`
}

func GenerateXML(pois []poi.POI, outFile string) error {
	o := NewOSM(len(pois))
	for i, p := range pois {
		o.Node = append(o.Node, NewNode(p, -1*(i+1)))
		o.Bounds.Expand(p)
	}
	data, err := xml.MarshalIndent(o, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(outFile, data, os.ModePerm)
}

func GenerateUpdateXML(pois []poi.POI, outFile string) error {
	o := NewOSMUpdate(pois)
	data, err := xml.MarshalIndent(o, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(outFile, data, os.ModePerm)
}

type OSMUpdate struct {
	XMLName   xml.Name `xml:"osm"`
	Version   string   `xml:"version,attr"`
	Generator string   `xml:"generator,attr"`
	POIs      []poi.POI
	Bounds    *Bounds `xml:"bounds"`
}

func NewOSMUpdate(pois []poi.POI) *OSMUpdate {
	b := NewBounds()
	for _, p := range pois {
		b.Expand(p)
	}
	return &OSMUpdate{
		Version:   "0.6",
		Generator: "JOSM",
		Bounds:    b,
		POIs:      pois,
	}
}

type Modify struct {
	POIs []poi.POI
}
