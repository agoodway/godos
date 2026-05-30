package cmd

import (
	"github.com/goodway/godos/config"
	"github.com/goodway/godos/internal/todex"
)

func getAPIService(requireToken bool) (*todex.Service, error) {
	baseURL, err := config.APIBaseURL()
	if err != nil {
		return nil, err
	}

	token := ""
	if requireToken {
		token, err = config.APIToken()
		if err != nil {
			return nil, err
		}
	}

	return todex.New(todex.ServiceConfig{BaseURL: baseURL, Token: token})
}
