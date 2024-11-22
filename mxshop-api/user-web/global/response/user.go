package response

import (
	"fmt"
	"time"
)

type JsonTime time.Time

func (j JsonTime) MarshalJSON() ([]byte, error) {
	var stmp = fmt.Sprintf("\"%s\"", time.Time(j).Format("2006-01-02"))
	return []byte(stmp), nil
}

type UserResponse struct {
	Id       int32    `json:"id"`
	Password string   `json:"password"`
	Mobile   string   `json:"mobile"`
	NickName string   `json:"nickName"`
	BirthDay JsonTime `json:"birthDay"`
	Gender   string   `json:"gender"`
	Role     int32    `json:"role"`
}
