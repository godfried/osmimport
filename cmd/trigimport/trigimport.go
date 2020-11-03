package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/godfried/osmimport/osm"
	"github.com/godfried/osmimport/osm/overpass"
	"github.com/godfried/osmimport/poi"
	"github.com/godfried/osmimport/sources/trig"
)

func main() {
	log.SetOutput(os.Stdout)
	bbox := poi.CircleBox{}
	limit := flag.Int("limit", 10, "number of points to export")
	out := flag.String("out", fmt.Sprintf("trig-poi-%s.xml", time.Now().Format(time.RFC3339)), "path to output file")
	flag.Float64Var(&bbox.Lat, "lat", -33.4, "latitude around which to focus")
	flag.Float64Var(&bbox.Lon, "lon", 20.0, "longitude around which to focus")
	flag.Float64Var(&bbox.RadiusKM, "radius", 20, "radius around centre to select points from, in km")
	flag.Parse()
	err := run(bbox, *limit, *out)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(bbox poi.CircleBox, limit int, out string) error {
	db, err := trig.Connect()
	if err != nil {
		return err
	}
	defer db.Close()
	pois, err := db.QueryIncomplete(100000)
	if err != nil {
		return err
	}
	log.Printf("loaded %d POIs", len(pois))
	boundedPOIs := make([]poi.POI, 0, len(pois))
	q := overpass.BuildQuery([]poi.Attribute{{Key: "man_made", Value: "survey_point"}}, bbox.RadiusKM*1000, bbox.Lat, bbox.Lon)
	results, err := overpass.RunQuery(q)
	if err != nil {
		return err
	}
	resultMap := make(map[string]*overpass.Element, len(results))
	for _, r := range results {
		ref, ok := r.TagMap["ref"]
		if !ok {
			continue
		}
		if exists, ok := resultMap[ref]; ok {
			var ename, rname string
			for _, n := range exists.Names() {
				ename = n.Value
				break
			}
			for _, n := range r.Names() {
				rname = n.Value
				break
			}
			log.Printf("ref %s already exists: %s %f %f and %s %f %f", ref, rname, r.Lat, r.Lon, ename, exists.Lat, exists.Lon)
			continue
		}
		resultMap[ref] = r
	}
	wg := &sync.WaitGroup{}
	defer wg.Wait()
	for _, p := range pois {
		if len(boundedPOIs) >= limit {
			break
		}
		if !bbox.Contains(p) {
			continue
		}

		match, ok := resultMap[p.Number.String()]
		if ok && poi.Distance(match.Latitude(), match.Longitude(), p.Latitude(), p.Longitude()) < 1000 {
			wg.Add(1)
			go update(p, match.ID, db, wg)
			continue
		}
		potentialMatch, ok := resultMap[strconv.Itoa(p.Number.Number)]
		if ok && poi.Distance(potentialMatch.Latitude(), potentialMatch.Longitude(), p.Latitude(), p.Longitude()) < 1000 {
			log.Printf("adding fixme to %s", p.Name)
			p.AddTag("fixme", "check existing survey_point")
		}
		boundedPOIs = append(boundedPOIs, p)
		log.Printf("selected trig beacon %s:%s (total %d)", p.Name, p.Number, len(boundedPOIs))
	}
	log.Printf("filtered to %d POIs", len(boundedPOIs))
	return osm.GenerateXML(boundedPOIs, out)
}

func update(t *trig.Trig, id uint64, db *trig.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	t.OSMID = id
	err := db.UpdateOSMID(t)
	if err != nil {
		log.Print(err)
	} else {
		log.Printf("Added OSMID %d for %s:%s", t.OSMID, t.Name, t.Number)
	}
}
