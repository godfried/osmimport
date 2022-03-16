package sagns

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/godfried/osmimport/poi"

	"strconv"

	"log"
)

// Format is:
// Name, Feature_Description, pklid, Latitude, Longitude, Date, MapInfo, Province, fklFeatureSubTypeID,
// Previous_Name, fklMagisterialDistrictID, ProvinceID, fklLanguageID, fklDisteral, Local Municipality,
// Sound, District Municipality, fklLocalMunic, Comments, Meaning
func Read(inputFile string) ([]*SAGNSPOI, error) {
	f, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = 20
	pois := make([]*SAGNSPOI, 0, 1000)
	_, err = r.Read()
	if err != nil {
		return nil, err
	}
	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		p, err := NewPOI(record)
		if err != nil {
			log.Printf("could not parse geoname '%s': %s", record, err)
			continue
		}
		pois = append(pois, p)
	}
	return pois, nil
}

func NewPOI(record []string) (*SAGNSPOI, error) {
	id, err := strconv.ParseUint(record[2], 10, 64)
	if err != nil {
		log.Print(err)
	}
	lat, err := strconv.ParseFloat(record[3], 64)
	if err != nil {
		return nil, err
	}
	lon, err := strconv.ParseFloat(record[4], 64)
	if err != nil {
		return nil, err
	}
	date, err := time.Parse("02-01-2006", record[5])
	if err != nil {
		log.Print(err)
	}
	names := make([]poi.Name, 0, 2)
	switch {
	case record[0] != "":
		names = append(names, poi.Name{Key: poi.NameKeyDefault, Value: record[0]})
	case record[9] != "":
		names = append(names, poi.Name{Key: poi.NameKeyOld, Value: record[9]})
	}
	f := Feature(strings.ReplaceAll(strings.ToLower(record[1]), " ", "_"))
	if len(f.OSMTags()) == 0 {
		return nil, fmt.Errorf("no tags available for feature %s", f)
	}
	return &SAGNSPOI{
		names:                names,
		feature:              f,
		id:                   uint32(id),
		latitude:             lat,
		longitude:            lon,
		date:                 date,
		mapInfo:              record[6],
		province:             record[7],
		localMunicipality:    record[14],
		districtMunicipality: record[16],
		comments:             record[18],
		meaning:              record[19],
	}, nil
}

type SAGNSPOI struct {
	names                []poi.Name // 0
	feature              Feature    // 1
	id                   uint32     // 2
	latitude             float64    // 3
	longitude            float64    // 4
	date                 time.Time  // 5
	mapInfo              string     // 6
	province             string     // 7
	localMunicipality    string     // 14
	districtMunicipality string     // 16
	comments             string     // 18
	meaning              string     // 19
	tags                 map[string]string
}

func (s SAGNSPOI) Latitude() float64 {
	return s.latitude
}

func (s *SAGNSPOI) AddTag(key, value string) {
	if s.tags == nil {
		s.tags = make(map[string]string, 8)
	}
	s.tags[key] = value
}

func (s SAGNSPOI) Longitude() float64 {
	return s.longitude
}

func (s SAGNSPOI) Names() []poi.Name {
	return s.names
}

func (s SAGNSPOI) ID() uint32 {
	return s.id
}

func (s SAGNSPOI) Feature() Feature {
	return s.feature
}
func (s SAGNSPOI) Tags() map[string]string {
	tags := s.feature.OSMTags()
	if s.comments != "" {
		tags["description"] = s.meaning
	}
	tags["sagns_id"] = strconv.FormatUint(uint64(s.id), 10)
	tags["source"] = "sagns"
	for _, n := range s.Names() {
		tags[string(n.Key)] = n.Value
	}
	for k, v := range s.tags {
		tags[k] = v
	}
	return tags
}

func (s SAGNSPOI) OSMFilter() []poi.Attribute {
	return []poi.Attribute{{Key: "sagns_id", Value: strconv.FormatUint(uint64(s.id), 10)}} //s.feature.filter()
}

func (s SAGNSPOI) String() string {
	return fmt.Sprintf("%v", s.Tags())
}

type Feature string

