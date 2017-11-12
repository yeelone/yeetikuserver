package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

type FileInfoResponse struct {
	UID      int64  `json:"uid"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Url      string `json:"url"`
	Response string `json:"response"`
}

func MkDir(path string) error {
	if _, err := os.Stat(path); err == nil {
		return nil
	}

	err := os.MkdirAll(path, 0711)
	if err != nil {
		return err
	}
	return nil

}

/* @params : dir 表示在/upload/ 下的目录
@params : key 表示formdata 中的key
todo: 要加入限制图片大小的功能
*/
func uploadImage(r *http.Request, dir string, key string) (response FileInfoResponse, err error) {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile(key)
	if err != nil {
		return response, err
	}
	defer file.Close()

	path := "/upload/img/" + dir
	if err = MkDir("." + path); err != nil {
		return response, err
	}

	filename, _ := renameFile(handler.Filename)
	//filename := handler.Filename
	fullfilename := path + "/" + filename
	targetPath := "."
	url := "/static/" + dir + "/" + filename
	f, err := os.OpenFile(targetPath+fullfilename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return response, err
	}
	defer f.Close()
	io.Copy(f, file)

	response = FileInfoResponse{UID: 1, Name: filename, Status: "done", Url: url, Response: "hello"}
	return response, nil
}

/*@return
**	path
**	filename
**	fileSuffix
**	err
 */
func uploadFile(r *http.Request, dir string, key string) (path, filename, fileSuffix string, err error) {
	r.ParseMultipartForm(32 << 20)
	file, handler, err := r.FormFile(key)
	if err != nil {
		fmt.Println(err)
		return "", "", "", err
	}
	defer file.Close()

	path = "/upload/file/" + dir
	if err = MkDir("." + path); err != nil {
		return "", "", "", err
	}

	filename, fileSuffix = renameFile(handler.Filename)
	fullfilename := path + "/" + filename
	targetPath := "."
	f, err := os.OpenFile(targetPath+fullfilename, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println(err)
		return "", "", "", err
	}
	defer f.Close()
	io.Copy(f, file)

	return path, filename, fileSuffix, nil
}

func renameFile(filename string) (name, subffix string) {
	t := strconv.FormatInt(time.Now().UnixNano(), 10)

	var filenameWithSuffix string
	filenameWithSuffix = path.Base(filename)

	var fileSuffix string
	fileSuffix = path.Ext(filenameWithSuffix) //获取文件后缀

	var filenameOnly string
	filenameOnly = strings.TrimSuffix(filenameWithSuffix, fileSuffix) //获取文件名

	return filenameOnly + t + fileSuffix, fileSuffix

}
