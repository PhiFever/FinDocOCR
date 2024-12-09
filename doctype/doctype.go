package doctype

// DocumentType 用于标识不同类型的文档
type DocumentType string

// 定义票据类型常量
const (
	TypeVatInvoice          = "vat_invoice"
	TypeTaxiReceipt         = "taxi_receipt"
	TypeTrainTicket         = "train_ticket"
	TypeQuotaInvoice        = "quota_invoice"
	TypeAirTicket           = "air_ticket"
	TypeRollNormalInvoice   = "roll_normal_invoice"
	TypePrintedInvoice      = "printed_invoice"
	TypePrintedElecInvoice  = "printed_elec_invoice"
	TypeBusTicket           = "bus_ticket"
	TypeTollInvoice         = "toll_invoice"
	TypeFerryTicket         = "ferry_ticket"
	TypeMotorVehicleInvoice = "motor_vehicle_invoice"
	TypeUsedVehicleInvoice  = "used_vehicle_invoice"
	TypeTaxiOnlineTicket    = "taxi_online_ticket"
	TypeLimitInvoice        = "limit_invoice"
	TypeShoppingReceipt     = "shopping_receipt"
	TypePosInvoice          = "pos_invoice"
	TypeOthers              = "others"
)

type Document interface {
	AmendData()
}

// DocumentCollection 定义文档集合接口
type DocumentCollection interface {
	Add(doc Document)
	SaveToFile() error
}
