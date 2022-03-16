package geonames

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"log"
)

type GeoName struct {
	Name  string
	Type  string
	Lat   float64
	Lon   float64
	ID    uint64
	OSMID uint64
}

func (g GeoName) String() string {
	return fmt.Sprintf("GeoName{Name: %s; Latitude: %f; Longitude: %f}", g.Name, g.Lat, g.Lon)
}

func (g GeoName) Latitude() float64 {
	return g.Lat
}

func (g GeoName) Longitude() float64 {
	return g.Lon
}

func (g GeoName) Names() []string {
	return []string{g.Name}
}

func (g GeoName) OSMTags() map[string]string {
	tags := map[string]string{
		"name": g.Name,
	}
	if g.ID != 0 {
		tags["sagns_id"] = strconv.FormatUint(g.ID, 10)
	}
	return tags
}

func ReadGeoNames(inputFile string, bounds common.BBox, types map[string]struct{}) (GeoNames, error) {
	f, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = '\t'
	geoNames := make(GeoNames, 0, 1000)
	header, err := r.Read()
	if err != nil {
		return nil, err
	}
	latIndex := indexOf(header, "LAT")
	lonIndex := indexOf(header, "LONG")
	typeIndex := indexOf(header, "DSG")
	nameIndex := indexOf(header, "FULL_NAME_RO")
	for {
		vals, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		g, err := NewGeoName(vals[latIndex], vals[lonIndex], vals[nameIndex], vals[typeIndex])
		if err != nil {
			log.Printf("could not parse GeoName record '%s': %s", vals, err)
			continue
		}
		if _, ok := types[strings.ToLower(g.Type)]; !ok || !bounds.Contains(g.Lat, g.Lon) {
			continue
		}
		geoNames = append(geoNames, g)
	}
	return geoNames, nil
}

func NewGeoName(latStr, lonStr, name, tipe string) (*GeoName, error) {
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return nil, err
	}
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return nil, err
	}
	return &GeoName{Name: name, Type: strings.ToLower(tipe), Lat: lat, Lon: lon}, nil
}

func indexOf(vals []string, search string) int {
	search = strings.ToLower(search)
	for i, v := range vals {
		if strings.ToLower(strings.TrimSpace(v)) == search {
			return i
		}
	}
	panic(fmt.Errorf("%s not found in %s", search, vals))
}

type GeoNames []*GeoName

func (gs GeoNames) FindNearby(lat, lon, radius float64) GeoPoints {
	nearby := make(GeoPoints, 0, len(gs)/10)
	for _, g := range gs {
		dist := common.Distance(g.Lat, g.Lon, lat, lon)
		if dist < radius {
			nearby = append(nearby, &GeoPoint{g, dist})
		}
	}
	sort.Sort(nearby)
	return nearby
}

func (gs GeoNames) ToPOIs() []common.POI {
	pois := make([]common.POI, 0, len(gs))
	for _, g := range gs {
		pois = append(pois, g)
	}
	return pois
}

type GeoPoint struct {
	*GeoName
	Distance float64
}

type GeoPoints []*GeoPoint

func (gs GeoPoints) Less(i, j int) bool {
	return gs[i].Distance < gs[j].Distance
}

func (gs GeoPoints) Swap(i, j int) {
	gs[i], gs[j] = gs[j], gs[i]
}

func (gs GeoPoints) Len() int {
	return len(gs)
}
