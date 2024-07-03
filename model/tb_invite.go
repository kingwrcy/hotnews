package model

import (
	"gorm.io/gorm"
	"time"
)

const TableNameTbInviteRecord = "tb_invite_record"

type TbInviteRecord struct {
	gorm.Model
	Username         string    `gorm:"column:username;type:varchar(30);"`
	Code             string    `gorm:"column:code;type:varchar(30);unique"`
	InvitedUsername  string    `gorm:"column:invitedUsername;type:varchar(30)"`
	InvitedUserEmail string    `gorm:"column:invitedUserEmail;type:varchar(100)"`
	InvalidAt        time.Time `gorm:"column:invalidAt;type:timestamp"`
	Status           string    `gorm:"column:status;type:varchar(20)"`
}

func (*TbInviteRecord) TableName() string {
	return TableNameTbInviteRecord
}
