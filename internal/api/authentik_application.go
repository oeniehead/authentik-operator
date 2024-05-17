package api

import (
	"goauthentik.io/api/v3"
)

func CreateApplication(cl *AuthentikApiClient, application *api.Application) (*api.Application, error) {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	request := api.ApplicationRequest{
		Name:     application.Name,
		Slug:     application.Slug,
		Group:    application.Group,
		Provider: application.Provider,
	}

	newApplication, _, err := apiClient.CoreApi.CoreApplicationsCreate(authCtx).ApplicationRequest(request).Execute()

	if err != nil {
		return nil, err
	}

	return newApplication, nil
}

func GetApplication(cl *AuthentikApiClient, name string) (*api.Application, error) {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	resp, _, err := apiClient.CoreApi.CoreApplicationsList(authCtx).Name(name).Execute()

	if err != nil {
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, nil
	} else {
		return &resp.Results[0], nil
	}
}

func DeleteApplication(cl *AuthentikApiClient, application string) error {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	existingApplication, err := GetApplication(cl, application)

	if err != nil {
		return err
	}

	if existingApplication == nil {
		return nil
	}

	_, err = apiClient.CoreApi.CoreApplicationsDestroy(authCtx, existingApplication.Slug).Execute()

	return err
}

func BindApplicationToGroup(cl *AuthentikApiClient, applicationId string, groupId string) error {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	request := api.PolicyBindingRequest{
		Target: applicationId,
		Group:  *api.NewNullableString(&groupId),
	}

	_, _, err := apiClient.PoliciesApi.PoliciesBindingsCreate(authCtx).PolicyBindingRequest(request).Execute()

	return err
}

func GetGroupBinding(cl *AuthentikApiClient, applicationId string, groupId string) (*api.PolicyBinding, error) {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	bindings, _, err := apiClient.PoliciesApi.PoliciesBindingsList(authCtx).Target(applicationId).Execute()
	if err != nil {
		return nil, err
	}

	for _, binding := range bindings.Results {
		bindingGroupId := binding.Group.Get()
		if *bindingGroupId == groupId {
			return &binding, nil
		}
	}

	return nil, err
}
