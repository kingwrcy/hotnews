package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/kingwrcy/hn/vo"
)

type SaveSettingsRequest vo.SaveSettingsRequest

const TableNameTbSettings = "tb_settings"

type TbSettings struct {
	ID      uint                `gorm:"primarykey"`
	Content SaveSettingsRequest `gorm:"type:jsonb"`
}

func (*TbSettings) TableName() string {
	return TableNameTbSettings
}

// Scan 实现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (j *SaveSettingsRequest) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}

	var buf SaveSettingsRequest
	err := json.Unmarshal(bytes, &buf)
	*j = buf
	return err
}

// Value 实现 driver.Valuer 接口，Value 返回 json value
func (j SaveSettingsRequest) Value() (driver.Value, error) {
	return json.Marshal(&j)
}
