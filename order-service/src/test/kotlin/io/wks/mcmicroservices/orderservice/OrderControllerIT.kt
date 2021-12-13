package io.wks.mcmicroservices.orderservice

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.AfterEach
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.http.HttpEntity
import org.springframework.http.HttpHeaders
import org.springframework.http.MediaType
import org.springframework.test.web.servlet.MockMvc
import org.springframework.test.web.servlet.post
import org.zalando.problem.Status
import org.zalando.problem.violations.ConstraintViolationProblem
import org.zalando.problem.violations.Violation
import java.net.URI
import java.time.OffsetDateTime

internal class OrderControllerIT : BaseSpringBootTest() {

    @Autowired
    private lateinit var orderRepository: OrderRepository

    @Autowired
    private lateinit var controller: OrderController

    @Autowired
    private lateinit var mockMvc: MockMvc

    @AfterEach
    fun tearDown() {
        orderRepository.deleteAll()
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
                    "violations":[{"field":"items","message":"Must be valid"}]
                }""".trimIndent()
                )
            }
        }
    }

    @Test
    fun `GIVEN toppings WHEN creating order THEN order is created and can be retrieved`() {
        // GIVEN
        val orderRequest = OrderRequest(toppings = Toppings("Zucchini", "Rice", "Avocado"))

        // WHEN
        val orderResponse = restTemplate.postForEntity(
            "http://localhost:$port/orders/api/v1/orders",
            orderRequest,
            String::class.java
        )

        // THEN
        assertThat(orderResponse.statusCodeValue).isEqualTo(200)

        val ordersResponse = restTemplate.getForEntity(
            "http://localhost:$port/orders/api/v1/orders",
            OrdersResponse::class.java
        )

        assertThat(ordersResponse.body)
            .usingRecursiveComparison()
            .ignoringFieldsOfTypes(OrderId::class.java, OffsetDateTime::class.java)
            .isEqualTo(
                OrdersResponse(
                    orders = listOf(
                        Order(
                            id = OrderId(),
                            toppings = Toppings(
                                "Zucchini",
                                "Rice",
                                "Avocado"
                            ),
                            status = Order.Status.PREPARING,
                            version = 1
                        )
                    )
                )
            )
    }
}