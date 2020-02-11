package main

import (
	"flag"
	"log"
	"os"

	"github.com/godfried/osmimport/osm"
	"github.com/godfried/osmimport/osm/overpass"
	"github.com/godfried/osmimport/poi"
)

func main() {
	log.SetOutput(os.Stdout)
	query := `
[out:json];
(
	way["sagnsid"](-35.42486791930557,16.34765625,-22.91792293614603,31.0166015625);
	node["sagnsid"](-35.42486791930557,16.34765625,-22.91792293614603,31.0166015625);
	relation["sagnsid"](-35.42486791930557,16.34765625,-22.91792293614603,31.0166015625);
);
out meta;
`
	out := flag.String("out", "sagns-id-fix.xml", "path to output file")
	flag.Parse()
	es, err := overpass.RunQuery(query)
	if err != nil {
		log.Fatal(err)
	}
	pois := make([]poi.POI, 0, len(es))
	for _, e := range es {
		v := e.TagMap["sagnsid"]
		delete(e.TagMap, "sagnsid")
		e.TagMap["sagns_id"] = v
		pois = append(pois, e)
	}
	err = osm.GenerateUpdateXML(pois, *out)
	if err != nil {
		log.Fatal(err)
	}
}
