package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"net/http"
)

/*
在并发下,User应该使用actor模型来处理各种渠道发送的请求
建议使用channel
*/
type User struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Version int    `json:"version"`
	Enable  bool   `json:"enable"`
}

func (u *User) GetAggregateName() string {
	return "User"
}

func (u *User) GetVersion() int {
	return u.Version
}

func (u *User) GetId() string {
	return u.Id
}

func (u *User) Apply(c Event) error {
	//sendEvent
	err := SendEvent(c, u)
	if err != nil {
		return err
	}
	return nil
}

func LoadUser(id string) *User {
	u := &User{}
	u.Id = id
	u.Version = -1

	projection, err := LoadProjection(u)
	if err != nil {
		panic(err.Error())
	}
	s, ok := projection.(string)
	if !ok {
		panic(fmt.Errorf("projection load error,result:%+v", projection))
	}
	if len(s) <= 0 {
		return u
	}

	err = json.Unmarshal([]byte(s), u)
	if err != nil {
		panic(err.Error())
	}
	return u
}

type CreateUser struct {
	Name string `json:"name"`
	Id   string `json:"id"`
}

func (c *CreateUser) GetType() string {
	return "CreateUser"
}

type EnableUser struct {
	Id string `json:"id"`
}

func (e *EnableUser) GetType() string {
	return "EnableUser"
}

type DisableUser struct {
	Id string `json:"id"`
}

func (d *DisableUser) GetType() string {
	return "DisableUser"
}

func main() {
	engine := gin.Default()
	//User
	engine.GET("/user/:Id/profile", func(ctx *gin.Context) {
		param := ctx.Param("Id")
		user := LoadUser(param)
		ctx.JSON(http.StatusOK, user)
	})
	//createUser
	engine.POST("/createUser", func(ctx *gin.Context) {
		c := new(CreateUser)
		err := ctx.BindJSON(c)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		v4, err := uuid.NewV4()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		c.Id = v4.String()
		user := LoadUser(c.Id)
		if err = user.Apply(c); err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, c)
	})

	//enableUser
	engine.POST("/enableUser", func(ctx *gin.Context) {
		c := new(EnableUser)
		err := ctx.BindJSON(c)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		if err := LoadUser(c.Id).Apply(c); err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, "success")
	})

	//disableUser
	engine.POST("/disableUser", func(ctx *gin.Context) {
		c := new(DisableUser)
		err := ctx.BindJSON(c)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		if err := LoadUser(c.Id).Apply(c); err != nil {
			ctx.JSON(http.StatusInternalServerError, err.Error())
			return
		}
		ctx.JSON(http.StatusOK, "success")
	})

	engine.Run(":8090")
}
