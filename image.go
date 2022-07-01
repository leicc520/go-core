package core

import (
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"strings"

	"github.com/leicc520/go-orm/log"
	"github.com/nfnt/resize"
	"golang.org/x/image/bmp"
)

type ImageSt struct {
	file string
}

//获取图片数据资料信息
func NewImageSt(filePath string) *ImageSt {
	return &ImageSt{file: filePath}
}

//图片裁剪 规则:如果精度为0则精度保持不变 返回:error
func (s *ImageSt) Clip(x0, y0, x1, y1, quality int) (string, error) {
	dst := strings.Replace(s.file, ".", "_cp.", 1)
	in, err1  := os.Open(s.file)
	out, err2 := os.Create(dst)
	defer func() {//处理退出之后关闭文件
		if in != nil {
			in.Close()
		}
		if out != nil {
			out.Close()
		}
		if err := recover(); err != nil {
			log.Write(log.ERROR, err)
		}
	}()
	if err1 != nil || err2 != nil {
		return "", errors.New("打开图片文件失败")
	}
	origin, fm, err := image.Decode(in)
	if err != nil {
		return "", err
	}

	switch fm {
	case "jpeg":
		img := origin.(*image.YCbCr)
		subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.YCbCr)
		err  = jpeg.Encode(out, subImg, &jpeg.Options{quality})
	case "png":
		switch origin.(type) {
		case *image.NRGBA:
			img := origin.(*image.NRGBA)
			subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.NRGBA)
			err = png.Encode(out, subImg)
		case *image.RGBA:
			img := origin.(*image.RGBA)
			subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.RGBA)
			err = png.Encode(out, subImg)
		}
	case "gif":
		img := origin.(*image.Paletted)
		subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.Paletted)
		err = gif.Encode(out, subImg, &gif.Options{})
	case "bmp":
		switch origin.(type) {
		case *image.NRGBA:
			img := origin.(*image.NRGBA)
			subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.NRGBA)
			err = bmp.Encode(out, subImg)
		case *image.RGBA:
			img := origin.(*image.RGBA)
			subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.RGBA)
			err = bmp.Encode(out, subImg)
		}
	default:
		err = errors.New("ERROR FORMAT")
	}
	return dst, err
}

//缩略图生成 规则: 如果width 或 hight其中有一个为0，则大小不变 如果精度为0则精度保持不变
func (s *ImageSt) Scale(width, height, quality int) (string, error) {
	dst := strings.Replace(s.file, ".", "_sm.", 1)
	in, err1  := os.Open(s.file)
	out, err2 := os.Create(dst)
	defer func() {//处理退出之后关闭文件
		if in != nil {
			in.Close()
		}
		if out != nil {
			out.Close()
		}
		if err := recover(); err != nil {
			log.Write(log.ERROR, err)
		}
	}()
	if err1 != nil || err2 != nil {
		return "", errors.New("打开图片文件失败")
	}
	origin, fm, err := image.Decode(in)
	if err != nil {
		return "", err
	}
	if width == 0 || height == 0 {
		width  = origin.Bounds().Max.X
		height = origin.Bounds().Max.Y
	}
	if quality == 0 {
		quality = 100
	}
	canvas := resize.Thumbnail(uint(width), uint(height), origin, resize.Lanczos3)
	switch fm {
	case "jpeg":
		err = jpeg.Encode(out, canvas, &jpeg.Options{quality})
	case "png":
		err = png.Encode(out, canvas)
	case "gif":
		err = gif.Encode(out, canvas, &gif.Options{})
	case "bmp":
		err = bmp.Encode(out, canvas)
	default:
		err = errors.New("ERROR FORMAT")
	}
	return dst, err
}

