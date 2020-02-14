package trig

import (
	"fmt"
	"strconv"

	"github.com/godfried/osmimport/poi"
)

type Trig struct {
	Lat, Lon    float64
	Name        string
	Ele         float64
	Number      *BeaconNumber
	Description string
	CreatedBy   string
	OSMID       uint64
	tags        map[string]string
}

type BeaconNumber struct {
	Area   int
	Number int
}

func (b BeaconNumber) String() string {
	return fmt.Sprintf("%d-%d", b.Area, b.Number)
}

func (t Trig) Latitude() float64 {
	return t.Lat
}

func (t Trig) Longitude() float64 {
	return t.Lon
}

func (t Trig) Names() []poi.Name {
	return []poi.Name{{Key: poi.NameKeyDefault, Value: t.Name}}
}
func (t *Trig) AddTag(key, value string) {
	if t.tags == nil {
		t.tags = make(map[string]string, 8)
	}
	t.tags[key] = value
}

func (t Trig) Tags() map[string]string {
	tags := map[string]string{
		"man_made": "survey_point",
		"ref":      t.Number.String(),
		"source":   "ngi",
	}
	for k, v := range t.tags {
		tags[k] = v
	}
	for _, n := range t.Names() {
		tags[string(n.Key)] = n.Value
	}
	if t.Ele != 0 {
		tags["ele"] = strconv.FormatFloat(t.Ele, 'f', -1, 64)
	}
	if t.Description != "" {
		tags["description"] = t.Description
	}
	return tags
}

func (t Trig) OSMFilter() []poi.Attribute {
	return []poi.Attribute{{Key: "ref", Value: t.Number.String()}}
}

func (t Trig) String() string {
	return fmt.Sprintf("%v", t.Tags())
}
