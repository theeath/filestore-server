package handler

import (
	"encoding/json"
	"filestore-server/common"
	"filestore-server/config"
	"filestore-server/db"
	"filestore-server/meta"
	"filestore-server/mq"
	"filestore-server/store/oss"
	"filestore-server/util"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)
//处理文件上传
func UploadHandler(w http.ResponseWriter, r *http.Request)  {
	if r.Method == "GET"{
		//返回上传html页面
		data, err := ioutil.ReadFile("../../static/view/index.html")
		if err != nil {
			io.WriteString(w,err.Error())
			return
		}
		io.WriteString(w,string(data))
	}else if r.Method == "POST" {
		//接收文件流即存储到本地目录
		file, head, err := r.FormFile("file")
		if err != nil {
			fmt.Printf("failed to get data ,err:%s\n",err.Error())
			return
		}
		defer file.Close()
		fileMeta := meta.FileMeta{
			FileName:head.Filename,
			Location:"/Users/czf/temp/"+head.Filename,
			UploadAt:time.Now().Format("2006-01-02 15:04:05"),
		}
		//os.MkdirAll("")
		newFile, err := os.Create(fileMeta.Location)
		if err != nil {
			fmt.Printf("failed to create file, err:%s\n",err.Error())
		}
		defer newFile.Close()
		fileMeta.FileSize,err = io.Copy(newFile,file)
		if err != nil {
			fmt.Printf("failed to save data to file ,err:%s\n",err.Error())
			return
		}
		newFile.Seek(0,0)
		fileMeta.FileSha1 = util.FileSha1(newFile)
		newFile.Seek(0, 0)
		//写入oss
		ossPath := "oss/"+fileMeta.FileSha1
		//err = oss.Bucket().PutObject(ossPath,newFile)
		//if err != nil {
		//	fmt.Println(err.Error())
		//	w.Write([]byte("upload failed"))
		//	return
		//}
		//fileMeta.Location = ossPath
		data := mq.TransferData{
			FileHash:      fileMeta.FileSha1,
			CurLocation:   fileMeta.Location,
			DestLocation:  ossPath,
			DestStoreType: common.StoreOSS,
		}
		pubData,_ := json.Marshal(data)
		suc := mq.Publish(config.TransExchangeName,config.TransOSSRoutingKey,pubData)
		if !suc{
			//TODO:加入重拾发送消息逻辑
		}
		//meta.UpdateFileMeta(fileMeta)
		_ = meta.UpdateFileMetaDB(fileMeta)
		//:更新用户文件表记录
		username := r.Form.Get("username")
		suc = db.OnUserFileUploadFinished(username,fileMeta.FileSha1,fileMeta.FileName,fileMeta.FileSize)
		if suc{
			http.Redirect(w,r,"/static/view/home.html",http.StatusFound)
		}else {
			w.Write([]byte("upload failed"))
		}

	}
}
//上传已完成
func UploadScuHandler(w http.ResponseWriter, r *http.Request)  {
	io.WriteString(w,"upload finished")
}
//:获取文件元信息
func GetFileMetaHandler(w http.ResponseWriter,r *http.Request)  {
	//解析url传递的参数，对于POST则解析响应包的主体（request body）
	r.ParseForm()

	filehash := r.Form["filehash"][0]
	//fMeta := meta.GetFileMeta(filehash)
	fMeta,err := meta.GetFileMetaDB(filehash)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(fMeta)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
//查询批量文件元信息
func FileQueryHandler(w http.ResponseWriter,r *http.Request)  {
	r.ParseForm()
	limitCnt,_ := strconv.Atoi(r.Form.Get("limit"))
	username := r.Form.Get("username")
	userFiles,err := db.QueryUserFileMetas(username,limitCnt)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	data, err := json.Marshal(userFiles)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(data)
}
//下载的接口
func DownLoadHandler(w http.ResponseWriter,r *http.Request)  {
	r.ParseForm()
	fsha1 := r.Form.Get("filehash")
	//根据sha1获取文件元信息
	 fm := meta.GetFileMeta(fsha1)
	 f, err := os.Open(fm.Location)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	 defer f.Close()
	 data, err := ioutil.ReadAll(f)
	 if err != nil{
		 w.WriteHeader(http.StatusInternalServerError)
		 return
	 }
	 w.Header().Set("Content-Type","application/octect-stream")
	 w.Header().Set("Content-Descrption","attachment;filename=\""+fm.FileName+"\"")

	 w.Write(data)
}
//更新元信息接口（重命名）
func FileMetaUpdateHandler(w http.ResponseWriter,r *http.Request)  {
	r.ParseForm()
	opType := r.Form.Get("op")
	filesha1 := r.Form.Get("filehash")
	newFileName := r.Form.Get("filename")
	if opType != "0"{
		w.WriteHeader(http.StatusForbidden)
		return
	}
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	curFileMeta := meta.GetFileMeta(filesha1)
	curFileMeta.FileName = newFileName
	meta.UpdateFileMeta(curFileMeta)
	w.WriteHeader(http.StatusOK)
	data, err := json.Marshal(curFileMeta)
	if err != nil{
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
//删除文件及元信息
func FileDeleteHandler(w http.ResponseWriter, r *http.Request)  {
	r.ParseForm()
	filesha1 := r.Form.Get("filehash")
	fMeta := meta.GetFileMeta(filesha1)
	os.Remove(fMeta.Location)
	meta.RemoveFileMeta(filesha1)
	w.WriteHeader(http.StatusOK)
}
//尝试秒传接口
func TryFastUploadHandler(w http.ResponseWriter,r *http.Request)  {
	r.ParseForm()
	//解析请求参数
	username := r.Form.Get("username")
	filehash := r.Form.Get("filehash")
	filename := r.Form.Get("filename")
	filesize, _ := strconv.Atoi(r.Form.Get("filesize"))
	//从文件表中查询相同hash的文件记录
	fileMeta, err := meta.GetFileMetaDB(filehash)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//查不到记录则返回秒传失败
	if fileMeta == nil {
		resp :=util.RespMsg{
			Code: -1,
			Msg:  "秒传失败，请访问普通上传接口",
		}
		w.Write(resp.JSONBytes())
		return
	}
	//上传过则将文件信息写入用户文件表，返回成功
	suc := db.OnUserFileUploadFinished(username,filehash,filename,int64(filesize))
	if suc {
		resp := util.RespMsg{
			Code: 0,
			Msg:  "秒传成功",
		}
		w.Write(resp.JSONBytes())
		return
	}else {
		resp := util.RespMsg{
			Code: -2,
			Msg:  "秒传失败，稍后重试",
			Data: nil,
		}
		w.Write(resp.JSONBytes())
		return
	}
}
//生成oss文件的下载地址
func DownloadURLHandler(w http.ResponseWriter,r *http.Request)  {
	r.ParseForm()
	filehash := r.Form.Get("filehash")
	//从文件表查找记录
	row,_ := db.GetFileMeta(filehash)
	signedURL := oss.DownloadURL(row.FileAddr.String)
	w.Write([]byte(signedURL))

}
