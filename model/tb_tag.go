package model

import "gorm.io/gorm"

const TableNameTbTag = "tb_tag"

type TbTag struct {
	gorm.Model
	Name      string   `gorm:"column:name;type:varchar(50);unique"`
	Desc      string   `gorm:"column:desc;type:varchar(128)"`
	Posts     []TbPost `gorm:"many2many:tb_post_tag"`
	Parent    *TbTag
	ParentID  *uint
	Children  []TbTag `gorm:"foreignkey:ParentID"`
	CssClass  string  `gorm:"column:css_class;type:varchar(120)"`
	ShowInHot string  `gorm:"column:show_in_hot;type:varchar(5)"`
	ShowInAll string  `gorm:"column:show_in_all;type:varchar(5)"`
}

func (*TbTag) TableName() string {
	return TableNameTbTag
}
