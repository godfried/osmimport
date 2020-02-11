package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/godfried/osmimport/osm"
	"github.com/godfried/osmimport/osm/overpass"
	"github.com/godfried/osmimport/poi"

	"github.com/godfried/osmimport/sources/sagns"
)

func main() {
	bbox := poi.CircleBox{}
	sagnsSource := flag.String("csv", "", "path to CSV with SAGNS data")
	flag.Float64Var(&bbox.Lat, "lat", -33.4, "latitude around which to focus")
	flag.Float64Var(&bbox.Lon, "lon", 20.0, "longitude around which to focus")
	flag.Float64Var(&bbox.RadiusKM, "radius", 20, "radius around centre to select points from, in km")
	out := flag.String("out", fmt.Sprintf("sagns-poi-%s.xml", time.Now().Format(time.RFC3339)), "path to output file")
	flag.Parse()
	pois, err := sagns.Read(*sagnsSource)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	log.Printf("loaded %d POIs", len(pois))
	boundedPOIs := make([]poi.POI, 0, len(pois))
	for _, p := range pois {
		if len(boundedPOIs) >= 10 {
			break
		}
		if bbox.Contains(p) && !overpass.HasMatches(p, 2000) {
			boundedPOIs = append(boundedPOIs, p)
		}
	}
	log.Printf("filtered to %d POIs", len(boundedPOIs))
	err = osm.GenerateXML(boundedPOIs, *out)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
