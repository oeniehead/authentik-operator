package api

import (
	"goauthentik.io/api/v3"
)

func CreateProvider(cl *AuthentikApiClient, provider *api.OAuth2Provider) (*api.OAuth2Provider, error) {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	// Ignore the type of the provider, we force it to oauth for now
	request := api.OAuth2ProviderRequest{
		Name:               provider.Name,
		AuthenticationFlow: provider.AuthenticationFlow,
		AuthorizationFlow:  provider.AuthorizationFlow,
		ClientType:         provider.ClientType,
		RedirectUris:       provider.RedirectUris,
		PropertyMappings:   provider.PropertyMappings,
	}

	newProvider, _, err := apiClient.ProvidersApi.ProvidersOauth2Create(authCtx).OAuth2ProviderRequest(request).Execute()

	if err != nil {
		return nil, err
	}

	return newProvider, nil
}

func GetProvider(cl *AuthentikApiClient, name string) (*api.OAuth2Provider, error) {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	resp, _, err := apiClient.ProvidersApi.ProvidersOauth2List(authCtx).Name(name).Execute()

	if err != nil {
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, nil
	} else {
		return &resp.Results[0], nil
	}
}

func GetScopeMapping(cl *AuthentikApiClient, name string) (*api.ScopeMapping, error) {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	resp, _, err := apiClient.PropertymappingsApi.PropertymappingsScopeList(authCtx).ScopeName(name).Execute()

	if err != nil {
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, nil
	} else {
		return &resp.Results[0], nil
	}
}

func GetFlow(cl *AuthentikApiClient, name string, flowType string) (*api.Flow, error) {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	resp, _, err := apiClient.FlowsApi.FlowsInstancesList(authCtx).Slug(name).Designation(flowType).Execute()

	if err != nil {
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, nil
	} else {
		return &resp.Results[0], nil
	}
}

func DeleteProvider(cl *AuthentikApiClient, provider string) error {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	existingProvider, err := GetProvider(cl, provider)

	if err != nil {
		return err
	}

	if existingProvider == nil {
		return nil
	}

	providerUUID := existingProvider.Pk

	_, err = apiClient.ProvidersApi.ProvidersOauth2Destroy(authCtx, providerUUID).Execute()

	return err
}
