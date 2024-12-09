package vat

import (
	"FinDocOCR/config"
	"FinDocOCR/doctype"
	"fmt"
	"github.com/tidwall/gjson"
	"github.com/xuri/excelize/v2"
	"log"
	"strings"
)

var logger = config.GetLogger()

// Doc represents VAT Doc data
type Doc struct {
	DocCode          string
	DocNumber        string
	Date             string
	CommodityName    string
	TotalAmount      string
	CommodityTaxRate string
	TotalTax         string
}

func (d *Doc) String() string {
	return fmt.Sprintf("DocCode: %s, DocNumber: %s, Date: %s, CommodityName: %s, TotalAmount: %s, CommodityTaxRate: %s, TotalTax: %s",
		d.DocCode, d.DocNumber, d.Date, d.CommodityName, d.TotalAmount, d.CommodityTaxRate, d.TotalTax)
}

func (d *Doc) AmendData() {
	// 处理日期
	d.Date = strings.ReplaceAll(d.Date, "年", ".")
	d.Date = strings.ReplaceAll(d.Date, "月", ".")
	d.Date = strings.ReplaceAll(d.Date, "日", "")

	d.CommodityName = d.CommodityName[strings.LastIndex(d.CommodityName, "*")+1:]
}

type Processor struct{}

func (p *Processor) Process(data []byte) (doctype.Document, error) {
	d := Doc{}
	wordsResult := gjson.GetBytes(data, "words_result.0.result")
	if !wordsResult.Exists() {
		errCode := gjson.GetBytes(data, "error_code").String()
		return &d, fmt.Errorf("no words_result in response, error_code: %s", errCode)
	}

	d.DocCode = wordsResult.Get("InvoiceCodeConfirm.0.word").String()
	d.DocNumber = wordsResult.Get("InvoiceNumConfirm.0.word").String()
	d.Date = wordsResult.Get("InvoiceDate.0.word").String()
	d.CommodityName = wordsResult.Get("CommodityName.0.word").String()
	d.TotalAmount = wordsResult.Get("TotalAmount.0.word").String()
	d.CommodityTaxRate = wordsResult.Get("CommodityTaxRate.0.word").String()
	d.TotalTax = wordsResult.Get("TotalTax.0.word").String()
	d.AmendData()

	logger.Info("VAT Doc data: ", d)
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
		filename  = "增值税发票处理结果.xlsx"
		sheetName = "Sheet1" // Excel 默认的工作表名称
	)

	// 创建流式写入器，使用正确的 sheet 名称
	sw, err := f.NewStreamWriter(sheetName)
	if err != nil {
		return fmt.Errorf("failed to create stream writer: %w", err)
	}

	// 写入表头
	headers := []interface{}{
		"发票代码", "发票号码", "开票日期",
		"货物名称", "金额", "税率", "税额",
	}
	if err := sw.SetRow("A1", headers); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}

	// 写入数据行
	for i, d := range *docs {
		rowData := []interface{}{
			d.DocCode,
			d.DocNumber,
			d.Date,
			d.CommodityName,
			d.TotalAmount,
			d.CommodityTaxRate,
			d.TotalTax,
		}

		if err := sw.SetRow(fmt.Sprintf("A%d", i+2), rowData); err != nil {
			return fmt.Errorf("failed to write row %d: %w", i+2, err)
		}
	}

	// 刷新流式写入器
	if err := sw.Flush(); err != nil {
		return fmt.Errorf("failed to flush stream writer: %w", err)
	}

	// 保存文件
	if err := f.SaveAs(filename); err != nil {
		return fmt.Errorf("failed to save file: %w", err)
	}

	return nil
}
