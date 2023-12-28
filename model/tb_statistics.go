package model

import "time"

const TableNameTbStatistics = "tb_statistics"

type TbStatistics struct {
	IPHash    string `gorm:"column:ip_hash;primarykey"`
	IP        string `gorm:"column:ip;type:varchar(128)"`
	Refer     string `gorm:"column:refer;type:varchar(256)"`
	Target    string `gorm:"column:target;type:varchar(256)"`
	Country   string `gorm:"column:country;type:varchar(64)"`
	Device    string `gorm:"column:device;type:varchar(64)"`
	Mobile    bool   `gorm:"column:mobile;type:int"`
	Tablet    bool   `gorm:"column:tablet;type:int"`
	Desktop   bool   `gorm:"column:desktop;type:int"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (*TbStatistics) TableName() string {
	return TableNameTbStatistics
}
