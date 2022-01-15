package io.wks.mcmicroservices.orderservice

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.http.MediaType
import org.springframework.test.web.servlet.MockMvc
import org.springframework.test.web.servlet.get
import org.springframework.test.web.servlet.post
import java.time.OffsetDateTime
import java.util.concurrent.TimeUnit

internal class OrderControllerIT : BaseSpringBootTest() {

    @Autowired
    private lateinit var orderRepository: OrderRepository

    @Autowired
    private lateinit var controller: OrderController

    @Autowired
    private lateinit var mockMvc: MockMvc

    @Autowired
    private lateinit var orderCreatedConsumer: OrderCreatedTopicConsumer

    @AfterEach
    fun tearDown() {
        orderRepository.deleteAll()
        orderCreatedConsumer.clear()
    }

    @Test
    fun contextLoads() {
        assertThat(controller).isNotNull
    }

    @Test
    fun `GIVEN empty list of toppings WHEN creating order THEN bad request is returned`() {
        // GIVEN
        mockMvc.post("/orders/api/v1/orders") {
            contentType = MediaType.APPLICATION_JSON
            content = "{\"toppings\":[]}"
        }.andExpect {
            status { isBadRequest() }
            header {
                content {
                    contentType("application/problem+json;charset=UTF-8")
                }
            }
            content {
                json(
                    """{
                    "title":"Constraint Violation",
                    "status":400,
                    "type":"https://zalando.github.io/problem/constraint-violation",
                    "violations":[{"field":"toppings","message":"Toppings are required"}]
                }""".trimIndent()
                )
            }
        }
    }

    @Test
    fun `GIVEN list of empty toppings WHEN creating order THEN bad request is returned`() {
        // GIVEN
        mockMvc.post("/orders/api/v1/orders") {
            contentType = MediaType.APPLICATION_JSON
            content = "{\"toppings\":[\"\",\"\"]}"
        }.andExpect {
            status { isBadRequest() }
            header {
                content {
                    contentType("application/problem+json;charset=UTF-8")
                }
            }
            content {
                json(
                    """{
                    "title":"Constraint Violation",
                    "status":400,
                    "type":"https://zalando.github.io/problem/constraint-violation",
                    "violations":[{"field":"items","message":"Topping can not be blank"}]
                }""".trimIndent()
                )
            }
        }
    }

    @Test
    fun `GIVEN toppings WHEN creating order THEN order is created and can be retrieved`() {
        // GIVEN
        val orderRequest = OrderRequest(toppings = Toppings("Zucchini", "Rice", "Avocado"))
        mockMvc.post("/orders/api/v1/orders") {
            contentType = MediaType.APPLICATION_JSON
            content = objectMapper.writeValueAsString(orderRequest)
        }.andExpect {
            status { isOk() }
        }

        mockMvc.get("/orders/api/v1/orders") {
            contentType = MediaType.APPLICATION_JSON
        }.andExpect {
            status { isOk() }
            content{
                json("""{
                    "orders":[{
                        "toppings":["Avocado","Rice","Zucchini"],
                        "status":"PREPARING",
                        "version":1
                    }]
                }""".trimIndent())
            }
        }

        orderCreatedConsumer.latch.await(1, TimeUnit.MINUTES)
        val order = orderCreatedConsumer.pop()?.order
        assertThat(order).isNotNull
        assertThat(order!!).usingRecursiveComparison()
            .ignoringFieldsOfTypes(OrderId::class.java, OffsetDateTime::class.java)
            .isEqualTo(Order(
                id = OrderId(),
                toppings = orderRequest.toppings,
                status = Order.Status.PREPARING,
                version = 1
            ))
    }
}