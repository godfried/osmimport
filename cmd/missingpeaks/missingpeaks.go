package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/godfried/osmimport/osm/overpass"
	"github.com/godfried/osmimport/poi"
	"github.com/godfried/osmimport/sources/wcpeaks"
)

func main() {
	bbox := poi.CircleBox{}
	peaksSource := flag.String("csv", "", "path to CSV with peaks data")
	flag.Float64Var(&bbox.Lat, "lat", -33.4, "latitude around which to focus")
	flag.Float64Var(&bbox.Lon, "lon", 20.0, "longitude around which to focus")
	flag.Float64Var(&bbox.RadiusKM, "radius", 500, "radius around centre to select points from, in km")
	flag.Parse()
	err := run(*peaksSource, bbox)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func run(peaksSource string, bbox poi.CircleBox) error {
	peaks, err := wcpeaks.Read(peaksSource)
	if err != nil {
		return err
	}
	q := buildQuery(bbox, query1600)
	results, err := overpass.RunQuery(q)
	if err != nil {
		return err
	}
	for _, r := range results {
		found := false
		if len(r.Names()) == 0 {
			continue
		}
		eleF, _ := strconv.ParseFloat(r.TagMap["ele"], 64)
		ele := int(eleF)
		partials := []*wcpeaks.Peak{}
	peakLoop:
		for _, p := range peaks {
			for _, n := range r.Names() {
				name := n.Value
				diff := ele - int(p.Ele)
				if diff < 0 {
					diff *= -1
				}
				if p.Name == name && diff < 50 {
					found = true
					break peakLoop
				} else if diff < 50 && (strings.Contains(p.Name, name) || strings.Contains(name, p.Name)) {
					partials = append(partials, p)
				}
			}
		}
		if found {
			continue
		}
		name := r.Names()[0].Value
		if len(partials) == 0 {
			log.Printf("No match for peak: %s %s %f %f", name, r.TagMap["ele"], r.Lat, r.Lon)
			log.Printf("https://htonl.dev.openstreetmap.org/ngi-tiles/#15/%f/%f", r.Lat, r.Lon)
		} else {
			log.Printf("partial matches for %s %s %f %f", name, r.TagMap["ele"], r.Lat, r.Lon)
			for _, p := range partials {
				log.Printf("%s %s %f", p.Name, p.Range, p.Ele)
			}
			log.Printf("https://htonl.dev.openstreetmap.org/ngi-tiles/#15/%f/%f", r.Lat, r.Lon)
		}
	}
	return nil
}

func buildQuery(bb poi.CircleBox, queryTemplate string) string {
	var b strings.Builder
	err := template.Must(template.New("query").Parse(queryTemplate)).Execute(&b, bb)
	if err != nil {
		panic(err)
	}
	return b.String()
}

const query1600 = `
[out:json];
(
	node[natural=peak][name!=""](around:{{.Radius}},{{.Lat}},{{.Lon}})
	(if:t["ele"] > 1599); 
);
out meta;
`
const query1000 = `
[out:json];
(
	node[natural=peak][name!=""](around:{{.Radius}},{{.Lat}},{{.Lon}})
	(if:t["ele"] < 1600 && t["ele] > 999); 
);
out meta;
`
