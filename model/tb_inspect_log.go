package model

import (
	"encoding/json"
	"gorm.io/gorm"
)

const TableNameTbInspectLog = "tb_inspect_log"

type TbInspectLog struct {
	gorm.Model
	PostID      uint `gorm:"column:post_id"`
	Post        TbPost
	InspectType string `gorm:"column:inspect_type;type:varchar(30)"`
	Reason      string `gorm:"column:reason;type:varchar(256)"`
	Result      string `gorm:"column:result;type:varchar(50)"`
	Action      string `gorm:"column:action;type:varchar(120)"`
	InspectorID uint   `gorm:"column:inspector_id"`
	Inspector   TbUser
	Title       string `gorm:"column:title;type:varchar(256)"`
}

func (*TbInspectLog) TableName() string {
	return TableNameTbInspectLog
}

func (t *TbInspectLog) String() string {
	marshal, _ := json.Marshal(t)
	return string(marshal)
}
