package poi

import "math"

func hsin(theta float64) float64 {
	return math.Pow(math.Sin(theta/2), 2)
}

func Distance(lat1, lon1, lat2, lon2 float64) float64 {
	var la1, lo1, la2, lo2 float64
	la1 = lat1 * math.Pi / 180
	lo1 = lon1 * math.Pi / 180
	la2 = lat2 * math.Pi / 180
	lo2 = lon2 * math.Pi / 180

	const r = 6378100.0

	h := hsin(la2-la1) + math.Cos(la1)*math.Cos(la2)*hsin(lo2-lo1)
	return 2 * r * math.Asin(math.Sqrt(h))
}

type BBox struct {
	MinLat, MaxLat float64
	MinLon, MaxLon float64
}

func (b BBox) IsZero() bool {
	return b.MinLat == 0 && b.MaxLat == 0 && b.MinLon == 0 && b.MaxLon == 0
}

func (b BBox) Contains(p POI) bool {
	if b.IsZero() {
		return true
	}
	return p.Latitude() <= b.MaxLat && p.Latitude() >= b.MinLat && p.Longitude() <= b.MaxLon && p.Longitude() >= b.MinLon
}
