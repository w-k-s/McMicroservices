package io.wks.mcmicroservices.orderservice

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired

class OrderRepositoryIT : BaseSpringBootTest(){

    @Autowired
    lateinit var underTest : OrderRepository

    @Test
    fun `GIVEN an order WHEN it is saved THEN it can be retrieved`(){
        // GIVEN
        val order = Order(
            OrderId(),
            Toppings("Cheese", "Banana"),
            Order.Status.PREPARING,
        )

        // WHEN
        underTest.save(order)

        // THEN
        assertThat(underTest.findById(order.id).get()).isEqualTo(order.copy(version = 1))
    }
}