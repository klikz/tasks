package main

import (
	"encoding/base64"
	"fmt"
	"image/jpeg"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Request struct {
	Image string `json:"image"`
}

type Responce struct {
	Result string      `json:"result"`
	Err    interface{} `json:"error"`
	Data   interface{} `json:"data"`
}

func main() {
	r := gin.Default()
	port := ":8000" //или в .env

	r.POST("/recieve_base64_jpg", func(c *gin.Context) {
		req := Request{}
		resp := Responce{}

		if err := c.ShouldBind(&req); err != nil {
			resp.Result = "error"
			resp.Err = err.Error()
			c.JSON(http.StatusBadRequest, resp)
			c.Abort()
			return
		}
		reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(req.Image))
		image, err := jpeg.Decode(reader)
		if err != nil {
			resp.Result = "error"
			resp.Err = err.Error()
			c.JSON(http.StatusBadRequest, resp)
			c.Abort()
			return
		}
		fileName := uuid.New().String() + ".jpg"
		imageFileName := `d:\some_folder\media\` + fileName

		f, err := os.OpenFile(imageFileName, os.O_WRONLY|os.O_CREATE, 0777)

		if err != nil {
			resp.Result = "error"
			resp.Err = err.Error()
			c.JSON(http.StatusBadRequest, resp)
			c.Abort()
			return
		}
		defer f.Close()

		err = jpeg.Encode(f, image, &jpeg.Options{Quality: 75})
		if err != nil {
			resp.Result = "error"
			resp.Err = err.Error()
			c.JSON(http.StatusBadRequest, resp)
			c.Abort()
			return
		}

		resp.Result = "ok"
		c.JSON(http.StatusOK, resp)

	})

	r.POST("/recieve_binary", func(c *gin.Context) {
		resp := Responce{}
		file, err := c.FormFile("image")
		if err != nil {
			resp.Result = "error"
			resp.Err = err.Error()
			c.JSON(http.StatusBadRequest, resp)
			c.Abort()
			return
		}
		fmt.Println(file.Filename)

		temp, err := file.Open()
		if err != nil {
			resp.Result = "error"
			resp.Err = err.Error()
			c.JSON(http.StatusBadRequest, resp)
			c.Abort()
			return
		}

		localFileName := "d:\\some_folder\\media\\" + file.Filename //todo сделать логика для избежания конфиликта имен файлов

		out, err := os.Create(localFileName)
		if err != nil {
			resp.Result = "error"
			resp.Err = err.Error()
			c.JSON(http.StatusBadRequest, resp)
			c.Abort()
			return
		}

		defer out.Close()
		_, err = io.Copy(out, temp)
		if err != nil {
			resp.Result = "error"
			resp.Err = err.Error()
			c.JSON(http.StatusBadRequest, resp)
			c.Abort()
			return
		}

		resp.Result = "ok"
		c.JSON(http.StatusOK, resp)

	})

	r.GET("/file_list", func(c *gin.Context) {
		files, err := ioutil.ReadDir("d:\\some_folder\\media\\")
		resp := Responce{}

		if err != nil {
			resp.Result = "error"
			resp.Err = err.Error()
			c.JSON(http.StatusBadRequest, resp)
			c.Abort()
			return
		}
		type FileData struct {
			FileName  string    `json:"file_name"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		}

		filesInfo := []FileData{}
		for _, f := range files {
			fileInfo := FileData{}
			d := f.Sys().(*syscall.Win32FileAttributeData)
			fileInfo.FileName = f.Name()
			fileInfo.CreatedAt = time.Unix(0, d.CreationTime.Nanoseconds())
			fileInfo.UpdatedAt = time.Unix(0, d.LastWriteTime.Nanoseconds())

			filesInfo = append(filesInfo, fileInfo)
		}

		resp.Result = "ok"
		resp.Data = filesInfo

		c.JSON(http.StatusOK, resp)
	})

	r.GET("/file_get/:name", func(c *gin.Context) {
		type FileName struct {
			Name string `uri:"name" binding:"required"`
		}
		var fileName FileName
		if err := c.ShouldBindUri(&fileName); err != nil {
			resp := Responce{}
			resp.Result = "error"
			resp.Err = err.Error()
			c.JSON(http.StatusBadRequest, resp)
			c.Abort()
			return
		}
		fmt.Println("d:\\some_folder\\media\\" + fileName.Name)
		c.File("d:\\some_folder\\media\\" + fileName.Name)

	})

	// todo
	// создать базу пользователей с токенами и int значением для загрузку и для получения листа
	// при запросе на загрузку/скачивание в начале про верить инт пользователя для загрузки на < 10
	// если true то ++
	// если false return errors.New("превышен лимит")
	// получения листа r.GET("/file_list" условия < 100

	r.Run(port)
}
