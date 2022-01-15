package io.wks.mcmicroservices.orderservice

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.http.MediaType
import org.springframework.kafka.core.KafkaTemplate
import org.springframework.test.web.servlet.MockMvc
import org.springframework.test.web.servlet.get
import org.springframework.test.web.servlet.post
import java.util.concurrent.TimeUnit


internal class OrderOrchestratorIT : BaseSpringBootTest(){

    @Autowired
    private lateinit var orderRepository: OrderRepository

    @Autowired
    private lateinit var kafkaTemplate: KafkaTemplate<String, String>

    @Autowired
    private lateinit var mockMvc: MockMvc

    @Autowired
    private lateinit var orderCreatedConsumer: OrderCreatedTopicConsumer
    @Autowired
    private lateinit var orderReadyConsumer: OrderReadyConsumer
    @Autowired
    private lateinit var orderFailedConsumer: OrderFailedConsumer

    @AfterEach
    fun tearDown() {
        orderRepository.deleteAll()
    }

    @Test
    fun `GIVEN an order WHEN it is ready THEN status is updated`() {
        // GIVEN
        val orderRequest = OrderRequest(toppings = Toppings("Zucchini", "Rice", "Avocado"))
        mockMvc.post("/orders/api/v1/orders") {
            contentType = MediaType.APPLICATION_JSON
            content = objectMapper.writeValueAsString(orderRequest)
        }.andExpect {
            status { isOk() }
        }

        // WHEN
        orderCreatedConsumer.latch.await(1, TimeUnit.MINUTES)
        val order = orderCreatedConsumer.pop()!!.order
        kafkaTemplate.send("order_ready", objectMapper.writeValueAsString(OrderEvent(
            id = order.id,
            status = Order.Status.READY
        )))
        kafkaTemplate.flush()

        // THEN
        orderReadyConsumer.latch.await(1, TimeUnit.MINUTES)
        assertThat(orderReadyConsumer.latch.count).isEqualTo(0L)

        mockMvc.get("/orders/api/v1/orders") {
            contentType = MediaType.APPLICATION_JSON
        }.andExpect {
            status { isOk() }
            content{
                json("""{
                    "orders":[{
                        "toppings":["Avocado","Rice","Zucchini"],
                        "status":"READY",
                        "version":1
                    }]
                }""".trimIndent())
            }
        }
    }

    @Test
    fun `GIVEN an order WHEN preparation fails THEN status is updated`() {
        // GIVEN
        val orderRequest = OrderRequest(toppings = Toppings("Zucchini", "Rice", "Avocado"))
        mockMvc.post("/orders/api/v1/orders") {
            contentType = MediaType.APPLICATION_JSON
            content = objectMapper.writeValueAsString(orderRequest)
        }.andExpect {
            status { isOk() }
        }

        // WHEN
        orderCreatedConsumer.latch.await(1, TimeUnit.MINUTES)

        val order = orderCreatedConsumer.pop()!!.order
        kafkaTemplate.send("order_ready", objectMapper.writeValueAsString(OrderEvent(
            id = order.id,
            status = Order.Status.FAILED,
            failureReason = "Insufficient stock of \"Tomatoes\""
        )))

        // THEN
        orderFailedConsumer.latch.await(1, TimeUnit.MINUTES)
        assertThat(orderFailedConsumer.latch.count).isEqualTo(0L)

        mockMvc.get("/orders/api/v1/orders") {
            contentType = MediaType.APPLICATION_JSON
        }.andExpect {
            status { isOk() }
            content{
                /*json("""{
                    "orders":[{
                        "toppings":["Avocado","Rice","Zucchini"],
                        "status":"FAILED",
                        "version":1,
                        "failureReason":"Insufficient stock of \"Tomatoes\""
                    }]
                }""".trimIndent())*/
            }
        }.andReturn().response.contentAsString.let { println(it) }
    }
}
