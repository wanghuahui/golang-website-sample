package main

import (
	"net/http"
	"strings"

	"golang-website-sample/webserver/model"

	"github.com/labstack/echo"
)

// ルーティングに対応するハンドラを設定します。
func setRoute(e *echo.Echo) {
	e.GET("/", handleIndexGet)
	e.GET("/register", handleRegisterGet)
	e.POST("/register", handleRegisterPost)
	e.GET("/login", handleLoginGet)
	e.POST("/login", handleLoginPost)
	e.POST("/logout", handleLogoutPost)
	e.GET("/users/:user_id", handleUsers)
	e.POST("/users/:user_id", handleUsers)

	// 管理者のみが参照できるページ
	admin := e.Group("/admin", MiddlewareAuthAdmin)
	admin.GET("", handleAdmin)
	admin.POST("", handleAdmin)
	admin.GET("/users", handleAdminUsersGet)
}

// GET:/
func handleIndexGet(c echo.Context) error {
	return c.Render(http.StatusOK, "index", "world")
}

// GET:/users/:user_id
// POST:/users/:user_id
func handleUsers(c echo.Context) error {
	userID := c.Param("user_id")
	err := CheckUserID(c, userID)
	if err != nil {
		c.Echo().Logger.Debugf("User Page[%s] Role Error. [%s]", userID, err)
		msg := "ログインしていません。"
		return c.Render(http.StatusOK, "error", msg)
	}
	users, err := userDA.FindByUserID(c.Param("user_id"), model.FindFirst)
	if err != nil {
		return c.Render(http.StatusOK, "error", err)
	}
	user := users[0]
	return c.Render(http.StatusOK, "user", user)
}

// GET:/admin
// POST:/admin
func handleAdmin(c echo.Context) error {
	return c.Render(http.StatusOK, "admin", nil)
}

// GET:/admin/users
func handleAdminUsersGet(c echo.Context) error {
	users, err := userDA.FindAll()
	if err != nil {
		return c.Render(http.StatusOK, "error", err)
	}
	return c.Render(http.StatusOK, "admin_users", users)
}

// GET:/register
func handleRegisterGet(c echo.Context) error {
	return c.Render(http.StatusOK, "register", nil)
}

// POST:/register
func handleRegisterPost(c echo.Context) error {
	userID := c.FormValue("userid")
	user := &model.User{}
	if ok := user.UserIDIsExist(userID); ok {
		c.Echo().Logger.Debugf("user id is exist")
		idTip := "用户ID已存在"
		data := map[string]string{"user_id": "", "password": "", "id_tip": idTip}
		return c.Render(http.StatusOK, "register", data)
	}
	password := c.FormValue("password")
	passwordVer := c.FormValue("password_verify")
	if strings.Compare(password, passwordVer) != 0 {
		msg := "两次输入密码不一致"
		data := map[string]string{"user_id": userID, "password": "", "msg": msg}
		return c.Render(http.StatusOK, "register", data)
	}
	name := c.FormValue("fullname")

	user = &model.User{
		UserID:   userID,
		Password: model.StringMD5(password),
		FullName: name,
		Roles:    []model.Role{"user"},
	}
	err := UserRegister(c, user)
	if err != nil {
		user.UserDelete(userID)
		c.Echo().Logger.Debugf("create user info error", err)
		msg := "创建用户信息错误"
		data := map[string]string{"user_id": userID, "password": "", "msg": msg}
		return c.Render(http.StatusOK, "register", data)
	}
	return c.Redirect(http.StatusFound, "/users/"+userID)
}

// GET:/login
func handleLoginGet(c echo.Context) error {
	return c.Render(http.StatusOK, "login", nil)
}

// POST:/login
func handleLoginPost(c echo.Context) error {
	userID := c.FormValue("userid")
	password := c.FormValue("password")
	err := UserLogin(c, userID, password)
	if err != nil {
		c.Echo().Logger.Debugf("User[%s] Login Error. [%s]", userID, err)
		msg := "用户id或密码错误。"
		data := map[string]string{"user_id": userID, "password": "", "msg": msg}
		return c.Render(http.StatusOK, "login", data)
	}
	// 检查用户是否为管理员
	isAdmin, err := CheckRoleByUserID(userID, model.RoleAdmin)
	if err != nil {
		c.Echo().Logger.Debugf("Admin Role Check Error. [%s]", userID, err)
		isAdmin = false
	}
	if isAdmin {
		// 如果是管理者身份，则跳转至管理者页面
		c.Echo().Logger.Debugf("User is Admin. [%s]", userID)
		return c.Redirect(http.StatusTemporaryRedirect, "/admin")
	}
	return c.Redirect(http.StatusTemporaryRedirect, "/users/"+userID)
}

// POST:/logout
func handleLogoutPost(c echo.Context) error {
	err := UserLogout(c)
	if err != nil {
		c.Echo().Logger.Debugf("User Logout Error. [%s]", err)
		return c.Render(http.StatusOK, "login", nil)
	}
	msg := "退出登录。"
	data := map[string]string{"user_id": "", "password": "", "msg": msg}
	return c.Render(http.StatusOK, "login", data)
}
