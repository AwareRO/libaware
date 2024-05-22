package geoip

type Location struct {
	Status      string  `json:"status"`
	Latitude    float64 `json:"lat"`
	Longitude   float64 `json:"lon"`
	Country     string  `json:"country"`
	CountryCode string  `json:"countryCode"`
	Region      string  `json:"region"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Isp         string  `json:"isp"`
	Org         string  `json:"org"`
	As          string  `json:"as"`
}

type Find func(ip string) (*Location, error)
