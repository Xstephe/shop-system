package forms

type SendSmsForm struct {
	Mobile string `form:"mobile" json:"mobile" binding:"required,mobile"`
	Type   uint   `form:"type" json:"type" binding:"required,oneof=1 2"` //1表示注册，2表示找回密码
}
