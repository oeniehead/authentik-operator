package api

import (
	"fmt"
	"goauthentik.io/api/v3"
)

func GetUser(cl *AuthentikApiClient, name string) (*api.User, error) {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	resp, _, err := apiClient.CoreApi.CoreUsersList(authCtx).Name(name).Execute()

	if err != nil {
		return nil, err
	}

	if len(resp.Results) == 0 {
		return nil, nil
	} else {
		return &resp.Results[0], nil
	}
}

func CreateUser(cl *AuthentikApiClient, user *api.User) (*api.User, error) {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	existingUser, err := GetUser(cl, user.Name)

	if err != nil {
		return nil, err
	}

	if existingUser == nil {
		createRequest := api.NewUserRequest(user.Username, user.Name)
		createRequest.SetEmail(*user.Email)
		createRequest.SetIsActive(true)

		existingUser, _, err = apiClient.CoreApi.CoreUsersCreate(authCtx).UserRequest(*createRequest).Execute()

		if err != nil {
			return nil, err
		}
	}

	if existingUser == nil {
		return nil, fmt.Errorf("unable to create/fetch user")
	}

	for _, group := range user.Groups {
		err = AddToGroup(cl, existingUser, group)

		if err != nil {
			return nil, err
		}
	}

	return existingUser, nil
}

func SynchronizeGroups(cl *AuthentikApiClient, existingUser *api.User, targetGroups []string) error {
	var existingGroups []string
	for _, groupId := range existingUser.Groups {
		group, err := GetGroupById(cl, groupId)

		if err != nil {
			return err
		}

		if group == nil {
			return fmt.Errorf("group %s not found", group)
		}

		existingGroups = append(existingGroups, group.Name)
	}

	extraGroups, newGroups := difference(existingGroups, targetGroups)

	for _, group := range extraGroups {
		err := RemoveFromGroup(cl, existingUser, group)

		if err != nil {
			return err
		}
	}

	for _, group := range newGroups {
		err := AddToGroup(cl, existingUser, group)

		if err != nil {
			return err
		}
	}

	return nil
}

func DeleteUser(cl *AuthentikApiClient, user string) error {
	apiClient := cl.apiClient
	authCtx := cl.ctx

	existingUser, err := GetUser(cl, user)

	if err != nil {
		return err
	}

	if existingUser == nil {
		return nil
	}

	userUUID := existingUser.Pk

	_, err = apiClient.CoreApi.CoreUsersDestroy(authCtx, userUUID).Execute()

	return err
}
