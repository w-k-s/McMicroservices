package io.wks.mcmicroservices.orderservice

import com.fasterxml.jackson.databind.ObjectMapper
import org.apache.kafka.clients.consumer.ConsumerRecord
import org.springframework.context.annotation.Profile
import org.springframework.kafka.annotation.KafkaListener
import org.springframework.stereotype.Component
import java.util.*
import java.util.concurrent.CountDownLatch

@Profile("test")
@Component
class OrderCreatedTopicConsumer(private val objectMapper: ObjectMapper) {

    private val stack = Stack<String>()
    val latch = CountDownLatch(1)

    @KafkaListener(topics = ["order_created"])
    fun receive(consumerRecord: ConsumerRecord<String, String>) {
        stack.push(consumerRecord.value())
        latch.countDown()
    }

    fun pop(): OrderCreatedEvent? {
        return when (this.stack.isEmpty()) {
            true -> null
            false -> objectMapper.readValue(this.stack.pop(), OrderCreatedEvent::class.java)
        }
    }

    fun clear(){
        this.stack.clear()
    }
}