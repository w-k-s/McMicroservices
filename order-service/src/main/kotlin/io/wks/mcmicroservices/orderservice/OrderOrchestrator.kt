package io.wks.mcmicroservices.orderservice

import com.fasterxml.jackson.databind.ObjectMapper
import org.springframework.kafka.core.KafkaTemplate
import org.springframework.stereotype.Service
import org.springframework.kafka.annotation.KafkaListener


@Service
class OrderOrchestrator(
    private val orderRepository: OrderRepository,
    private val kafkaTemplate: KafkaTemplate<String, String>,
    private val objectMapper: ObjectMapper
) {

    fun newOrder(order: Order): Order {
        return orderRepository.save(order).also {
            // TODO: Use transactional outbox pattern
            kafkaTemplate.send("order_created", objectMapper.writeValueAsString(OrderCreatedEvent(it)))
        }
    }

    @KafkaListener(
        topics = ["order_ready"],
        groupId = "group_id"
    )
    fun orderReady(message: String) {
        val orderEvent = objectMapper.readValue(message, OrderEvent::class.java)
        orderRepository.setOrderReady(orderEvent.id)
    }

    @KafkaListener(
        topics = ["order_failed"],
        groupId = "group_id"
    )
    fun orderFailed(message: String) {
        val orderEvent = objectMapper.readValue(message, OrderEvent::class.java)
        orderRepository.setOrderFailed(
            orderEvent.id,
            orderEvent.failureReason ?: "unknown reason"
        )
    }
}