const (
	FeatureAgriVillage = Feature("agrivillage")
	FeatureAirfield    = Feature("airfield")
	FeatureAirport     = Feature("airport")
	FeatureArea        = Feature("area")
	FeatureBattlefield = Feature("battlefield")
	FeatureBay = Feature("bay")
	FeatureBeach = Feature("beach")
	Feature = Feature("border_post")
	FeatureBowLake = Feature("bow_lake")
	Feature = Feature("")
	Feature = Feature("")
	Feature = Feature("")
	Feature = Feature("")
	Feature = Feature("")
	Feature = Feature("")
	Feature = Feature("")
case "Bow Lake":
	return map[string]string{
		"natural": "water",
		"water":   "oxbow",
	}
case "Brickworks":
	return map[string]string{
		"industrial": "brickyard",
	}
case "Bridge":
	return map[string]string{
		"bridge": "yes",
	}
case "Bush Area":
	return map[string]string{
		"natural": "scrub",
	}
case "Canal":
	return map[string]string{
		"waterway": "canal",
	}
case "Cemetery":
	return map[string]string{
		"landuse": "cemetery",
	}
case "Cliff":
	return map[string]string{
		"natural": "cliff",
	}
case "Coastal Rock":
	return map[string]string{
		"natural": "bare_rock",
	}
case "Coastline":
	return map[string]string{}
case "Coastline_Beach":
	return map[string]string{
		"natural": "beach",
	}
case "College":
	return map[string]string{}
case "Cove":
	return map[string]string{
		"natural": "bay",
	}
case "Dam":
	return map[string]string{
		"natural": "water",
		"water":   "reservoir",
	}
case "Dam Wall":
	return map[string]string{
		"waterway": "dam",
	}
case "Dock":
	return map[string]string{"waterway": "dock"}
case "Double Non Perennial":
	return map[string]string{
		"waterway":     "river",
		"intermittent": "yes",
	}
case "Double Perennial":
	return map[string]string{
		"waterway":     "river",
		"intermittent": "no",
	}
case "Drift":
	return map[string]string{}
case "Dry":
	return map[string]string{
		"natural": "desert",
	}
case "Dry Area":
	return map[string]string{
		"natural": "desert",
	}
case "Dry Water Course":
	return map[string]string{
		"waterway":     "river",
		"intermittent": "yes",
	}
case "Forest":
	return map[string]string{
		"natural": "wood",
	}
case "Furrow":
	return map[string]string{
		"waterway": "ditch",
	}
case "Game Reserve":
	return map[string]string{
		"boundary":      "protected_area",
		"landuse":       "conservation",
		"protect_class": "1",
	}
case "Gorge":
	return map[string]string{
		"natural": "stream",
	}
case "Group of huts":
	return map[string]string{
		"place": "hamlet",
	}
case "Guard Post":
	return map[string]string{
		"barrier": "border_control",
	}
case "Harbour":
	return map[string]string{
		"harbour": "yes",
	}
case "Heritage resource":
	return map[string]string{}
case "Hill":
	return map[string]string{
		"natural": "peak",
	}
case "Historical":
	return map[string]string{
		"historic": "yes",
	}
case "Holy Grave":
	return map[string]string{
		"historic": "tomb",
	}
case "Hospital":
	return map[string]string{
		"amenity": "hospital",
	}
case "Hotel":
	return map[string]string{
		"tourism": "hotel",
	}
case "Industrial":
	return map[string]string{
		"landuse": "industrial",
	}
case "Interchange":
	return map[string]string{
		"highway": "motorway junction",
	}
case "Island":
	return map[string]string{
		"place": "island",
	}
case "Island Real":
	return map[string]string{
		"place": "island",
	}
case "Junction":
	return map[string]string{
		"highway": "motorway junction",
	}
case "Kloof":
	return map[string]string{
		"waterway": "stream",
	}
case "Kop":
	return map[string]string{
		"natural": "peak",
	}
case "Lagoon":
	return map[string]string{
		"natural": "water",
		"water":   "lagoon",
	}
case "Lake":
	return map[string]string{
		"natural": "water",
		"water":   "lake",
	}
case "Lake Vlei":
	return map[string]string{
		"natural": "water",
		"water":   "lake",
	}
case "Land Development":
	return map[string]string{
		"landuse": "construction",
	}
case "Landing Strip":
	return map[string]string{
		"aeroway":        "aerodrome",
		"aerodrome:type": "airfield",
	}
case "Lighthouse/Marine_Beacon":
	return map[string]string{
		"man_made": "lighthouse",
	}
case "Main":
	return map[string]string{}
case "Marsh Vlei":
	return map[string]string{
		"natural": "wetland",
		"wetland": "marsh",
	}
case "Mission":
	return map[string]string{
		"place": "hamlet",
	}
case "Mountain":
	return map[string]string{
		"natural": "peak",
	}
case "Mountain Peak":
	return map[string]string{
		"natural": "peak",
	}
case "Mountain Range":
	return map[string]string{
		"natural": "mountain_range",
	}
case "Mouth":
	return map[string]string{
		"waterway": "river",
	}
case "Museum":
	return map[string]string{
		"tourism": "museum",
	}
case "Nature Reserve":
	return map[string]string{
		"boundary":      "protected_area",
		"landuse":       "conservation",
		"protect_class": "1",
	}
case "Non Perennial":
	return map[string]string{
		"waterway":     "river",
		"intermittent": "yes",
	}
case "Non_Perennial":
	return map[string]string{
		"waterway":     "river",
		"intermittent": "yes",
	}
case "Observatory":
	return map[string]string{
		"landuse":  "observatory",
		"man_made": "telescope",
	}
case "Ocean":
	return map[string]string{}
case "Other":
	return map[string]string{}
case "Pan":
	return map[string]string{
		"natural": "desert",
	}
case "Pass":
	return map[string]string{
		"mountain_pass": "yes",
	}
case "Pass Neks":
	return map[string]string{
		"natural": "saddle",
	}
case "Patrol Post":
	return map[string]string{
		"barrier": "border_control",
	}
case "Peak":
	return map[string]string{
		"natural": "peak",
	}
case "Perennial":
	return map[string]string{
		"waterway":     "river",
		"intermittent": "no",
	}
case "Plain":
	return map[string]string{
		"natural": "grassland",
	}
case "Plantation":
	return map[string]string{
		"landuse": "forest",
	}
case "Plateau":
	return map[string]string{
		"natural": "plateau",
	}
case "Police_Station":
	return map[string]string{
		"amenity": "police",
	}
case "Post Office":
	return map[string]string{
		"amenity": "post_office",
	}
case "Power station":
	return map[string]string{
		"power": "substation",
	}
case "Prison":
	return map[string]string{
		"amenity": "prison",
	}
case "Protected Area":
	return map[string]string{
		"boundary":      "protected_area",
		"landuse":       "conservation",
		"protect_class": "1",
	}
case "Quarry":
	return map[string]string{
		"landuse": "quarry",
	}
case "Railway":
	return map[string]string{
		"railway": "rail",
	}
case "Railway Station":
	return map[string]string{
		"railway": "station",
	}
case "Railway Tunnel":
	return map[string]string{
		"railway": "rail",
		"tunnel":  "yes",
	}
case "Research Centre":
	return map[string]string{
		"amenity": "research_institute",
	}
case "Research Institute":
	return map[string]string{
		"amenity": "research_institute",
	}
case "Residential Town":
	return map[string]string{
		"place": "town",
	}
case "Residential Township":
	return map[string]string{
		"place": "town",
	}
case "Ridge":
	return map[string]string{
		"natural": "ridge",
	}
case "River (not specified)":
	return map[string]string{
		"waterway": "river",
	}
case "River Bend":
	return map[string]string{
		"waterway": "river",
	}
case "Road":
	return map[string]string{
		"highway": "unclassified",
	}
case "Rock":
	return map[string]string{
		"natural": "bare_rock",
	}
case "Rock Outcrop":
	return map[string]string{
		"natural": "bare_rock",
	}
case "Ruin":
	return map[string]string{
		"historic": "ruin",
	}
case "Sandy Area":
	return map[string]string{
		"natural": "sand",
	}
case "Sawmill":
	return map[string]string{
		"craft": "sawmill",
	}
case "School":
	return map[string]string{
		"school": "yes",
	}
case "Settlement":
	return map[string]string{
		"place": "village",
	}
case "Single Non Perennial":
	return map[string]string{
		"waterway":     "river",
		"intermittent": "yes",
	}
case "Single Perennial":
	return map[string]string{
		"waterway":     "river",
		"intermittent": "no",
	}
case "Siphon":
	return map[string]string{
		"waterway": "canal",
	}
case "Spa":
	return map[string]string{
		"amenity":   "public_bath",
		"bath:type": "thermal",
	}
case "State":
	return map[string]string{
		"landuse": "forest",
	}
case "Station":
	return map[string]string{
		"railway": "station",
	}
case "Studam":
	return map[string]string{
		"waterway": "weir",
	}
case "Tower":
	return map[string]string{
		"man_made": "tower",
	}
case "Town":
	return map[string]string{
		"place": "town",
	}
case "Township":
	return map[string]string{
		"place": "town",
	}
case "Trail Hiking":
	return map[string]string{
		"highway": "path",
	}
case "Tunnel":
	return map[string]string{
		"tunnel": "yes",
	}
case "Urban Area":
	return map[string]string{
		"place": "suburb",
	}
case "Valley":
	return map[string]string{
		"natural": "valley",
	}
case "Village":
	return map[string]string{
		"place": "village",
	}
case "Village Settlement":
	return map[string]string{
		"place": "village",
	}
case "Water":
	return map[string]string{
		"natural": "water",
	}
case "Weir":
	return map[string]string{
		"waterway": "weir",
	}
case "Yard":
	return map[string]string{}
	FeatureAgriVillage = Feature("agrivillage")
	FeatureAgriVillage = Feature("agrivillage")
)

