package sagns

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
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
	f := feature(record[1])
	if len(f.tags()) == 0 {
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
	feature              feature    // 1
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
}

func (s SAGNSPOI) Latitude() float64 {
	return s.latitude
}

func (s SAGNSPOI) Longitude() float64 {
	return s.longitude
}

func (s SAGNSPOI) Names() []poi.Name {
	return s.names
}

func (s SAGNSPOI) Tags() map[string]string {
	tags := s.feature.tags()
	//tags["start_date"] = s.date.Format("2006-01-02")
	if s.comments != "" {
		tags["description"] = s.comments
	}
	tags["sagnsid"] = strconv.FormatUint(uint64(s.id), 10)
	tags["source"] = "sagns"
	for _, n := range s.Names() {
		tags[string(n.Key)] = n.Value
	}
	return tags
}

func (s SAGNSPOI) OSMFilter() (key, value string) {
	return s.feature.filter()
}

func (s SAGNSPOI) String() string {
	return fmt.Sprintf("%v", s.Tags())
}

type feature string

func (f feature) filter() (key, value string) {
	switch string(f) {
	case "Agrivillage":
		return "place", "village"
	case "Airfield":
		return "aeroway", "aerodrome"
	case "Airport":
		return "aeroway", "aerodrome"
	case "Area":
		return "", ""
	case "Battlefield":
		return "historic", "battlefield"
	case "Bay":
		return "natural", "bay"
	case "Beach":
		return "natural", "beach"
	case "Border Post":
		return "barrier", "border_control"
	case "Bow Lake":
		return "natural", "water"
	case "Brickworks":
		return "industrial", "brickyard"
	case "Bridge":
		return "bridge", "yes"
	case "Bush Area":
		return "natural", "scrub"
	case "Canal":
		return "waterway", "canal"
	case "Cemetery":
		return "landuse", "cemetery"
	case "Cliff":
		return "natural", "cliff"
	case "Coastal Rock":
		return "natural", "bare_rock"
	case "Coastline":
		return "", ""
	case "Coastline_Beach":
		return "natural", "beach"
	case "College":
		return "", ""
	case "Cove":
		return "natural", "bay"
	case "Dam":
		return "natural", "water"
	case "Dam Wall":
		return "waterway", "dam"
	case "Dock":
		return "waterway", "dock"
	case "Double Non Perennial":
		return "waterway", "river"
	case "Double Perennial":
		return "waterway", "river"
	case "Drift":
		return "", ""
	case "Dry":
		return "natural", "desert"
	case "Dry Area":
		return "natural", "desert"
	case "Dry Water Course":
		return "waterway", "river"
	case "Forest":
		return "natural", "wood"
	case "Furrow":
		return "waterway", "ditch"
	case "Game Reserve":
		return "boundary", "protected_area"
	case "Gorge":
		return "natural", "gorge"
	case "Group of huts":
		return "place", "hamlet"
	case "Guard Post":
		return "barrier", "border_control"
	case "Harbour":
		return "harbour", "yes"
	case "Heritage resource":
		return "historic", "yes"
	case "Hill":
		return "natural", "peak"
	case "Historical":
		return "historic", "yes"
	case "Holy Grave":
		return "historic", "tomb"
	case "Hospital":
		return "amenity", "hospital"
	case "Hotel":
		return "tourism", "hotel"
	case "Industrial":
		return "landuse", "industrial"
	case "Interchange":
		return "highway", "motorway junction"
	case "Island":
		return "place", "island"
	case "Island Real":
		return "place", "island"
	case "Junction":
		return "highway", "motorway junction"
	case "Kloof":
		return "natural", "gorge"
	case "Kop":
		return "natural", "peak"
	case "Lagoon":
		return "natural", "water"
	case "Lake":
		return "natural", "water"
	case "Lake Vlei":
		return "natural", "water"
	case "Land Development":
		return "landuse", "construction"
	case "Landing Strip":
		return "aeroway", "aerodrome"
	case "Lighthouse/Marine_Beacon":
		return "man_made", "lighthouse"
	case "Main":
		return "", ""
	case "Marsh Vlei":
		return "natural", "wetland"
	case "Mission":
		return "place", "hamlet"
	case "Mountain":
		return "natural", "peak"
	case "Mountain Peak":
		return "natural", "peak"
	case "Mountain Range":
		return "natural", "mountain_range"
	case "Mouth":
		return "natural", "water"
	case "Museum":
		return "tourism", "museum"
	case "Nature Reserve":
		return "boundary", "protected_area"
	case "Non Perennial":
		return "waterway", "river"
	case "Non_Perennial":
		return "waterway", "river"
	case "Observatory":
		return "landuse", "observatory"
	case "Ocean":
		return "", ""
	case "Other":
		return "", ""
	case "Pan":
		return "natural", "desert"
	case "Pass":
		return "mountain_pass", "yes"
	case "Pass Neks":
		return "mountain_pass", "yes"
	case "Patrol Post":
		return "barrier", "border_control"
	case "Peak":
		return "natural", "peak"
	case "Perennial":
		return "waterway", "river"
	case "Plain":
		return "natural", "grassland"
	case "Plantation":
		return "landuse", "forest"
	case "Plateau":
		return "natural", "plateau"
	case "Police_Station":
		return "amenity", "police"
	case "Post Office":
		return "amenity", "post_office"
	case "Power station":
		return "power", "substation"
	case "Prison":
		return "amenity", "prison"
	case "Protected Area":
		return "boundary", "protected_area"
	case "Quarry":
		return "landuse", "quarry"
	case "Railway":
		return "railway", "rail"
	case "Railway Station":
		return "railway", "station"
	case "Railway Tunnel":
		return "railway", "rail"
	case "Research Centre":
		return "amenity", "research_institute"
	case "Research Institute":
		return "amenity", "research_institute"
	case "Residential Town":
		return "place", "town"
	case "Residential Township":
		return "place", "town"
	case "Ridge":
		return "natural", "ridge"
	case "River (not specified)":
		return "waterway", "river"
	case "River Bend":
		return "natural", "water"
	case "Road":
		return "highway", "unclassified"
	case "Rock":
		return "natural", "bare_rock"
	case "Rock Outcrop":
		return "natural", "bare_rock"
	case "Ruin":
		return "historic", "ruin"
	case "Sandy Area":
		return "natural", "sand"
	case "Sawmill":
		return "craft", "sawmill"
	case "School":
		return "school", "yes"
	case "Settlement":
		return "place", "village"
	case "Single Non Perennial":
		return "waterway", "river"
	case "Single Perennial":
		return "waterway", "river"
	case "Siphon":
		return "waterway", "canal"
	case "Spa":
		return "amenity", "public_bath"
	case "State":
		return "landuse", "forest"
	case "Station":
		return "railway", "station"
	case "Studam":
		return "waterway", "weir"
	case "Tower":
		return "man_made", "tower"
	case "Town":
		return "place", "town"
	case "Township":
		return "place", "town"
	case "Trail Hiking":
		return "highway", "path"
	case "Tunnel":
		return "tunnel", "yes"
	case "Urban Area":
		return "place", "suburb"
	case "Valley":
		return "natural", "valley"
	case "Village":
		return "place", "village"
	case "Village Settlement":
		return "place", "village"
	case "Water":
		return "natural", "water"
	case "Weir":
		return "waterway", "weir"
	case "Yard":
		return "", ""
	default:
		log.Printf("unknown feature %s", f)
		return "", ""
	}
}

func (f feature) tags() map[string]string {
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
			"natural": "gorge",
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
			"natural": "gorge",
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
			"natural": "water",
			"water":   "river",
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
			"mountain_pass": "yes",
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
			"natural": "water",
			"water":   "river",
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
