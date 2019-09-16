package handler

import (
	"filestore-server/db"
	"filestore-server/util"
	"fmt"
	"github.com/garyburd/redigo/redis"
	"math"
	"net/http"
	"os"
	"path"
	"strconv"
	rPool"filestore-server/cache/redis"
	"strings"
	"time"
)
//初始化信息
type MultipartUploadInfo struct {
	FileHash string
	FileSize int
	UploadID string
	ChunkSize int
	ChunkCount int
}
//初始化分块上传
func InitialMultipartUploadHandler(w http.ResponseWriter,r*http.Request)  {
	//1.解析用户请求参数
	r.ParseForm()
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filesize ,err:= strconv.Atoi(r.Form.Get("filesize"))
	if err != nil {
		w.Write(util.NewRespMsg(-1,"params invalid",nil).JSONBytes())
		return
	}
	//2.获得Redis的一个连接。
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()
	//3. 生成分块上传的初始化信息
	upInfo := MultipartUploadInfo{
		FileHash:   filehash,
		FileSize:   filesize,
		UploadID:   username+fmt.Sprintf("%x",time.Now().UnixNano()),
		ChunkSize:  5 * 1024 * 1024,//5m
		ChunkCount: int(math.Ceil(float64(filesize)/(5*1024*1024))),
	}
	//4.将初始化信息写入reids缓存
	rConn.Do("HSET","MP_"+upInfo.UploadID,"chunkcount",upInfo.ChunkCount)
	rConn.Do("HSET","MP_"+upInfo.UploadID,"filehash",upInfo.FileHash)
	rConn.Do("HSET","MP_"+upInfo.UploadID,"filesize",upInfo.FileSize)


	//5. 将响应初始化数据返回到客户端
	w.Write(util.NewRespMsg(0,"ok",upInfo).JSONBytes())
}
//上传文件分块
func UploadPartHandler(w http.ResponseWriter,r *http.Request)  {
	//1.解析用户请求参数
	r.ParseForm()
	//username := r.Form.Get("username")
	uploadID := r.Form.Get("uploadid")
	chunkIndex := r.Form.Get("index")
	//2.获得Redis连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()
	//3.获得文件句柄，用于存储分块内容
	fpath := "/Users/czf/temp/data/" + uploadID + "/" + chunkIndex
	os.MkdirAll(path.Dir(fpath), 0744)
	fd, err := os.Create(fpath)
	if err != nil {
		w.Write(util.NewRespMsg(-1,"upload part failed",nil).JSONBytes())
		return
	}
	defer fd.Close()
	buf := make([]byte,1024*1024)
	for  {
		n, err := r.Body.Read(buf)
		fd.Write(buf[:n])
		if err != nil {
			break
		}
	}
	//4.更新Redis缓存状态
	rConn.Do("HSET","MP_"+uploadID,"chkidx_"+chunkIndex,1)
	//5.返回处理结果到客户端
	w.Write(util.NewRespMsg(0,"ok",nil).JSONBytes())
}
func CompleteUploadHandler(w http.ResponseWriter,r *http.Request)  {
	//1.解析用户请求参数
	username := r.Form.Get("username")
	upid := r.Form.Get("uploadid")
	filename := r.Form.Get("filename")
	filehash := r.Form.Get("filehash")
	filesize ,_:= strconv.Atoi(r.Form.Get("filesize"))
	//2.获得Redis连接池中的一个连接
	rConn := rPool.RedisPool().Get()
	defer rConn.Close()
	//3.通过uoloadid查询Redis并判断是否所有分块上传完成
	 data, err := redis.Values(rConn.Do("HGETALL","MP_"+upid))
	 if err != nil{
	 	w.Write(util.NewRespMsg(-1,"complete upload failed",nil).JSONBytes())
		 return
	 }
	 totalCount := 0
	 chunkCount := 0
	for i:=0;i < len(data) ;i+=2  {
		k := string(data[i].([]byte))
		v := string(data[i+1].([]byte))
		if k == "chunkcount"{
			totalCount,_ = strconv.Atoi(v)
		}else if strings.HasPrefix(k,"chkidx_")&&v == "1" {
			chunkCount++
		}
	}
	if totalCount != chunkCount{
		w.Write(util.NewRespMsg(-2,"invalid request",nil).JSONBytes())
		return
	}
	//4.TODO: 合并分块
	//5.更新唯一文件表以及用户文件表
	db.OnFileUploadFinished(filehash,filename,int64(filesize),"")
	db.OnUserFileUploadFinished(username,filehash,filename,int64(filesize))
	//6.响应处理结果
	w.Write(util.NewRespMsg(0,"ok",nil).JSONBytes())

}