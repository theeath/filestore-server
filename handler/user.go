package handler

import (
	"filestore-server/db"
	"filestore-server/util"
	"fmt"
	"net/http"
	"time"
)

const   (
	pwd_salt = "#890"
)
//处理用户注册请求
func SignUpHandler(w http.ResponseWriter, r *http.Request)  {
	if r.Method == http.MethodGet {
		//data,err := ioutil.ReadFile("./static/view/signup.html")
		//if err != nil {
		//	w.WriteHeader(http.StatusInternalServerError)
		//	return
		//}
		//w.Write(data)
		http.Redirect(w, r, "/static/view/signup.html", http.StatusFound)
		return
	}
	r.ParseForm()
	username := r.Form.Get("username")
	passwd := r.Form.Get("password")
	if len(username) < 3 || len(passwd) < 5{
		w.Write([]byte("invalid parameter"))
		return
	}
	encPasswd := util.Sha1([]byte(passwd+pwd_salt))
	suc := db.UserSignUp(username,encPasswd)
	if suc {
		w.Write([]byte("SUCCESS"))
	}else {
		w.Write([]byte("FAILED"))
	}

}
//处理用户登录请求
func SignInHandler(w http.ResponseWriter,r *http.Request)  {
	if r.Method == http.MethodGet {
		// data, err := ioutil.ReadFile("./static/view/signin.html")
		// if err != nil {
		// 	w.WriteHeader(http.StatusInternalServerError)
		// 	return
		// }
		// w.Write(data)
		http.Redirect(w, r, "/static/view/signin.html", http.StatusFound)
		return
	}

	r.ParseForm()
	username := r.Form.Get("username")
	passwd := r.Form.Get("password")
	encPasswd := util.Sha1([]byte(passwd+pwd_salt))
	//1.校验用户名和密码
	pwdChecked := db.UserSignIn(username,encPasswd)
	if !pwdChecked {
		w.Write([]byte("FAILED"))
		return
	}

	//2.生成访问凭证（token）
	token := GenToken(username)
	upRes := db.UpdateToken(username,token)
	if !upRes {
		w.Write([]byte("FAILED"))
		return
	}
	//3.登录成功后重定向到首页
	//w.Write([]byte("http://"+r.Host+"/static/view/home.html"))
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + r.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	w.Write(resp.JSONBytes())
}

// GenToken : 生成token
func GenToken(username string) string {
	// 40位字符:md5(username+timestamp+token_salt)+timestamp[:8]
	ts := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := util.MD5([]byte(username + ts + "_tokensalt"))
	return tokenPrefix + ts[:8]
}
//查询用户信息
func UserInfoHandler(w http.ResponseWriter,r *http.Request)  {
	//解析请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	//token := r.Form.Get("token")
	//验证token是否有效
	//isValidToken := IsTokenValid(token)
	//if !isValidToken {
	//	w.WriteHeader(http.StatusForbidden)
	//	return
	//}
	//查询用户
	user, err := db.GetUserInfo(username)

	if err != nil{
		w.WriteHeader(http.StatusForbidden)
		return
	}
	//组装并且响应用户数据
	resp := util.RespMsg{
		Code: 0,
		Msg:  "OK",
		Data: user,
	}
	w.Write(resp.JSONBytes())
}
//验证token是否有效
func IsTokenValid(token string)bool  {
	if len(token) != 40 {
		return false
	}
	// TODO: 判断token的时效性，是否过期
	// TODO: 从数据库表tbl_user_token查询username对应的token信息
	// TODO: 对比两个token是否一致
	return true
}