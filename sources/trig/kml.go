package trig

import "encoding/xml"

type KML struct {
	XMLName  xml.Name `xml:"kml"`
	Text     string   `xml:",chardata"`
	XMLNS    string   `xml:"xmlns,attr"`
	GX       string   `xml:"gx,attr"`
	KML      string   `xml:"kml,attr"`
	Atom     string   `xml:"atom,attr"`
	Document Document `xml:"Document"`
}

type Document struct {
	Text     string     `xml:",chardata"`
	Name     string     `xml:"name"`
	LookAt   LookAt     `xml:"LookAt"`
	Style    []Style    `xml:"Style"`
	StyleMap []StyleMap `xml:"StyleMap"`
	Folder   Folder     `xml:"Folder"`
}

type LookAt struct {
	Text         string `xml:",chardata"`
	Longitude    string `xml:"longitude"`
	Latitude     string `xml:"latitude"`
	Altitude     string `xml:"altitude"`
	Heading      string `xml:"heading"`
	Tilt         string `xml:"tilt"`
	Range        string `xml:"range"`
	AltitudeMode string `xml:"altitudeMode"`
}

type Style struct {
	Text       string     `xml:",chardata"`
	ID         string     `xml:"id,attr"`
	IconStyle  IconStyle  `xml:"IconStyle"`
	LabelStyle LabelStyle `xml:"LabelStyle"`
}

type LabelStyle struct {
	Text  string `xml:",chardata"`
	Scale string `xml:"scale"`
}

type IconStyle struct {
	Text  string `xml:",chardata"`
	Color string `xml:"color"`
	Scale string `xml:"scale"`
	Icon  Icon   `xml:"Icon"`
}

type Icon struct {
	Text string `xml:",chardata"`
	Href string `xml:"href"`
}

type StyleMap struct {
	Text string `xml:",chardata"`
	ID   string `xml:"id,attr"`
	Pair []Pair `xml:"Pair"`
}

type Pair struct {
	Text     string `xml:",chardata"`
	Key      string `xml:"key"`
	StyleUrl string `xml:"styleUrl"`
}

type Folder struct {
	Text      string      `xml:",chardata"`
	Name      string      `xml:"name"`
	Region    Region      `xml:"Region"`
	Placemark []Placemark `xml:"Placemark"`
}

type Region struct {
	Text         string       `xml:",chardata"`
	LatLonAltBox LatLonAltBox `xml:"LatLonAltBox"`
	Lod          Lod          `xml:"Lod"`
}

type Placemark struct {
	Text        string `xml:",chardata"`
	Name        string `xml:"name"`
	Description string `xml:"description"`
	StyleUrl    string `xml:"styleUrl"`
	Point       Point  `xml:"Point"`
}

type Point struct {
	Text        string `xml:",chardata"`
	Coordinates string `xml:"coordinates"`
}

type Lod struct {
	Text          string `xml:",chardata"`
	MinLodPixels  string `xml:"minLodPixels"`
	MaxLodPixels  string `xml:"maxLodPixels"`
	MinFadeExtent string `xml:"minFadeExtent"`
	MaxFadeExtent string `xml:"maxFadeExtent"`
}

type LatLonAltBox struct {
	Text        string `xml:",chardata"`
	North       string `xml:"north"`
	South       string `xml:"south"`
	East        string `xml:"east"`
	West        string `xml:"west"`
	MinAltitude string `xml:"minAltitude"`
	MaxAltitude string `xml:"maxAltitude"`
}
