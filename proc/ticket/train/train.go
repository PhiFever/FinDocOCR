package train

import (
	"FinDocOCR/config"
	"FinDocOCR/doctype"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/xuri/excelize/v2"
	"log"
	"strings"
	"time"
)

var logger = config.GetLogger()

// Doc represents train ticket data
type Doc struct {
	Name               string
	StartDate          string
	StartingStation    string
	ArrivalDate        string
	DestinationStation string
	SeatCategory       string
	TicketRates        string
}

func (d *Doc) String() string {
	return fmt.Sprintf("Name: %s, StartDate: %s, StartingStation: %s, ArrivalDate: %s, DestinationStation: %s, SeatCategory: %s, TicketRates: %s",
		d.Name, d.StartDate, d.StartingStation, d.ArrivalDate, d.DestinationStation, d.SeatCategory, d.TicketRates)
}

func (d *Doc) AmendData() {
	// 替换日期格式
	d.StartDate = strings.ReplaceAll(d.StartDate, "年", ".")
	d.StartDate = strings.ReplaceAll(d.StartDate, "月", ".")
	d.StartDate = strings.ReplaceAll(d.StartDate, "日", "")

	// 更新座位类别
	if d.SeatCategory == "新空调硬卧" {
		d.SeatCategory = "火车(硬卧)"
	} else if d.SeatCategory == "二等座" {
		d.SeatCategory = "动车（二等座）"
	}

	if d.ArrivalDate != "" {
		d.ArrivalDate = strings.ReplaceAll(d.ArrivalDate, "年", ".")
		d.ArrivalDate = strings.ReplaceAll(d.ArrivalDate, "月", ".")
		d.ArrivalDate = strings.ReplaceAll(d.ArrivalDate, "日", "")
	} else {
		// 如果没有抵达日期，默认设置为出发日期
		d.ArrivalDate = d.StartDate
		if d.SeatCategory == "火车(硬卧)" {
			// 抵达日期为之后一天
			startDateTime, err := time.Parse("2006.01.02", d.StartDate)
			if err == nil {
				arrivalDateTime := startDateTime.AddDate(0, 0, 1)
				d.ArrivalDate = arrivalDateTime.Format("2006.01.02")
			}
		}
	}

	// 去除票价中的符号
	d.TicketRates = strings.ReplaceAll(d.TicketRates, "￥", "")
	d.TicketRates = strings.ReplaceAll(d.TicketRates, "元", "")
}

type Processor struct{}

func (p *Processor) Process(data []byte) (doctype.Document, error) {
	d := Doc{}
	wordsResult := gjson.GetBytes(data, "words_result.0.result")
	if !wordsResult.Exists() {
		errCode := gjson.GetBytes(data, "error_code").String()
		return &d, fmt.Errorf("no words_result in response, error_code: %s", errCode)
	}

	// 修改解析路径，获取数组中的 word 字段
	d.Name = wordsResult.Get("name.0.word").String()
	d.StartDate = wordsResult.Get("date.0.word").String()
	d.StartingStation = wordsResult.Get("starting_station.0.word").String()
	d.DestinationStation = wordsResult.Get("destination_station.0.word").String()
	d.SeatCategory = wordsResult.Get("seat_category.0.word").String()
	d.TicketRates = wordsResult.Get("ticket_rates.0.word").String()
	d.ArrivalDate = ""
	d.AmendData()

	logger.Info("Train ticket data: ", d)
	return &d, nil
}

type Docs []Doc

func (docs *Docs) Add(doc doctype.Document) {
	d, ok := doc.(*Doc)
	if !ok {
		logger.Error("Failed to assert Doc type")
		return
	}

	*docs = append(*docs, *d)
}

func (docs *Docs) SaveToFile() error {
	// 初始化 Excel 文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// 设置常量
	const (
		filename  = "火车票处理结果.xlsx"
		sheetName = "Sheet1"
	)

	// 定义表头
	headers := []string{
		"*人员",
		"*出发日期",
		"*起始地",
		"*抵达日期",
		"*目的地",
		"*交通工具",
		"*票价",
	}

	// 写入表头
	for i, header := range headers {
		cell, err := excelize.CoordinatesToCellName(i+1, 1)
		if err != nil {
			return fmt.Errorf("failed to convert coordinates for header %d: %w", i+1, err)
		}

		if err := f.SetCellValue(sheetName, cell, header); err != nil {
			return fmt.Errorf("failed to write header at cell %s: %w", cell, err)
		}
	}

	// 写入数据行
	for rowNum, ticket := range *docs {
		rowValues := []interface{}{
			ticket.Name,
			ticket.StartDate,
			ticket.StartingStation,
			ticket.ArrivalDate,
			ticket.DestinationStation,
			ticket.SeatCategory,
			ticket.TicketRates,
		}

		// 写入每一列的数据
		for colNum, value := range rowValues {
			cell, err := excelize.CoordinatesToCellName(colNum+1, rowNum+2)
			if err != nil {
				return fmt.Errorf("failed to convert coordinates for row %d, column %d: %w",
					rowNum+2, colNum+1, err)
			}

			if err := f.SetCellValue(sheetName, cell, value); err != nil {
				return fmt.Errorf("failed to write value at cell %s: %w", cell, err)
			}
		}
	}

	// 保存文件
	if err := f.SaveAs(filename); err != nil {
		return fmt.Errorf("failed to save file %s: %w", filename, err)
	}

	return nil
}
