package geoip

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

const (
	ipAPISuccess = "success"
	src          = "http://ip-api.com/json"
	ipAPIMethod  = "GET"
)

func NewIPApiFinder() Find {
	return func(ip string) (*Location, error) {
		logger := log.Error().Str("source", src).Str("method", ipAPIMethod)

		request, err := http.NewRequest(ipAPIMethod, fmt.Sprintf("%s/%s", src, ip), nil)
		if err != nil {
			logger.Err(err).Msg("Failed to make request")

			return nil, err
		}

		response, err := (&http.Client{}).Do(request)
		if err != nil {
			logger.Err(err).Msg("Failed to get data")

			return nil, err
		}
		defer response.Body.Close()

		buf, _ := io.ReadAll(response.Body)
		response.Body.Close()
		response.Body = io.NopCloser(bytes.NewBuffer(buf))

		loc := Location{}

		err = json.NewDecoder(response.Body).Decode(&loc)
		if err != nil || loc.Status != ipAPISuccess {
			logger.Err(err).Str("body", string(buf)).Msg("Failed to parse response")

			if err == nil {
				err = errors.New("ip-api error")
			}

			return nil, err
		}

		return &loc, nil
	}
}
