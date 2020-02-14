package trig

import (
	"encoding/gob"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

func ReadGOB(gobfile string) ([]*Trig, error) {
	f, err := os.Open(gobfile)
	if err != nil {
		return nil, err
	}
	dec := gob.NewDecoder(f)
	trigs := make([]*Trig, 0, 4096)
	read := 0
outer:
	for {
		read++
		t := new(Trig)
		err := dec.Decode(t)
		switch err {
		case io.EOF:
			break outer
		case nil:
			log.Printf("READ %s", t.Name)
			trigs = append(trigs, t)
		default:
			return nil, err
		}
	}
	return trigs, nil
}

func ReadFile(kmlFile string) ([]*Trig, error) {
	kml, err := ParseFile(kmlFile)
	if err != nil {
		return nil, err
	}
	return kml.ExtractPoints(), nil
}

func Read(r io.Reader) ([]*Trig, error) {
	kml, err := Parse(r)
	if err != nil {
		return nil, err
	}
	return kml.ExtractPoints(), nil
}

func Parse(r io.Reader) (*KML, error) {
	kml := new(KML)
	d := xml.NewDecoder(r)
	err := d.Decode(kml)
	if err != nil {
		return nil, err
	}
	return kml, nil
}

func ParseFile(kmlFile string) (*KML, error) {
	f, err := os.Open(kmlFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}

func (k KML) ExtractPoints() []*Trig {
	trigs := make([]*Trig, 0, len(k.Document.Folder.Placemark))
	for _, p := range k.Document.Folder.Placemark {
		t, err := newTrig(p)
		if err != nil {
			log.Print(err)
			continue
		}
		trigs = append(trigs, t)
	}
	return trigs
}

func newBeaconNumber(val string) (*BeaconNumber, error) {
	vals := strings.Split(val, "-")
	if len(vals) != 2 {
		return nil, fmt.Errorf("cannot parse beacon number: %s", val)
	}
	area, err := strconv.Atoi(vals[0])
	if err != nil {
		return nil, fmt.Errorf("cannot parse beacon number: %s: %s", val, err)
	}
	number, err := strconv.Atoi(vals[1])
	if err != nil {
		return nil, fmt.Errorf("cannot parse beacon number: %s: %s", val, err)
	}
	return &BeaconNumber{Area: area, Number: number}, nil
}

func newTrig(p Placemark) (*Trig, error) {
	t := &Trig{tags: make(map[string]string, 8)}
	var err error
	descs := strings.Split(strings.TrimSuffix(strings.TrimPrefix(p.Description, "<![CDATA["), "]]>"), "<br></br>")
	for _, d := range descs {
		vals := strings.Split(d, "=")
		if len(vals) < 2 {
			continue
		}
		k, v := strings.ToLower(strings.TrimSpace(vals[0])), strings.TrimSpace(vals[1])
		if v == "" {
			continue
		}
		switch k {
		case "name":
			t.Name = v
		case "latitude":
			t.Lat, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse latitude in: %#v: %s", p, err)
			}
		case "longitude":
			t.Lon, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse longitude in: %#v: %s", p, err)
			}
		case "ortho ht":
			t.Ele, err = strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("cannot parse elevation in: %#v: %s", p, err)
			}
		case "lo", "y", "x":
		case "beacon number":
			t.Number, err = newBeaconNumber(v)
			if err != nil {
				return nil, err
			}
		case "description":
			t.Description = v
		case "created by":
			t.CreatedBy = v
		default:
			panic("unhandled key: " + k)
		}
	}
	if t.Number == nil {
		return nil, fmt.Errorf("beacon number not set in: %#v", p)
	}
	lat, lon, err := parseCoords(p.Point.Coordinates)
	if err != nil {
		log.Print(err)
	} else {
		t.Lat = lat
		t.Lon = lon
	}
	return t, nil
}

func parseCoords(val string) (lat, lon float64, err error) {
	coords := strings.Split(val, ",")
	if len(coords) < 2 {
		return 0, 0, fmt.Errorf("cannot parse coordinates: %s", val)
	}
	lon, err = strconv.ParseFloat(coords[0], 64)
	if err != nil {
		return 0, 0, err
	}
	lat, err = strconv.ParseFloat(coords[1], 64)
	if err != nil {
		return 0, 0, err
	}
	return lat, lon, err
}
