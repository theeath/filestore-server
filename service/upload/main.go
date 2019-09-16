package main

import (
	"filestore-server/handler"
	"fmt"
	"net/http"
)

func main() {
	//路由规则
	http.Handle("/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir("../../static"))))
	http.HandleFunc("/file/upload",handler.HTTPInterceptor(handler.UploadHandler))

	http.HandleFunc("/file/upload/suc",handler.UploadScuHandler)
	http.HandleFunc("/file/meta",handler.HTTPInterceptor(handler.GetFileMetaHandler))
	http.HandleFunc("/file/download",handler.HTTPInterceptor(handler.DownLoadHandler))
	http.HandleFunc("/file/update",handler.HTTPInterceptor(handler.FileMetaUpdateHandler))
	http.HandleFunc("/file/fastupload",handler.HTTPInterceptor(handler.TryFastUploadHandler))
	http.HandleFunc("/file/downloadurl",handler.HTTPInterceptor(handler.DownloadURLHandler))

	http.HandleFunc("/file/delete",handler.HTTPInterceptor(handler.FileDeleteHandler))
	http.HandleFunc("/file/query",handler.HTTPInterceptor(handler.FileQueryHandler))
	http.HandleFunc("/user/signup",handler.SignUpHandler)
	http.HandleFunc("/user/signin",handler.SignInHandler)
	http.HandleFunc("/user/info",handler.HTTPInterceptor(handler.UserInfoHandler))

	//分块上传接口
	http.HandleFunc("/file/mpupload/init",handler.HTTPInterceptor(handler.InitialMultipartUploadHandler))
	http.HandleFunc("/file/mpupload/uppart",handler.HTTPInterceptor(handler.UploadPartHandler))
	http.HandleFunc("/file/mpupload/complete",handler.HTTPInterceptor(handler.CompleteUploadHandler))




	err := http.ListenAndServe(":8080",nil)
	if err != nil {
		fmt.Printf("failed to start server ,err:%s ",err.Error())
	}
}


//docker run -p 3308:3306 --name sql2 -e MYSQL_ROOT_PASSWORD=123456 -d mysql:5.7
//docker exec -it sql2 /bin/bash
//CREATE USER 'reader'@'192.168.0.5' IDENTIFIED WITH mysql_native_password BY 'reader';
//docker exec -it 627a2368c865 /bin/bash
//CHANGE MASTER TO MASTER_HOST='172.17.0.2',MASTER_USER='slave',MASTER_PASSWORD='123456',MASTER_LOG_FILE='mysql-bin.000001',MASTER_LOG_POS=154;
//GRANT REPLICATION slave, REPLICATION CLIENT ON *.* TO 'reader'@'%';
//grant replication slave on *.* to 'reader'@'192.168.0.5' identifiedby 'reader';
//create user 'reader'@'192.168.0.5' identified by 'reader';
//GRANT ALL PRIVILEGES ON *.* TO 'reader'@'192.168.0.5' WITH GRANT OPTION;
//CHANGE MASTER TO MASTER_LOG_FILE='mysql-bin.000002',MASTER_LOG_POS=155;
//docker run -d --hostname rabbit-svr --name rabbit -p 5672:5672 -p 15672:15672 -p 25672:25672 -v /data/rabbitmq:/var/lib/rabbitmq rabbitmq:management