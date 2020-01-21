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
	sagnsSource := flag.String("csv", "", "path to CSV with SAGNS data")
	out := flag.String("out", fmt.Sprintf("sagns-poi-%s.xml", time.Now().Format(time.RFC3339)), "path to output file")
	flag.Parse()
	bbox := poi.BBox{
		MinLat: -33.9,
		MaxLon: 19.229,
		MaxLat: -33.6,
		MinLon: 19.0,
	}
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
		if bbox.Contains(p) && !overpass.HasMatches(p, 1000) {
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