func (f Feature) filter() []poi.Attribute {
	tags := f.OSMTags()
	filter := make([]poi.Attribute, 0, len(tags)+1)
	for k, v := range tags {
		switch k {
		case "waterway":
			filter = append(filter,
				poi.Attribute{Key: "waterway", Value: "river"},
				poi.Attribute{Key: "waterway", Value: "stream"},
				poi.Attribute{Key: "waterway", Value: "canal"},
				poi.Attribute{Key: "waterway", Value: "ditch"},
			)
		case "place":
			filter = append(filter,
				poi.Attribute{Key: "place", Value: "island"},
				poi.Attribute{Key: "place", Value: "suburb"},
				poi.Attribute{Key: "place", Value: "town"},
				poi.Attribute{Key: "place", Value: "village"},
				poi.Attribute{Key: "place", Value: "hamlet"},
			)
		default:
			filter = append(filter, poi.Attribute{Key: k, Value: v})
		}
	}
	return filter
}

func (f Feature) OSMTags() map[string]string {
	switch string(f) {
	case "Agrivillage":
		return map[string]string{
			"place": "village",
		}
	case "Airfield":
		return map[string]string{
			"aeroway":        "aerodrome",
			"aerodrome:type": "airfield",
		}
	case "Airport":
		return map[string]string{
			"aeroway": "aerodrome",
		}
	case "Area":
		return map[string]string{}
	case "Battlefield":
		return map[string]string{
			"historic": "battlefield",
		}
	case "Bay":
		return map[string]string{
			"natural": "bay",
		}
	case "Beach":
		return map[string]string{
			"natural": "beach",
		}
	case "Border Post":
		return map[string]string{
			"barrier": "border_control",
		}
	case "Bow Lake":
		return map[string]string{
			"natural": "water",
			"water":   "oxbow",
		}
	case "Brickworks":
		return map[string]string{
			"industrial": "brickyard",
		}
	case "Bridge":
		return map[string]string{
			"bridge": "yes",
		}
	case "Bush Area":
		return map[string]string{
			"natural": "scrub",
		}
	case "Canal":
		return map[string]string{
			"waterway": "canal",
		}
	case "Cemetery":
		return map[string]string{
			"landuse": "cemetery",
		}
	case "Cliff":
		return map[string]string{
			"natural": "cliff",
		}
	case "Coastal Rock":
		return map[string]string{
			"natural": "bare_rock",
		}
	case "Coastline":
		return map[string]string{}
	case "Coastline_Beach":
		return map[string]string{
			"natural": "beach",
		}
	case "College":
		return map[string]string{}
	case "Cove":
		return map[string]string{
			"natural": "bay",
		}
	case "Dam":
		return map[string]string{
			"natural": "water",
			"water":   "reservoir",
		}
	case "Dam Wall":
		return map[string]string{
			"waterway": "dam",
		}
	case "Dock":
		return map[string]string{"waterway": "dock"}
	case "Double Non Perennial":
		return map[string]string{
			"waterway":     "river",
			"intermittent": "yes",
		}
	case "Double Perennial":
		return map[string]string{
			"waterway":     "river",
			"intermittent": "no",
		}
	case "Drift":
		return map[string]string{}
	case "Dry":
		return map[string]string{
			"natural": "desert",
		}
	case "Dry Area":
		return map[string]string{
			"natural": "desert",
		}
	case "Dry Water Course":
		return map[string]string{
			"waterway":     "river",
			"intermittent": "yes",
		}
	case "Forest":
		return map[string]string{
			"natural": "wood",
		}
	case "Furrow":
		return map[string]string{
			"waterway": "ditch",
		}
	case "Game Reserve":
		return map[string]string{
			"boundary":      "protected_area",
			"landuse":       "conservation",
			"protect_class": "1",
		}
	case "Gorge":
		return map[string]string{
			"natural": "stream",
		}
	case "Group of huts":
		return map[string]string{
			"place": "hamlet",
		}
	case "Guard Post":
		return map[string]string{
			"barrier": "border_control",
		}
	case "Harbour":
		return map[string]string{
			"harbour": "yes",
		}
	case "Heritage resource":
		return map[string]string{}
	case "Hill":
		return map[string]string{
			"natural": "peak",
		}
	case "Historical":
		return map[string]string{
			"historic": "yes",
		}
	case "Holy Grave":
		return map[string]string{
			"historic": "tomb",
		}
	case "Hospital":
		return map[string]string{
			"amenity": "hospital",
		}
	case "Hotel":
		return map[string]string{
			"tourism": "hotel",
		}
	case "Industrial":
		return map[string]string{
			"landuse": "industrial",
		}
	case "Interchange":
		return map[string]string{
			"highway": "motorway junction",
		}
	case "Island":
		return map[string]string{
			"place": "island",
		}
	case "Island Real":
		return map[string]string{
			"place": "island",
		}
	case "Junction":
		return map[string]string{
			"highway": "motorway junction",
		}
	case "Kloof":
		return map[string]string{
			"waterway": "stream",
		}
	case "Kop":
		return map[string]string{
			"natural": "peak",
		}
	case "Lagoon":
		return map[string]string{
			"natural": "water",
			"water":   "lagoon",
		}
	case "Lake":
		return map[string]string{
			"natural": "water",
			"water":   "lake",
		}
	case "Lake Vlei":
		return map[string]string{
			"natural": "water",
			"water":   "lake",
		}
	case "Land Development":
		return map[string]string{
			"landuse": "construction",
		}
	case "Landing Strip":
		return map[string]string{
			"aeroway":        "aerodrome",
			"aerodrome:type": "airfield",
		}
	case "Lighthouse/Marine_Beacon":
		return map[string]string{
			"man_made": "lighthouse",
		}
	case "Main":
		return map[string]string{}
	case "Marsh Vlei":
		return map[string]string{
			"natural": "wetland",
			"wetland": "marsh",
		}
	case "Mission":
		return map[string]string{
			"place": "hamlet",
		}
	case "Mountain":
		return map[string]string{
			"natural": "peak",
		}
	case "Mountain Peak":
		return map[string]string{
			"natural": "peak",
		}
	case "Mountain Range":
		return map[string]string{
			"natural": "mountain_range",
		}
	case "Mouth":
		return map[string]string{
			"waterway": "river",
		}
	case "Museum":
		return map[string]string{
			"tourism": "museum",
		}
	case "Nature Reserve":
		return map[string]string{
			"boundary":      "protected_area",
			"landuse":       "conservation",
			"protect_class": "1",
		}
	case "Non Perennial":
		return map[string]string{
			"waterway":     "river",
			"intermittent": "yes",
		}
	case "Non_Perennial":
		return map[string]string{
			"waterway":     "river",
			"intermittent": "yes",
		}
	case "Observatory":
		return map[string]string{
			"landuse":  "observatory",
			"man_made": "telescope",
		}
	case "Ocean":
		return map[string]string{}
	case "Other":
		return map[string]string{}
	case "Pan":
		return map[string]string{
			"natural": "desert",
		}
	case "Pass":
		return map[string]string{
			"mountain_pass": "yes",
		}
	case "Pass Neks":
		return map[string]string{
			"natural": "saddle",
		}
	case "Patrol Post":
		return map[string]string{
			"barrier": "border_control",
		}
	case "Peak":
		return map[string]string{
			"natural": "peak",
		}
	case "Perennial":
		return map[string]string{
			"waterway":     "river",
			"intermittent": "no",
		}
	case "Plain":
		return map[string]string{
			"natural": "grassland",
		}
	case "Plantation":
		return map[string]string{
			"landuse": "forest",
		}
	case "Plateau":
		return map[string]string{
			"natural": "plateau",
		}
	case "Police_Station":
		return map[string]string{
			"amenity": "police",
		}
	case "Post Office":
		return map[string]string{
			"amenity": "post_office",
		}
	case "Power station":
		return map[string]string{
			"power": "substation",
		}
	case "Prison":
		return map[string]string{
			"amenity": "prison",
		}
	case "Protected Area":
		return map[string]string{
			"boundary":      "protected_area",
			"landuse":       "conservation",
			"protect_class": "1",
		}
	case "Quarry":
		return map[string]string{
			"landuse": "quarry",
		}
	case "Railway":
		return map[string]string{
			"railway": "rail",
		}
	case "Railway Station":
		return map[string]string{
			"railway": "station",
		}
	case "Railway Tunnel":
		return map[string]string{
			"railway": "rail",
			"tunnel":  "yes",
		}
	case "Research Centre":
		return map[string]string{
			"amenity": "research_institute",
		}
	case "Research Institute":
		return map[string]string{
			"amenity": "research_institute",
		}
	case "Residential Town":
		return map[string]string{
			"place": "town",
		}
	case "Residential Township":
		return map[string]string{
			"place": "town",
		}
	case "Ridge":
		return map[string]string{
			"natural": "ridge",
		}
	case "River (not specified)":
		return map[string]string{
			"waterway": "river",
		}
	case "River Bend":
		return map[string]string{
			"waterway": "river",
		}
	case "Road":
		return map[string]string{
			"highway": "unclassified",
		}
	case "Rock":
		return map[string]string{
			"natural": "bare_rock",
		}
	case "Rock Outcrop":
		return map[string]string{
			"natural": "bare_rock",
		}
	case "Ruin":
		return map[string]string{
			"historic": "ruin",
		}
	case "Sandy Area":
		return map[string]string{
			"natural": "sand",
		}
	case "Sawmill":
		return map[string]string{
			"craft": "sawmill",
		}
	case "School":
		return map[string]string{
			"school": "yes",
		}
	case "Settlement":
		return map[string]string{
			"place": "village",
		}
	case "Single Non Perennial":
		return map[string]string{
			"waterway":     "river",
			"intermittent": "yes",
		}
	case "Single Perennial":
		return map[string]string{
			"waterway":     "river",
			"intermittent": "no",
		}
	case "Siphon":
		return map[string]string{
			"waterway": "canal",
		}
	case "Spa":
		return map[string]string{
			"amenity":   "public_bath",
			"bath:type": "thermal",
		}
	case "State":
		return map[string]string{
			"landuse": "forest",
		}
	case "Station":
		return map[string]string{
			"railway": "station",
		}
	case "Studam":
		return map[string]string{
			"waterway": "weir",
		}
	case "Tower":
		return map[string]string{
			"man_made": "tower",
		}
	case "Town":
		return map[string]string{
			"place": "town",
		}
	case "Township":
		return map[string]string{
			"place": "town",
		}
	case "Trail Hiking":
		return map[string]string{
			"highway": "path",
		}
	case "Tunnel":
		return map[string]string{
			"tunnel": "yes",
		}
	case "Urban Area":
		return map[string]string{
			"place": "suburb",
		}
	case "Valley":
		return map[string]string{
			"natural": "valley",
		}
	case "Village":
		return map[string]string{
			"place": "village",
		}
	case "Village Settlement":
		return map[string]string{
			"place": "village",
		}
	case "Water":
		return map[string]string{
			"natural": "water",
		}
	case "Weir":
		return map[string]string{
			"waterway": "weir",
		}
	case "Yard":
		return map[string]string{}
	default:
		log.Printf("unknown feature %s", f)
		return map[string]string{}
	}
}
