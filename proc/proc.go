package proc

import (
	"FinDocOCR/config"
	"FinDocOCR/doctype"
	"FinDocOCR/proc/invoice/vat"
	"FinDocOCR/proc/ticket/train"
	"fmt"
	"github.com/tidwall/gjson"
)

var logger = config.GetLogger()

// FinDocProcessor 票据处理器接口
type FinDocProcessor interface {
	Process(data []byte) (doctype.Document, error)
}

type ProcessorFactory struct {
	processors map[string]FinDocProcessor
}

func NewProcessorFactory() *ProcessorFactory {
	factory := &ProcessorFactory{
		processors: make(map[string]FinDocProcessor),
	}

	factory.processors[doctype.TypeVatInvoice] = &vat.Processor{}
	factory.processors[doctype.TypeTrainTicket] = &train.Processor{}
	// TODO:注册其他处理器...
	return factory
}

func (f *ProcessorFactory) GetProcessor(invoiceType string) (FinDocProcessor, error) {
	processor, exists := f.processors[invoiceType]
	if !exists {
		return nil, fmt.Errorf("unsupported invoice type: %s", invoiceType)
	}
	return processor, nil
}

func ProcessInvoice(data []byte) (doctype.Document, error) {
	factory := NewProcessorFactory()

	if !gjson.ValidBytes(data) {
		return nil, fmt.Errorf("invalid json data")
	}

	if gjson.GetBytes(data, "words_result_num").Int() != 1 {
		return nil, fmt.Errorf("该程序只支持单张票据的识别")
	}

	resultType := gjson.GetBytes(data, "words_result.0.type").String()
	logger.Info("resultType: ", resultType)
	processor, err := factory.GetProcessor(resultType)
	if err != nil {
		return nil, err
	}

	return processor.Process(data)
}

type DocumentFactory struct{}

func (f *DocumentFactory) CreateCollection(docType doctype.DocumentType) doctype.DocumentCollection {
	// TODO:注册其他处理器...
	switch docType {
	case doctype.TypeVatInvoice:
		return &vat.Docs{}
	case doctype.TypeTrainTicket:
		return &train.Docs{}
	default:
		logger.Error("Unsupported document type: ", docType)
		return nil
	}
}
