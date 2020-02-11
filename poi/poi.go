package poi

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
)

type POI interface {
	fmt.Stringer
	Latitude() float64
	Longitude() float64
	Names() []Name
	Tags() map[string]string
}

type poiDists struct {
	pois      []POI
	distances []float64
}

var r = strings.NewReplacer(" ", "", "-", "")

func normaliseName(name Name) string {
	return strings.ToLower(r.Replace(name.Value))
}

func SelectMatch(pois []POI, poi POI) POI {
	log.Printf("selecting matches from %s", pois)
	if len(pois) == 0 {
		return nil
	}
	if len(poi.Names()) == 0 {
		return pois[0]
	}
	var contains, fuzz POI
	fuzzRatio := 50.0
	names := poi.Names()
	normalisedNames := make([]string, 0, len(names))
	for _, n := range names {
		normalised := normaliseName(n)
		normalisedNames = append(normalisedNames, normalised)
	}
outer:
	for _, p := range pois {
		for _, name := range p.Names() {
			normalisedName := normaliseName(name)
			for _, n := range normalisedNames {
				if strings.EqualFold(normalisedName, n) {
					return p
				}
				if strings.Contains(normalisedName, n) || strings.Contains(n, normalisedName) {
					contains = p
					break outer
				}
				ratio := LevenshteinRatio(normalisedName, n)
				if ratio < fuzzRatio {
					fuzzRatio = ratio
					fuzz = p
				}
			}
		}
	}
	if contains != nil {
		return contains
	}
	if fuzz == nil {
		return nil
	}
	return fuzz
}

func SelectNearest(pois []POI, poi POI, radius float64) POI {
	if len(pois) == 1 {
		return pois[0]
	}
	nearest := nearestPOIs(pois, poi, radius)
	if len(nearest) == 0 {
		return nil
	}
	return nearest[0]
}

func nearestPOIs(pois []POI, poi POI, maxDist float64) []POI {
	pds := &poiDists{pois: make([]POI, 0, len(pois)), distances: make([]float64, 0, len(pois))}
	for _, p := range pois {
		dist := Distance(p.Latitude(), p.Longitude(), poi.Latitude(), poi.Longitude())
		if dist < maxDist {
			pds.distances = append(pds.distances, dist)
			pds.pois = append(pds.pois, p)
		}
	}
	sort.Sort(pds)
	return pds.pois
}

func (p *poiDists) Less(i, j int) bool {
	return p.distances[i] < p.distances[j]
}

func (p *poiDists) Swap(i, j int) {
	p.distances[i], p.distances[j] = p.distances[j], p.distances[i]
	p.pois[i], p.pois[j] = p.pois[j], p.pois[i]
}

func (p *poiDists) Len() int {
	return len(p.distances)
}

type PartitionedPOIs map[int]map[int][]POI

func CreatePartitionedPOIs() PartitionedPOIs {
	return make(PartitionedPOIs, 10)
}

func (p PartitionedPOIs) Add(poi POI) {
	lat := int(math.Round(poi.Latitude()))
	lon := int(math.Round(poi.Longitude()))
	entry, ok := p[lat]
	if !ok {
		p[lat] = make(map[int][]POI)
		entry = p[lat]
	}
	entry[lon] = append(entry[lon], poi)
}

type Attribute struct {
	Key   string
	Value string
}

type Name struct {
	Key   NameKey
	Value string
}

type NameKey string

const (
	NameKeyDefault       NameKey = "name"
	NameKeyInternational NameKey = "int_name"
	NameKeyNational      NameKey = "nat_name"
	NameKeyLocal         NameKey = "loc_name"
	NameKeyOld           NameKey = "old_name"
	NameKeyAlternative   NameKey = "alt_name"
	NameKeyEnglish       NameKey = "name:en"
	NameKeyAfrikaans     NameKey = "name:af"
)
