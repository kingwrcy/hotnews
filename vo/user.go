package vo

type LoginRequest struct {
	Username string `form:"username,omitempty"`
	Password string `form:"password,omitempty"`
}
type RegisterRequest struct {
	Username       string `form:"username,omitempty"`
	Password       string `form:"password,omitempty"`
	RepeatPassword string `form:"repeatPassword,omitempty"`
	Email          string `form:"email,omitempty"`
	Bio            string `form:"bio,omitempty"`
}

type Userinfo struct {
	Username  string
	Role      string
	ID        uint
	Email     string
	EmailHash string
}
