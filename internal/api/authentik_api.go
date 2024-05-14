package api

import (
	"context"
	"fmt"
	"goauthentik.io/api/v3"
	"os"
)

func GetClient(ctx context.Context) (api.APIClient, context.Context) {
	configuration := api.NewConfiguration()
	configuration.Host = os.Getenv("AUTHENTIK_URL")
	configuration.Scheme = "https"
	apiClient := api.NewAPIClient(configuration)
	authCtx := context.WithValue(ctx, api.ContextAccessToken, os.Getenv("AUTHENTIK_TOKEN"))

	return *apiClient, authCtx
}

func GetUser(ctx context.Context, name string) (*api.User, error) {
	apiClient, authCtx := GetClient(ctx)

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

func CreateOrUpdateUser(ctx context.Context, user *api.User) (*api.User, error) {
	existingUser, err := GetUser(ctx, user.Name)

	apiClient, authCtx := GetClient(ctx)

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
	} else {
		userUUID := existingUser.Pk
		updateRequest := api.NewUserRequest(user.Username, user.Name)
		updateRequest.SetEmail(*user.Email)
		updateRequest.SetIsActive(true)

		existingUser, _, err = apiClient.CoreApi.CoreUsersUpdate(authCtx, userUUID).UserRequest(*updateRequest).Execute()

		if err != nil {
			return nil, err
		}
	}

	if existingUser == nil {
		return nil, fmt.Errorf("unable to create/fetch user")
	}

	extraGroups, newGroups := difference(existingUser.Groups, user.Groups)

	for _, group := range extraGroups {
		err = RemoveFromGroup(ctx, existingUser, group)

		if err != nil {
			return nil, err
		}
	}

	for _, group := range newGroups {
		err = AddToGroup(ctx, existingUser, group)

		if err != nil {
			return nil, err
		}
	}

	return existingUser, nil
}

func DeleteUser(ctx context.Context, user string) error {
	existingUser, err := GetUser(ctx, user)

	if err != nil {
		return err
	}

	if existingUser == nil {
		return nil
	}

	apiClient, authCtx := GetClient(ctx)

	userUUID := existingUser.Pk

	_, err = apiClient.CoreApi.CoreUsersDestroy(authCtx, userUUID).Execute()

	return err
}

func RemoveFromGroup(ctx context.Context, user *api.User, groupName string) error {
	group, err := GetGroup(ctx, groupName)

	if err != nil {
		return err
	}

	apiClient, authCtx := GetClient(ctx)

	userAccountRequest := api.NewUserAccountRequest(user.Pk)

	_, err = apiClient.CoreApi.CoreGroupsRemoveUserCreate(authCtx, group.Pk).UserAccountRequest(*userAccountRequest).Execute()

	return err
}

func AddToGroup(ctx context.Context, user *api.User, groupName string) error {
	group, err := GetGroup(ctx, groupName)

	if err != nil {
		return err
	}

	apiClient, authCtx := GetClient(ctx)

	userAccountRequest := api.NewUserAccountRequest(user.Pk)

	_, err = apiClient.CoreApi.CoreGroupsAddUserCreate(authCtx, group.Pk).UserAccountRequest(*userAccountRequest).Execute()

	return err
}

func GetGroup(ctx context.Context, name string) (*api.Group, error) {
	apiClient, authCtx := GetClient(ctx)

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

func CreateOrUpdateGroup(ctx context.Context, group *api.Group) (*api.Group, error) {
	existingGroup, err := GetGroup(ctx, group.Name)

	apiClient, authCtx := GetClient(ctx)

	if err != nil {
		return nil, err
	}

	if existingGroup == nil {
		createRequest := api.NewGroupRequest(group.Name)
		createRequest.SetIsSuperuser(*group.IsSuperuser)

		resp, _, err := apiClient.CoreApi.CoreGroupsCreate(authCtx).GroupRequest(*createRequest).Execute()

		if err != nil {
			return nil, err
		} else {
			return resp, nil
		}
	} else {
		groupUUID := existingGroup.Pk
		updateRequest := api.NewGroupRequest(group.Name)
		updateRequest.SetIsSuperuser(*group.IsSuperuser)

		resp, _, err := apiClient.CoreApi.CoreGroupsUpdate(authCtx, groupUUID).GroupRequest(*updateRequest).Execute()

		if err != nil {
			return nil, err
		} else {
			return resp, nil
		}
	}
}

func DeleteGroup(ctx context.Context, group string) error {
	existingGroup, err := GetGroup(ctx, group)

	if err != nil {
		return err
	}

	if existingGroup == nil {
		return nil
	}

	apiClient, authCtx := GetClient(ctx)

	groupUUID := existingGroup.Pk

	_, err = apiClient.CoreApi.CoreGroupsDestroy(authCtx, groupUUID).Execute()

	return err
}

func difference(slice1 []string, slice2 []string) ([]string, []string) {
	var diffleft []string
	var diffright []string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				if i == 0 {
					diffleft = append(diffleft, s1)
				} else {
					diffright = append(diffright, s1)
				}
			}
		}
	}

	return diffleft, diffright
}
