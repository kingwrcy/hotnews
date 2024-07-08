package model

import (
	"encoding/json"
	"gorm.io/gorm"
	"html/template"
)

const TableNameTbPost = "tb_post"

type TbPost struct {
	gorm.Model
	Title           string        `gorm:"column:title;type:varchar(100);"`
	Link            string        `gorm:"column:link;type:varchar(256)"`
	Status          string        `gorm:"column:status;type:varchar(20)"`
	Content         string        `gorm:"column:content;type:text"`
	UnEscapeContent template.HTML `gorm:"-:all"`
	UpVote          int           `gorm:"column:upVote;type:int"`
	DownVote        int           `gorm:"column:downVote;type:int"`
	Type            string        `gorm:"column:type;type:varchar(20)"`
	User            TbUser        `gorm:"foreignKey:UserID"`
	UserID          uint
	Tags            []TbTag     `gorm:"many2many:tb_post_tag"`
	Remark          string      `gorm:"column:remark;type:varchar(256)"`
	Domain          string      `gorm:"column:domain;type:varchar(256)"`
	Pid             string      `gorm:"column:pid;type:varchar(20);unique"`
	CommentCount    int         `gorm:"column:commentCount;type:int"`
	Comments        []TbComment `gorm:"foreignKey:PostID"`
	Point           float64     `gorm:"column:point;type:decimal(20,10)"`
	UpVoted         int         `gorm:"<-"`
	DownVoted       int         `gorm:"<-"`
	Top             int         `gorm:"column:top;type:int;default:0"`
}

func (*TbPost) TableName() string {
	return TableNameTbPost
}

func (t *TbPost) String() string {
	marshal, _ := json.Marshal(t)
	return string(marshal)
}
