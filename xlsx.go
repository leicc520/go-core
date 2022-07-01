package core

import (
	"errors"
	"fmt"
	"time"
	
	"github.com/leicc520/go-orm"
	"github.com/leicc520/go-orm/log"
	"github.com/gin-gonic/gin"
	"github.com/tealeg/xlsx"
)

type XlsFormatRow func([]string) orm.SqlMap

//解析导出excel数据信息
func ExportToExcel(c *gin.Context, data []orm.SqlMap, header orm.SqlMap, sorts []string) error {
	var row *xlsx.Row = nil
	file := xlsx.NewFile()
	sheet, err := file.AddSheet("Sheet1")
	if err != nil {
		log.Write(log.ERROR, "添加业信息错误", err)
		return err
	}
	row  = sheet.AddRow()
	row.SetHeightCM(0.5)
	for _, key := range sorts {
		if _, ok := header[key]; ok {
			cell := row.AddCell()
			cell.Value = fmt.Sprintf("%v", header[key])
			cell.GetStyle().Alignment.Horizontal = "center"
			cell.GetStyle().Alignment.Vertical   = "center"
		}
	}
	for _, item := range data {
		row = sheet.AddRow()
		for _, key := range sorts {
			if str, ok := item[key]; ok {
				cell := row.AddCell()
				cell.Value = fmt.Sprintf("%v", str)
			}
		}
	}
	dateStr := time.Now().Format(orm.DATEYMDFormat)+orm.RandString(6)+".xlsx"
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition","attachment; filename="+dateStr)
	c.Header("Content-Transfer-Encoding", "binary")
	_ = file.Write(c.Writer)
	return nil
}

//读取xls到数组当中
func ReadExcel(file string, f XlsFormatRow, list []orm.SqlMap) ([]orm.SqlMap, error) {
	xlsFile, err := xlsx.OpenFile(file)
	if err != nil {
		log.Write(log.ERROR, "文件打开失败", err)
		return nil, errors.New("文件打开失败")
	}
	for _, sheet := range xlsFile.Sheets {
		for _, row := range sheet.Rows {
			var arrStr []string
			for _, cell := range row.Cells {
				lStr := cell.String()
				arrStr = append(arrStr, lStr)
			}
			if item := f(arrStr); item != nil {
				list = append(list, item)
			}
		}
	}
	return list, nil
}