package main

import (
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"image/color"
	"image/jpeg"

	"github.com/golang/freetype"
	"github.com/xuri/excelize/v2"
)

func main() {
	fileName := ""
	textMark := ""
	if len(os.Args) > 1 {
		fileName = os.Args[1]
	} else {
		fmt.Println("请输入文件位置：")
		fmt.Scanf("%s\n", &fileName)
	}
	fmt.Println("请输入水印文本：")
	fmt.Scanf("%s\n", &textMark)
	f, err := excelize.OpenFile(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName, excelize.Options{RawCellValue: true})
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Print()
	for i, row := range rows {
		for j := range row {
			x, _ := excelize.ColumnNumberToName(j + 1)
			axis := x + strconv.Itoa(i+1)
			link, err := f.GetCellFormula(sheetName, axis)
			if err == nil && link != "" {
				f.SetRowHeight(sheetName, i+1, 100)
				err := f.SetCellValue(sheetName, axis, " ")
				if err != nil {
					panic("del data error")
				}
				link = strings.Split(link, "\"")[1]
				fmt.Println("正在下载" + link)
				r, err := http.Get(link)
				if err != nil {
					panic("下载图片错误")
				}
				defer r.Body.Close()
				f.AddPictureFromBytes(sheetName, axis, `{
					"autofit":true
				}`, axis, ".jpg", imgMark(r.Body, textMark))
			}
		}
	}
	// 保存文件
	if err = f.Save(); err != nil {
		fmt.Println(err)
	}
	if err = f.Close(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("完成")
	select {}
}

func imgMark(r io.Reader, str string) (data []byte) {
	jpgimg, _ := jpeg.Decode(r)
	img := image.NewNRGBA(jpgimg.Bounds())
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, jpgimg.At(x, y))
		}
	}

	fontBytes, err := ioutil.ReadFile(filepath.Dir(os.Args[0]) + "\\font\\HarmonyOS_Sans_SC_Bold.ttf")
	if err != nil {
		fmt.Println(err)
	}

	font, err := freetype.ParseFont(fontBytes)
	if err != nil {
		fmt.Println(err)
	}

	f := freetype.NewContext()
	f.SetDPI(75)
	f.SetFont(font)
	f.SetFontSize(60)
	f.SetClip(jpgimg.Bounds())
	f.SetDst(img)
	f.SetSrc(image.NewUniform(color.RGBA{R: 255, G: 0, B: 0, A: 255}))

	pt := freetype.Pt(20, img.Bounds().Dy()-30)
	f.DrawString(str, pt)

	newfile, _ := os.Create("temp.jpg")
	defer newfile.Close()

	err = jpeg.Encode(newfile, img, &jpeg.Options{Quality: 80})

	if err != nil {
		fmt.Println(err)
	}

	afterimg, _ := os.Open("temp.jpg")
	data, _ = ioutil.ReadAll(afterimg)

	return data
}
