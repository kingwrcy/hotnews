package model

import "gorm.io/gorm"

const TableNameTbTag = "tb_tag"

type TbTag struct {
	gorm.Model
	Name     string   `gorm:"column:name;type:varchar(50);unique"`
	Desc     string   `gorm:"column:desc;type:varchar(128)"`
	Posts    []TbPost `gorm:"many2many:tb_post_tag"`
	Parent   *TbTag
	ParentID *uint
	Children []TbTag `gorm:"foreignkey:ParentID"`
	BGColor  string  `gorm:"column:bg_color;type:varchar(30)"`
}

func (*TbTag) TableName() string {
	return TableNameTbTag
}
