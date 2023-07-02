package model

import "gorm.io/gorm"

const TableNameTbUser = "tb_user"

type TbUser struct {
	gorm.Model
	Username     string `gorm:"column:id;type:varchar(30)"`
	Password     string `gorm:"column:password;type:varchar(100)"`
	Role         string `gorm:"column:role;type:varchar(20)"`
	Email        string `gorm:"column:email;type:varchar(100)"`
	Bio          string `gorm:"column:bio;type:varchar(100)"`
	CommentCount int    `gorm:"column:commentCount;type:int"`
	PostCount    int    `gorm:"column:postCount;type:int"`
}

func (*TbUser) TableName() string {
	return TableNameTbUser
}
