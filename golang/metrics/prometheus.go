package metrics

type PrometheusResponseMetric struct {
	App       string `json:"app"`
	Country   string `json:"country"`
	Crawler   string `json:"crawler"`
	Endpoint  string `json:"endpoint"`
	Latitude  string `json:"lat"`
	Longitude string `json:"lon"`
	Method    string `json:"method"`
	Status    string `json:"status"`
	IP        string `json:"ip"`
}

type PrometheusResponseResult struct {
	Metric PrometheusResponseMetric `json:"metric"`
	Value  []interface{}            `json:"value"`
}

type PrometheusResponseData struct {
	Result []PrometheusResponseResult `json:"result"`
}

type PrometheusResponse struct {
	Status string                 `json:"success"`
	Data   PrometheusResponseData `json:"data"`
}
