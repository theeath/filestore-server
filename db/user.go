package db

import (

	mydb "filestore-server/db/mysql"
	"fmt"
)
//用户注册接口
func UserSignUp(username string, passwd string)bool  {
	stmt, err := mydb.DBConn().Prepare(
		"insert ignore into tbl_user (user_name,user_pwd) values (?,?)")
	if err != nil {
		fmt.Printf("failed to insert, err:%s",err.Error())
		return false
	}
	ret,err := stmt.Exec(username,passwd)
	if err != nil{
		fmt.Printf("failed to insert ,err :%s",err.Error())
		return false
	}
	if rowsAffected, err := ret.RowsAffected();err == nil&&rowsAffected>0 {
		return true
		
	}
	return false
}
//用户登录接口,判断密码是否一致
func UserSignIn(username,encpwd string)  bool{
	stmt, err := mydb.DBConn().Prepare(
		"select * from tbl_user where user_name = ? limit 1")
	if err != nil{
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()
	rows, err := stmt.Query(username)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}else if rows == nil {
		fmt.Println("username not found:"+username)
		return false
	}
	pRows := mydb.ParseRows(rows)
	if len(pRows)>0&&string(pRows[0]["user_pwd"].([]byte)) == encpwd{
		return true
	}
	return false
}
//replace into replace into 跟 insert 功能类似，
// 不同点在于：replace into 首先尝试插入数据到表中，
// 1. 如果发现表中已经有此行数据（根据主键或者唯一索引判断）
// 则先删除此行数据，然后插入新的数据。
// 2. 否则，直接插入新数据。
//要注意的是：插入数据的表必须有主键或者是唯一索引！
// 否则的话，replace into 会直接插入数据，这将导致表中出现重复的数据。

//刷新用户的token
func UpdateToken(username string,token string)bool  {
	stmt, err := mydb.DBConn().Prepare(
		"replace into tbl_user_token (user_name,user_token) values (?,?)")
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	defer stmt.Close()
	_,err = stmt.Exec(username,token)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true

}

type User struct {
	Username string
	Email string
	Phone string
	SignupAt string
	LastActiveAt string
	Status int
}

func GetUserInfo(username string)(User,error)  {
	user := User{}
	stmt, err := mydb.DBConn().Prepare(
		"select user_name,signup_at from tbl_user where user_name = ? limit 1")
	if err != nil {
		fmt.Println(err.Error())
		return user,err
	}
	defer stmt.Close()
	err = stmt.QueryRow(username).Scan(&user.Username,&user.SignupAt)
	if err != nil{
		return user,err
	}
	return user,nil

}
