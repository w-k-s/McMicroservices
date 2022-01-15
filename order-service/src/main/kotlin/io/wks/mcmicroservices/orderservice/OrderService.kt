package io.wks.mcmicroservices.orderservice

import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional

@Service
@Transactional
class OrderService(
    private val orderRepository: OrderRepository,
    private val orderOrchestrator: OrderOrchestrator
) {


    fun createOrder(orderRequest: OrderRequest): Order {
        return orderOrchestrator.newOrder(Order(
            id = OrderId(),
            toppings = orderRequest.toppings,
            status = Order.Status.PREPARING
        ))
    }

    fun getOrders(): List<Order> {
        return orderRepository.findAll().toList()
    }
}