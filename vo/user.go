package vo

type LoginRequest struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}
type LoginResponse struct {
	Username string `json:"username,omitempty"`
	Token    string `json:"token,omitempty"`
	Role     string `json:"role,omitempty"`
	UID      string `json:"uid,omitempty"`
}
