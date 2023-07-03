package model

import "gorm.io/gorm"

const TableNameTbPost = "tb_post"

type TbPost struct {
	gorm.Model
	Title    string `gorm:"column:title;type:varchar(100);"`
	Link     string `gorm:"column:link;type:varchar(256)"`
	Status   string `gorm:"column:status;type:varchar(20)"`
	Content  string `gorm:"column:email;type:text"`
	UpVote   int    `gorm:"column:upVote;type:int"`
	DownVote int    `gorm:"column:downVote;type:int"`
	Type     string `gorm:"column:type;type:varchar(20)"`
	User     TbUser
	UserID   uint
	Tags     []TbTag `gorm:"many2many:tb_post_tag"`
	Remark   string  `gorm:"column:remark;type:varchar(256)"`
}

func (*TbPost) TableName() string {
	return TableNameTbPost
}
