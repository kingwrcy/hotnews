package model

import "gorm.io/gorm"

const TableNameTbVote = "tb_vote"

type TbVote struct {
	gorm.Model
	UserID    uint `gorm:"column:user_id;type:int"`
	User      TbUser
	PostID    *uint
	Post      TbPost
	CommentID *uint
	Comment   TbComment
	Action    string `gorm:"column:action;type:varchar(20)"`
}

func (*TbComment) TbVote() string {
	return TableNameTbVote
}
