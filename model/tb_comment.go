package model

import "gorm.io/gorm"

const TableNameTbComment = "tb_comment"

type TbComment struct {
	gorm.Model
	UserID          uint `gorm:"column:user_id;type:int"`
	User            TbUser
	Content         string `gorm:"column:content;type:longtext"`
	PostID          uint   `gorm:"column:post_id;type:int"`
	Post            TbPost
	ParentCommentID *uint `gorm:"column:parent_comment_id;type:int"`
	ParentComment   *TbComment
}

func (*TbComment) TableName() string {
	return TableNameTbComment
}
