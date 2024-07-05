package vo

type SaveSettingsRequest struct {
	RegMode string `form:"regMode" json:"regMode"`
	Css     string `form:"css" json:"css"`
	Js      string `form:"js" json:"js"`
}
