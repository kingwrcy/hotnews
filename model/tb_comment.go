package model

import "gorm.io/gorm"

const TableNameTbComment = "tb_comment"

type TbComment struct {
	gorm.Model
	UserID          uint `gorm:"column:user_id;type:int"`
	User            TbUser
	Content         string `gorm:"column:content;type:varchar"`
	CID             string `gorm:"column:cid;type:varchar(20)"`
	PostID          uint   `gorm:"column:post_id;type:int"`
	Post            TbPost
	ParentCommentID *uint `gorm:"column:parent_comment_id;type:int"`
	ParentComment   *TbComment
	UpVote          int         `gorm:"column:upVote;type:int"`
	DownVote        int         `gorm:"column:downVote;type:int"`
	Comments        []TbComment `gorm:"foreignKey:ParentCommentID"`
	UpVoted         int
}

func (*TbComment) TableName() string {
	return TableNameTbComment
}
