package main

import (
	"bufio"
	"encoding/json"
	"filestore-server/config"
	"filestore-server/db"
	"filestore-server/mq"
	"filestore-server/store/oss"
	"log"
	"os"
)
//处理文件转移的真正逻辑
func ProcessTransfer(msg []byte)bool  {
	log.Println(string(msg))
	//解析msg
	pubData := mq.TransferData{}
	err := json.Unmarshal(msg,&pubData)
	if err != nil{
		log.Println(err.Error())
		return false
	}
	//根据临时存储文件路径，创建文件句柄
	filed, err := os.Open(pubData.CurLocation)
	if err != nil{
		log.Println(err.Error())
		return false
	}
	//通过文件句柄将文件内容读出来，并且上传到oss
	err = oss.Bucket().PutObject(pubData.DestLocation,bufio.NewReader(filed))
	if err != nil{
		log.Println(err.Error())
		return false
	}
	//更新文件的存储路径到文件表
	suc := db.UpdateFileLocation(pubData.FileHash,pubData.DestLocation)
	if !suc{
		return false
	}
	return true
}
func main()  {
	log.Println("开始监听转移任务队列....")
	mq.StartConsume(
		config.TransOSSQueueName,
		"transfer_oss",ProcessTransfer)
}