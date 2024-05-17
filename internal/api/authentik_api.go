package api

import (
	"context"
	"goauthentik.io/api/v3"
	"os"
)

type AuthentikApiClient struct {
	ctx       context.Context
	apiClient *api.APIClient
}

func GetClient(ctx context.Context) AuthentikApiClient {
	configuration := api.NewConfiguration()
	configuration.Host = os.Getenv("AUTHENTIK_URL")
	configuration.Scheme = "https"
	apiClient := api.NewAPIClient(configuration)
	authCtx := context.WithValue(ctx, api.ContextAccessToken, os.Getenv("AUTHENTIK_TOKEN"))

	return AuthentikApiClient{
		ctx:       authCtx,
		apiClient: apiClient,
	}
}

func difference(slice1 []string, slice2 []string) ([]string, []string) {
	var diffleft []string
	var diffright []string

	for _, s1 := range slice1 {
		found := false
		for _, s2 := range slice2 {
			if s1 == s2 {
				found = true
				break
			}
		}
		if !found {
			diffleft = append(diffleft, s1)
		}
	}

	for _, s1 := range slice2 {
		found := false
		for _, s2 := range slice1 {
			if s1 == s2 {
				found = true
				break
			}
		}
		if !found {
			diffright = append(diffright, s1)
		}
	}

	return diffleft, diffright
}
