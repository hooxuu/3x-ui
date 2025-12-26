package controller

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mhsanaei/3x-ui/v2/database/model"
	"github.com/mhsanaei/3x-ui/v2/web/service"
	"github.com/mhsanaei/3x-ui/v2/web/session"
)

type UserController struct {
	userService service.UserService
}

func NewUserController(g *gin.RouterGroup) *UserController {
	a := &UserController{}
	a.initRouter(g)
	return a
}

func (a *UserController) initRouter(g *gin.RouterGroup) {
	g.GET("/list", a.getUsers)
	g.POST("/add", a.addUser)
	g.POST("/update/:id", a.updateUser)
	g.POST("/del/:id", a.delUser)
}

func (a *UserController) checkAdmin(c *gin.Context) bool {
	user := session.GetLoginUser(c)
	if user != nil && (user.Role == model.UserRoleAdmin || user.Role == "") {
		return true
	}
	jsonMsg(c, "Permission Denied", errors.New("admin role required"))
	return false
}

func (a *UserController) getUsers(c *gin.Context) {
	if !a.checkAdmin(c) {
		return
	}
	users, err := a.userService.GetAllUsers()
	if err != nil {
		jsonMsg(c, I18nWeb(c, "get"), err)
		return
	}
	jsonObj(c, users, nil)
}

func (a *UserController) addUser(c *gin.Context) {
	if !a.checkAdmin(c) {
		return
	}
	user := &model.User{}
	err := c.ShouldBind(user)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "create"), err)
		return
	}

	if user.Role == "" {
		user.Role = model.UserRoleTenant
	}

	err = a.userService.AddUser(user)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "create"), err)
		return
	}
	jsonMsg(c, I18nWeb(c, "success"), nil)
}

func (a *UserController) updateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "update"), err)
		return
	}

	loginUser := session.GetLoginUser(c)
	if loginUser == nil {
		jsonMsg(c, I18nWeb(c, "update"), errors.New("login required"))
		return
	}

	// Allow admin or self-update
	if loginUser.Role != model.UserRoleAdmin && loginUser.Role != "" && loginUser.Id != id {
		jsonMsg(c, "Permission Denied", errors.New("admin role required"))
		return
	}

	updates := make(map[string]interface{})
	contentType := c.GetHeader("Content-Type")

	if strings.Contains(contentType, "application/json") {
		if err := c.ShouldBindJSON(&updates); err != nil {
			jsonMsg(c, I18nWeb(c, "update"), err)
			return
		}
	} else {
		if err := c.Request.ParseForm(); err != nil {
			jsonMsg(c, I18nWeb(c, "update"), err)
			return
		}
		for _, field := range []string{"username", "password", "role", "remark"} {
			if values, ok := c.Request.PostForm[field]; ok && len(values) > 0 {
				updates[field] = values[0]
			}
		}
	}

	err = a.userService.UpdateUserById(id, updates)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "update"), err)
		return
	}
	jsonMsg(c, I18nWeb(c, "success"), nil)
}

func (a *UserController) delUser(c *gin.Context) {
	if !a.checkAdmin(c) {
		return
	}
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		jsonMsg(c, I18nWeb(c, "delete"), err)
		return
	}
	err = a.userService.DeleteUserById(id)
	if err != nil {
		jsonMsg(c, I18nWeb(c, "delete"), err)
		return
	}
	jsonMsg(c, I18nWeb(c, "success"), nil)
}
