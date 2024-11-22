package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"mxshop-api/user-web/forms"
	"mxshop-api/user-web/global"
	"mxshop-api/user-web/global/response"
	"mxshop-api/user-web/middlewares"
	"mxshop-api/user-web/models"
	"mxshop-api/user-web/proto"
)

// 去掉错误前缀
func removeTopStruct(fields map[string]string) map[string]string {
	rsp := map[string]string{}
	for field, err := range fields {
		rsp[field[strings.Index(field, ".")+1:]] = err
	}
	return rsp
}

// 将grpc的code转化成http的状态码
func HandleGrpcErrorToHttp(err error, ctx *gin.Context) {
	if err != nil {
		if s, ok := status.FromError(err); ok {
			switch s.Code() {
			case codes.NotFound:
				ctx.JSON(http.StatusNotFound, gin.H{
					"msg": s.Message(),
				})
			case codes.Internal:
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"msg": "内部错误",
				})
			case codes.InvalidArgument:
				ctx.JSON(http.StatusBadRequest, gin.H{
					"msg": "参数错误",
				})
			case codes.Unavailable:
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"msg": "用户服务不可用",
				})
			default:
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"msg": "其他错误",
				})
			}
			return
		}
	}
}

// validator错误中文翻译器
func HandlerValidatorError(err error, ctx *gin.Context) {
	if err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ctx.JSON(http.StatusOK, gin.H{
				"msg": err.Error(),
			})
			return
		}
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": removeTopStruct(errs.Translate(global.Trans)),
		})
		return
	}
}

// 获取用户列表
func GetUserList(ctx *gin.Context) {
	//拿到登录用户的id
	claims, _ := ctx.Get("claims")
	currentUser := claims.(*models.CustomClaims)
	zap.S().Infof("现在登录用户的ID为:%d", currentUser.ID)

	pn := ctx.DefaultQuery("pn", "0")
	pnInt, _ := strconv.Atoi(pn)
	page := ctx.DefaultQuery("pnsize", "0")
	pageInt, _ := strconv.Atoi(page)
	userListRsp, err := global.UserSrvClient.GetUserList(context.Background(), &proto.PageInfo{
		Pn:    uint32(pnInt),
		PSize: uint32(pageInt),
	})
	if err != nil {
		zap.S().Errorw("[GetUserList] 查询 [用户列表] 失败")
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	result := make([]interface{}, 0)
	for _, value := range userListRsp.Data {

		//采用结构体的方式返回  这种更加规范
		data := response.UserResponse{
			Id:       value.Id,
			Password: value.Password,
			Mobile:   value.Mobile,
			NickName: value.NickName,
			BirthDay: response.JsonTime(time.Unix(int64(value.BirthDay), 0)),
			Gender:   value.Gender,
			Role:     value.Role,
		}
		result = append(result, data)

		/*
			采用map的形式进行返回
			data := make(map[string]interface{})
			data["id"] = value.Id
			data["nickName"] = value.NickName
			data["gender"] = value.Gender
			data["mobile"] = value.Mobile
			data["role"] = value.Role
			data["birthday"] = value.BirthDay
			data["password"] = value.Password
			result = append(result, data)
		*/
	}
	ctx.JSON(http.StatusOK, result)
}

// 密码登陆
func PassWordLogin(ctx *gin.Context) {
	//表单验证
	passwordLoginFrom := forms.PassWordLoginForm{}
	err := ctx.ShouldBind(&passwordLoginFrom)
	if err != nil {
		HandlerValidatorError(err, ctx)
		return
	}

	//图片验证
	if !store.Verify(passwordLoginFrom.CaptchaId, passwordLoginFrom.Captcha, true) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"captcha": "验证码错误",
		})
		return
	}

	UserInfoRsp, err := global.UserSrvClient.GetUserByMobile(context.Background(), &proto.MobileRequest{
		Mobile: passwordLoginFrom.Mobile,
	})
	if err != nil {
		HandleGrpcErrorToHttp(err, ctx)
		return
	} else {
		//只是查询到了用户而已，并没有检查密码，要进行检查密码
		passRsp, pasErr := global.UserSrvClient.CheckPassword(context.Background(), &proto.PasswordCheckInfo{
			Password:          passwordLoginFrom.Password,
			EncryptedPassword: UserInfoRsp.Password,
		})
		if pasErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"msg": "登录失败",
			})
		} else {
			if !passRsp.Success {
				ctx.JSON(http.StatusBadRequest, gin.H{
					"msg": "密码错误",
				})
			} else {
				//生成token,用来做认证
				j := middlewares.NewJWT()
				claim := models.CustomClaims{
					ID:          uint(UserInfoRsp.Id),
					NickName:    UserInfoRsp.NickName,
					AuthorityId: uint(UserInfoRsp.Role),
					StandardClaims: jwt.StandardClaims{
						NotBefore: time.Now().Unix(),               //签名的生效时间
						Issuer:    "imooc",                         //认证机构
						ExpiresAt: time.Now().Unix() + 60*60*24*30, //签名的过期时间
					},
				}
				token, err := j.CreateToken(claim)
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{
						"msg": "生成token失败",
					})
					return
				}
				ctx.JSON(http.StatusOK, gin.H{
					"id":         UserInfoRsp.Id,
					"nickName":   UserInfoRsp.NickName,
					"token":      token,
					"expired_at": (time.Now().Unix() + 60*60*24*30) * 1000,
					"msg":        "登录成功",
				})
			}
		}
	}
}

// 新建用户
func Register(ctx *gin.Context) {
	//表单验证
	registerFrom := forms.RegisterForm{}
	err := ctx.ShouldBind(&registerFrom)
	if err != nil {
		HandlerValidatorError(err, ctx)
		return
	}

	//验证码校验，从redis中进行拉取
	rdb := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", global.ServerConfig.RedisInfo.Host, global.ServerConfig.RedisInfo.Port),
	})
	value, err := rdb.Get(context.Background(), registerFrom.Mobile).Result()
	if err == redis.Nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "注册的手机号未发送短信，该手机号不存在", //redis中的手机号没有存储对应的value
		})
		return
	} else {
		if value != registerFrom.Code {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"code": "验证码错误", //传入的code错误
			})
			return
		}
	}

	UserInfoRsp, err := global.UserSrvClient.CreateUser(context.Background(), &proto.CreateUserInfo{
		NickName: registerFrom.Mobile,
		Password: registerFrom.Password,
		Mobile:   registerFrom.Mobile,
	})

	if err != nil {
		zap.S().Errorw("[Register] 新建 [用户] 失败")
		HandleGrpcErrorToHttp(err, ctx)
		return
	}
	j := middlewares.NewJWT()
	claim := models.CustomClaims{
		ID:          uint(UserInfoRsp.Id),
		NickName:    UserInfoRsp.NickName,
		AuthorityId: uint(UserInfoRsp.Role),
		StandardClaims: jwt.StandardClaims{
			NotBefore: time.Now().Unix(),               //签名的生效时间
			Issuer:    "imooc",                         //认证机构
			ExpiresAt: time.Now().Unix() + 60*60*24*30, //签名的过期时间
		},
	}
	token, err := j.CreateToken(claim)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "生成token失败",
		})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{
		"id":         UserInfoRsp.Id,
		"nickName":   UserInfoRsp.NickName,
		"token":      token,
		"expired_at": (time.Now().Unix() + 60*60*24*30) * 1000,
		"msg":        "注册成功",
	})

}

//修改用户

//通过id获取用户

//通过手机号查询用户
