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
	limit := flag.Int("limit", 100, "number of points to process")
	flag.Parse()
	err := run(*sagnsSource, *out, bbox, *limit)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func run(sagnsSource, out string, bbox poi.CircleBox, limit int) error {
	pois, err := sagns.Read(sagnsSource)
	if err != nil {
		return err
	}
	q := overpass.BuildQuery(overpass.SAGNSQuery, nil, bbox.RadiusKM*1000, bbox.Lat, bbox.Lon)
	results, err := overpass.RunQuery(q)
	if err != nil {
		return err
	}
	resultMap := make(map[string]*overpass.Element, len(results))
	for _, r := range results {
		resultMap[r.TagMap["sagns_id"]] = r
	}
	log.Printf("Loaded %d sagns datapoints", len(resultMap))
	log.Printf("loaded %d POIs", len(pois))
	boundedPOIs := make([]poi.POI, 0, limit)
	for _, p := range pois {
		if len(boundedPOIs) >= limit {
			break
		}
		_, ok := resultMap[p.Tags()["sagns_id"]]
		if bbox.Contains(p) && !ok {
			boundedPOIs = append(boundedPOIs, p)
		}
	}
	log.Printf("filtered to %d POIs", len(boundedPOIs))
	return osm.GenerateXML(boundedPOIs, out)
}
