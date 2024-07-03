package model

import (
	"encoding/json"
	"gorm.io/gorm"
)

const TableNameTbMessage = "tb_message"

type TbMessage struct {
	gorm.Model
	Content    string `gorm:"column:content;type:varchar;"`
	Read       string `gorm:"column:read;type:varchar(1)"`
	FromUserID uint   `gorm:"column:from_user_id;type:int"`
	FromUser   TbUser `gorm:"foreignKey:FromUserID"`
	ToUserID   uint   `gorm:"column:to_user_id;type:int"`
	ToUser     TbUser `gorm:"foreignKey:ToUserID"`
}

func (*TbMessage) TableName() string {
	return TableNameTbMessage
}

func (t *TbMessage) String() string {
	marshal, _ := json.Marshal(t)
	return string(marshal)
}
