package vo

import (
	"encoding/json"
	"time"
)

type Memo struct {
	Content     string     `json:"content,omitempty"`
	PublishTime time.Time  `json:"publishTime" binding:"required"`
	Created     time.Time  `json:"created"`
	Updated     time.Time  `json:"updated"`
	Author      string     `json:"author,omitempty"`
	Website     string     `json:"website,omitempty" binding:"required"`
	MemoID      int32      `json:"memoId,omitempty" binding:"required"`
	AvatarURL   string     `json:"avatarUrl,omitempty"`
	Tags        string     `json:"tags,omitempty"`
	Email       string     `json:"email,omitempty"`
	UserID      int32      `json:"userId,omitempty" binding:"required"`
	Resource    []Resource `json:"resources,omitempty"`
}

func (m Memo) String() string {
	result, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	return string(result)
}

type Resource struct {
	Url         string `json:"url,omitempty"`
	Suffix      string `json:"suffix,omitempty"`
	PublicId    string `json:"publicId,omitempty"`
	FileType    string `json:"fileType,omitempty"`
	StorageType string `json:"storageType,omitempty"`
	FileName    string `json:"fileName,omitempty"`
}

type PageListMemoRequest struct {
	Page int `json:"page,omitempty"`
	Size int `json:"size,omitempty"`
}

type PageListMemoResponse struct {
	TotalPage int    `json:"totalPage,omitempty"`
	TotalRows int64  `json:"totalRows,omitempty"`
	List      []Memo `json:"list,omitempty"`
}
