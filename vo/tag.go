package vo

type EditTagVo struct {
	Name      string `form:"name,omitempty"`
	Desc      string `form:"desc,omitempty"`
	ID        uint   `form:"id,omitempty"`
	ParentID  *uint  `form:"parentID,omitempty"`
	ShowInAll string `form:"showInAll,omitempty"`
	ShowInHot string `form:"showInHot,omitempty"`
	CssClass  string `form:"cssClass,omitempty"`
}
