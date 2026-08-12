package dto

type AuthResponse struct {
	AccessToken  string               `json:"access_token"`
	RefreshToken string               `json:"refresh_token"`
	TokenType    string               `json:"token_type" example:"Bearer"`
	ExpiresIn    int64                `json:"expires_in" example:"1200"`
	Me           AuthMeResponse       `json:"me"`
	AdminGroups  []AdminGroupResponse `json:"admin_groups"`
}

type RegistrationAuthResponse struct {
	Message string `json:"message"`
	AuthResponse
}

type AuthMeResponse struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Us    string `json:"us"`
	Image string `json:"image"`
}
