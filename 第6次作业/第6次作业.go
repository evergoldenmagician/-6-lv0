

package main

import (
"fmt"
"github.com/gin-gonic/gin"
"log"
"net/http"
)

//账号和用户信息的映射关系
var account map[string]User = make(map[string]User)

func main() {
	router := gin.Default()

	//gin的Default方法创建一个路由handler
	//通过HTTP方法绑定路由规则和路由函数
	//request和response都封装到gin.Context的上下文环境。最后是启动路由的Run方法监听端口
	//curl http://localhost:8080/hello
	//router.GET("/hello", func(c *gin.Context) {
	//	c.String(http.StatusOK, "Hello World")
	//})

	// 冒号:加上一个参数名组成路由参数。
	// 可以使用c.Params的方法读取其值。
	// 这个值是字串string。/user/hello都可以匹配，而/user/和/user/hello/不会被匹配
	// curl http://localhost:8080/user/hello
	router.GET("/user/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.String(http.StatusOK, "Hello %s", name)
	})

	//除了路由参数，其他的参数两种，
	// 查询字符串query string和报文体body参数。
	// 所谓query string，即路由用?以后连接的key1=value2&key2=value2的形式的参数
	router.GET("/query/hello", func(c *gin.Context) {
		name := c.DefaultQuery("name", "guest")
		c.JSON(200, gin.H{
			"message": "success",
			"name" : name,
		})
	})
	// http的报文体传Body，
	// 四种：application/json （前后端交互restful），
	// application/x-www-form-urlencoded （把query string的内容，放到了body体里，同样也需要urlencode）,
	// application/xml	和multipart/form-data (常用于上传文件)。
	// 默认情况下，c.PostFROM解析的是x-www-form-urlencoded或from-data的参数
	//curl -X POST http://127.0.0.1:8080/form_post -H "Content-Type:application/x-www-form-urlencoded" -d "message=hello&nickname=crown"
	router.POST("/form_post", func(c *gin.Context) {
		message := c.PostForm("message")
		nick := c.DefaultPostForm("nickname", "guest")

		c.JSON(http.StatusOK, gin.H{
			"status":  gin.H{
				"status_code": http.StatusOK,
				"status":      "ok",
			},
			"message": message,
			"nickname":    nick,
		})
	})

	//参数绑定
	//使用JSON来通信，响应和请求的content-type都是application/json的格式。
	//同时兼容一些还是application/x-www-form-urlencoded的web表单页
	// curl -X POST http://127.0.0.1:8080/login -H "Content-Type:application/json" -d '{"username": "crown","password":"321", "age": 21}' | python -m json.tool
	//	curl -X POST http://127.0.0.1:8080/login -H "Content-Type:application/x-www-form-urlencoded" -d "username=crown&password=123&age=12" | python -m json.tool
	//router.POST("/login", func(c *gin.Context) {
	//	var user User
	//	var err error
	//	contentType := c.Request.Header.Get("Content-Type")
	//
	//	switch contentType {
	//	case "application/json":
	//		err = c.BindJSON(&user)
	//	case "application/x-www-form-urlencoded":
	//		err = c.BindWith(&user, binding.Form)
	//	}
	//
	//	if err != nil {
	//		fmt.Println(err)
	//		log.Fatal(err)
	//	}
	//
	//	c.JSON(http.StatusOK, gin.H{
	//		"user":   user.Username,
	//		"password": user.Password,
	//		"age":    user.Age,
	//	})
	//})

	//c.Bind()	: 自动推断content-type是x-www-form-urlencoded表单还是json的参数。
	router.POST("/login", func(c *gin.Context) {
		var user User

		err := c.Bind(&user)
		if err != nil {
			fmt.Println(err)
			log.Fatal(err)
		}

		if v, ok := account[user.Username]; ok && v.Password == user.Password {
			c.JSON(http.StatusOK, gin.H{
				"username":   v.Username,
				"password":     v.Password,
				"age":        v.Age,
			})
		} else {
			c.JSON(http.StatusOK, gin.H{
				"message" : "账号或者密码有误",
			})
		}
	})

	router.POST("/register", func(c *gin.Context) {
		var user User
		err := c.Bind(&user)
		if err != nil {
			fmt.Println(err)
			log.Fatal(err)
		}

		username := user.Username
		if _, ok := account[username]; ok {
			message := "用户名" + username + "已存在"
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": message,
			})
		} else {
			account[username] = user
			c.JSON(http.StatusOK, gin.H{
				"code": 200,
				"message": "注册成功",
			})
		}
	})

	//重定向
	router.GET("/redict/baidu", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "https://www.baidu.com")
	})

	//管理组织分组api
	someGroup := router.Group("/hello")
	{
		someGroup.GET("/getting", getting)
		someGroup.POST("/posting", posting)
	}

	router.GET("/auth/signin", func(c *gin.Context) {
		cookie := &http.Cookie{
			Name:     "session_id",
			Value:    "123",
			Path:     "/",
			HttpOnly: true,
		}
		http.SetCookie(c.Writer, cookie)
		c.String(http.StatusOK, "Login successful")
	})

	router.GET("/home", AuthMiddleWare(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "home"})
	})

	router.LoadHTMLGlob("templates/*")
	//router.LoadHTMLFiles("templates/template1.html", "templates/template2.html")
	router.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "My WebSite",
		})
	})

	router.Run(":8080")
}

//鉴权中间间
func AuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		if cookie, err := c.Request.Cookie("session_id"); err == nil {
			value := cookie.Value
			fmt.Println(value)
			if value == "123" {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		c.Abort()
		return
	}
}

func posting(c *gin.Context)  {
	username := c.DefaultPostForm("username", "guest")//可设置默认值
	msg := c.PostForm("msg")
	title := c.PostForm("title")
	fmt.Printf("username is %s, msg is %s, title is %s\n", username, msg, title)
}

func getting(c *gin.Context)  {
	name := c.DefaultQuery("name", "Guest") //可设置默认值
	// 是 c.Request.URL.Query().Get("lastname") 的简写
	lastname := c.Query("lastname")
	c.String(http.StatusOK, "Hello %s \n", name);
	fmt.Printf("Hello %s \n", name)
	fmt.Printf("Hello %s \n", lastname)
}

type User struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password   string `form:"password" json:"password" bdinding:"required"`
	Age      int    `form:"age" json:"age"`
}