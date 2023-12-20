package model

import "gorm.io/gorm"

const TableNameTbVote = "tb_vote"

type TbVote struct {
	gorm.Model
	UserID   uint   `gorm:"column:tb_user_id;type:int"`
	TargetID uint   `gorm:"column:target_id;type:int"`
	Action   string `gorm:"column:action;type:varchar(20)"`
	Type     string `gorm:"column:type;type:varchar(20)"`
}

func (*TbVote) TableName() string {
	return TableNameTbVote
}
