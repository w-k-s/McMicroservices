package kitchen

type OrderStatus string
const (
	OrderStatusPreparing OrderStatus = "PREPARING"
	OrderStatusReady     OrderStatus = "READY"
	OrderStatusFailed    OrderStatus = "FAILED"
)