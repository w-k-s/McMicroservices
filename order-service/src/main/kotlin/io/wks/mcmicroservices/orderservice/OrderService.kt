package io.wks.mcmicroservices.orderservice

import com.fasterxml.jackson.databind.ObjectMapper
import org.springframework.kafka.core.KafkaTemplate
import org.springframework.stereotype.Service
import org.springframework.transaction.annotation.Transactional

@Service
@Transactional
class OrderService(
    private val orderRepository: OrderRepository,
    private val kafkaTemplate: KafkaTemplate<kotlin.String, kotlin.String>,
    private val objectMapper: ObjectMapper
) {


    fun createOrder(orderRequest: OrderRequest): Order {
        return orderRepository.save(
            Order(
                id = OrderId(),
                toppings = orderRequest.toppings,
                status = Order.Status.PREPARING
            ).also {
                // TODO: Use transactional outbox pattern
                kafkaTemplate.send("order_created", objectMapper.writeValueAsString(OrderCreatedEvent(it)))
            }
        )
    }

    fun getOrders(): List<Order> {
        return orderRepository.findAll().toList()
    }
}