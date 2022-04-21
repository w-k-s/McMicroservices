package app

const (
	TopicCreateOrder string = "order_created"
	TopicOrderReady  string = "order_ready"
	TopicOrderFailed string = "order_failed"

	TopicInventoryDelivery string = "inventory_delivery"
)

func (a *App) configureEventHandlers() {
	a.consumer.AddTopicEventHandler(TopicInventoryDelivery, a.ReceiveInventory)
	a.consumer.AddTopicEventHandler(TopicCreateOrder, a.HandleOrderMessage)
}
