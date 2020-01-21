package osm

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	nominatimFromID     = "https://nominatim.openstreetmap.org/reverse?format=json&osm_type=N&osm_id=%d&accept-language=en"
	nominatimFromLatLon = "https://nominatim.openstreetmap.org/reverse?format=json&osm_type=N&lat=%f&lon=%f&accept-language=en"
)

type Place struct {
	PlaceID     uint64
	OSMType     string
	OSMID       uint64
	Lat         float64 `json:",string"`
	Lon         float64 `json:",string"`
	DisplayName string
	Address     Address
	BoundingBox []string
}

type Address struct {
	City          string
	PostCode      string
	Country       string
	CountryCode   string
	County        string
	StateDistrict string
	State         string
	Peak          string
}

func (a Address) String() string {
	return fmt.Sprintf("%#v", a)
}

func (a Address) Area() string {
	if a.State != "" {
		return a.State
	}
	if a.County != "" {
		return a.County
	}
	if a.StateDistrict != "" {
		return a.StateDistrict
	}
	if a.City != "" {
		return a.City
	}
	return "Unknown"
}

func PlaceFromNode(nodeID uint64) (*Place, error) {
	return getPlace(fmt.Sprintf(nominatimFromID, nodeID))
}

func PlaceFromCoords(lat, lon float64) (*Place, error) {
	return getPlace(fmt.Sprintf(nominatimFromLatLon, lat, lon))
}

func getPlace(nominatimURL string) (*Place, error) {
	c := &http.Client{Timeout: 20 * time.Second}
	resp, err := c.Get(nominatimURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Nominatim query failed with status=%s", resp.Status)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	place := new(Place)
	err = json.Unmarshal(body, place)
	if err != nil {
		return nil, err
	}
	return place, nil
}
