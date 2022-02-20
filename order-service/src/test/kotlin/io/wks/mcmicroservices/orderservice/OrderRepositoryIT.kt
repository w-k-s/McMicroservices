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
        assertThat(underTest.findById(order.id).get())
            .usingRecursiveComparison()
            .ignoringFields("createdAt", "updatedAt")
            .isEqualTo(order.copy(version = 1))
    }

    @Test
    fun `GIVEN an order WHEN it is ready THEN its status is updated`(){
        // GIVEN
        val order = underTest.save(Order(
            OrderId(),
            Toppings("Cheese", "Banana"),
            Order.Status.PREPARING,
        ))

        // WHEN
        underTest.setOrderReady(order.id)

        // THEN
        val expected = order.copy(
            status = Order.Status.READY
        )
        assertThat(underTest.findById(order.id).get())
            .usingRecursiveComparison()
            .ignoringFields("createdAt", "updatedAt")
            .isEqualTo(expected)
    }

    @Test
    fun `GIVEN an order WHEN it failed THEN its status is updated`(){
        // GIVEN
        val order = underTest.save(Order(
            OrderId(),
            Toppings("Cheese", "Banana"),
            Order.Status.PREPARING,
        ))

        // WHEN
        underTest.setOrderFailed(order.id, "Insufficient ingredient \"TOMATO\"")

        // THEN
        val expected = order.copy(
            status = Order.Status.FAILED,
            failureReason = "Insufficient ingredient \"TOMATO\""
        )
        assertThat(underTest.findById(order.id).get())
            .usingRecursiveComparison()
            .ignoringFields("createdAt", "updatedAt")
            .isEqualTo(expected)
    }
}