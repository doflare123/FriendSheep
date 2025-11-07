package dto

type AuthResponse struct {
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token"`
	AdminGroups  []AdminGroupResponse `json:"admin_groups"`
}
