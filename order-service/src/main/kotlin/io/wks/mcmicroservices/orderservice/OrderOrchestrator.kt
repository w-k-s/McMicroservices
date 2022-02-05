package io.wks.mcmicroservices.orderservice

import com.fasterxml.jackson.databind.ObjectMapper
import org.slf4j.LoggerFactory
import org.springframework.kafka.core.KafkaTemplate
import org.springframework.stereotype.Service
import org.springframework.kafka.annotation.KafkaListener


@Service
class OrderOrchestrator(
    private val orderRepository: OrderRepository,
    private val kafkaTemplate: KafkaTemplate<String, String>,
    private val objectMapper: ObjectMapper
) {
    companion object{
        val LOGGER = LoggerFactory.getLogger(OrderOrchestrator::class.java)
    }

    fun newOrder(order: Order): Order {
        return orderRepository.save(order).also {
            // TODO: Use transactional outbox pattern
            kafkaTemplate.send("order_created", objectMapper.writeValueAsString(OrderCreatedEvent(it)))
        }
    }

    @KafkaListener(
        topics = ["order_ready"],
        groupId = "\${spring.kafka.consumer.group-id}"
    )
    fun orderReady(message: String) {
        LOGGER.info("Message received in topic \"order_ready\": $message")
        val orderEvent = objectMapper.readValue(message, OrderEvent::class.java)
        orderRepository.setOrderReady(orderEvent.id).also {
            LOGGER.info("Order id '${orderEvent.id}' updated to ready: $it")
        }
    }

    @KafkaListener(
        topics = ["order_failed"],
        groupId = "\${spring.kafka.consumer.group-id}"
    )
    fun orderFailed(message: String) {
        LOGGER.info("Message received in topic \"order_failed\": $message")
        val orderEvent = objectMapper.readValue(message, OrderEvent::class.java)
        orderRepository.setOrderFailed(
            id = orderEvent.id,
            reason = orderEvent.failureReason ?: "unknown reason"
        ).also {
            LOGGER.info("Order id '${orderEvent.id}' updated to failed: $it")
        }
    }
}