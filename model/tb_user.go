package model

import "gorm.io/gorm"

const TableNameTbUser = "tb_user"

type TbUser struct {
	gorm.Model
	Username        string      `gorm:"column:username;type:varchar(30);unique"`
	Password        string      `gorm:"column:password;type:varchar(100)"`
	Role            string      `gorm:"column:role;type:varchar(20)"`
	Email           string      `gorm:"column:email;type:varchar(100)"`
	Bio             string      `gorm:"column:bio;type:varchar(100)"`
	CommentCount    int         `gorm:"column:commentCount;type:int"`
	PostCount       int         `gorm:"column:postCount;type:int"`
	Status          string      `gorm:"column:status;type:varchar(20)"`
	EmailHash       string      `gorm:"column:email_hash;type:varchar(80)"`
	Posts           []TbPost    `gorm:"foreignKey:UserID"`
	UpVotedPosts    []TbPost    `gorm:"many2many:tb_vote;"`
	Comments        []TbComment `gorm:"foreignKey:UserID"`
	UpVotedComments []TbComment `gorm:"many2many:tb_vote;"`
}

func (*TbUser) TableName() string {
	return TableNameTbUser
}
