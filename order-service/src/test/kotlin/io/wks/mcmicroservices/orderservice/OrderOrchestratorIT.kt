package io.wks.mcmicroservices.orderservice

import io.github.artsok.RepeatedIfExceptionsTest
import org.awaitility.Awaitility.await
import org.junit.jupiter.api.AfterEach
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.http.MediaType
import org.springframework.kafka.core.KafkaTemplate
import org.springframework.test.web.servlet.MockMvc
import org.springframework.test.web.servlet.get
import org.springframework.test.web.servlet.post
import java.util.concurrent.TimeUnit


internal class OrderOrchestratorIT : BaseSpringBootTest() {

    @Autowired
    private lateinit var orderRepository: OrderRepository

    @Autowired
    private lateinit var kafkaTemplate: KafkaTemplate<String, String>

    @Autowired
    private lateinit var mockMvc: MockMvc

    @Autowired
    private lateinit var orderCreatedConsumer: OrderCreatedTopicConsumer

    @AfterEach
    fun tearDown() {
        orderRepository.deleteAll()
    }

    @RepeatedIfExceptionsTest(repeats = 3)
    fun `GIVEN an order WHEN it is ready THEN status is updated`() {
        // GIVEN
        mockMvc.post("/orders/api/v1/orders") {
            contentType = MediaType.APPLICATION_JSON
            content = """{
                "toppings": ["Zucchini","Rice","Avocado"]
            }"""
        }.andExpect {
            status { isOk() }
        }

        // WHEN
        orderCreatedConsumer.latch.await(1, TimeUnit.MINUTES)
        val order = orderCreatedConsumer.pop()!!.order
        kafkaTemplate.send(
            "order_ready", objectMapper.writeValueAsString(
                OrderEvent(
                    id = order.id,
                    status = Order.Status.READY
                )
            )
        ).get()
        kafkaTemplate.flush()

        await()
            .atMost(1, TimeUnit.MINUTES)
            .pollDelay(10, TimeUnit.SECONDS)
            .pollInterval(10, TimeUnit.SECONDS)
            .untilAsserted {
                mockMvc.get("/orders/api/v1/orders") {
                    contentType = MediaType.APPLICATION_JSON
                }.andExpect {
                    status { isOk() }
                    content {
                        json(
                            """{
                    "orders":[{
                        "toppings":["Avocado","Rice","Zucchini"],
                        "status":"READY",
                        "version":1
                    }]
                }""".trimIndent()
                        )
                    }
                }
            }
    }

    @RepeatedIfExceptionsTest(repeats = 3)
    fun `GIVEN an order WHEN preparation fails THEN status is updated`() {
        // GIVEN
        mockMvc.post("/orders/api/v1/orders") {
            contentType = MediaType.APPLICATION_JSON
            content = """{
                "toppings": ["Zucchini","Rice","Avocado"]
            }"""
        }.andExpect {
            status { isOk() }
        }

        // WHEN
        orderCreatedConsumer.latch.await(1, TimeUnit.MINUTES)

        val order = orderCreatedConsumer.pop()!!.order
        kafkaTemplate.send(
            "order_failed", objectMapper.writeValueAsString(
                OrderEvent(
                    id = order.id,
                    status = Order.Status.FAILED,
                    failureReason = "Insufficient stock of \"Tomatoes\""
                )
            )
        ).get()
        kafkaTemplate.flush()

        // THEN
        await()
            .atMost(1, TimeUnit.MINUTES)
            .pollDelay(10, TimeUnit.SECONDS)
            .pollInterval(10, TimeUnit.SECONDS)
            .untilAsserted {
                mockMvc.get("/orders/api/v1/orders") {
                    contentType = MediaType.APPLICATION_JSON
                }.andExpect {
                    status { isOk() }
                    content {
                        json(
                            """{
                    "orders":[{
                        "toppings":["Avocado","Rice","Zucchini"],
                        "status":"FAILED",
                        "version":1,
                        "failureReason":"Insufficient stock of \"Tomatoes\""
                    }]
                }"""
                        )
                    }
                }
            }
    }
}
