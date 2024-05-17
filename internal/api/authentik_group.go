package api

import (
	"fmt"
	"goauthentik.io/api/v3"
)

func RemoveFromGroup(cl *AuthentikApiClient, user *api.User, groupName string) error {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	group, err := GetGroup(cl, groupName)

	if err != nil {
		return err
	}

	userAccountRequest := api.NewUserAccountRequest(user.Pk)

	_, err = apiClient.CoreApi.CoreGroupsRemoveUserCreate(authCtx, group.Pk).UserAccountRequest(*userAccountRequest).Execute()

	return err
}

func AddToGroup(cl *AuthentikApiClient, user *api.User, groupName string) error {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	group, err := GetGroup(cl, groupName)

	if err != nil {
		return err
	}

	if group == nil {
		return fmt.Errorf("group %s not found", groupName)
	}

	userAccountRequest := api.NewUserAccountRequest(user.Pk)

	_, err = apiClient.CoreApi.CoreGroupsAddUserCreate(authCtx, group.Pk).UserAccountRequest(*userAccountRequest).Execute()

	return err
}

func GetGroup(cl *AuthentikApiClient, name string) (*api.Group, error) {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	resp, _, err := apiClient.CoreApi.CoreGroupsList(authCtx).Name(name).Execute()

	if err != nil {
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, nil
	} else {
		return &resp.Results[0], nil
	}
}

func GetGroupById(cl *AuthentikApiClient, id string) (*api.Group, error) {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	resp, _, err := apiClient.CoreApi.CoreGroupsRetrieve(authCtx, id).Execute()

	if err != nil {
		return nil, err
	}

	if resp == nil {
		return nil, nil
	} else {
		return resp, nil
	}
}

func CreateGroup(cl *AuthentikApiClient, group *api.Group) (*api.Group, error) {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	existingGroup, err := GetGroup(cl, group.Name)

	if err != nil {
		return nil, err
	}

	if existingGroup == nil {
		createRequest := api.NewGroupRequest(group.Name)
		createRequest.SetIsSuperuser(*group.IsSuperuser)

		if group.Parent.Get() != nil {
			parentGroupName := group.Parent.Get()
			parentGroup, err := GetGroup(cl, *parentGroupName)

			if err != nil {
				return nil, err
			}

			createRequest.SetParent(parentGroup.Pk)
		}

		resp, _, err := apiClient.CoreApi.CoreGroupsCreate(authCtx).GroupRequest(*createRequest).Execute()

		if err != nil {
			return nil, err
		} else {
			return resp, nil
		}
	} else {
		return existingGroup, nil
	}
}

func DeleteGroup(cl *AuthentikApiClient, group string) error {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	existingGroup, err := GetGroup(cl, group)

	if err != nil {
		return err
	}

	if existingGroup == nil {
		return nil
	}

	groupUUID := existingGroup.Pk

	_, err = apiClient.CoreApi.CoreGroupsDestroy(authCtx, groupUUID).Execute()

	return err
}
