package io.wks.mcmicroservices.orderservice

import com.fasterxml.jackson.module.kotlin.jacksonObjectMapper
import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

class OrderRequestTest {

    @Test
    fun `GIVEN an order request WHEN it is serialized to json THEN toppings is array of strings`(){
        // GIVEN
        val request = OrderRequest(toppings = Toppings("Cheese","Ketchup"))

        // WHEN
        val json = jacksonObjectMapper().writeValueAsString(request)

        // THEN
        assertThat(json).isEqualTo("{\"toppings\":[\"Cheese\",\"Ketchup\"]}")
    }

    @Test
    fun `GIVEN an order request json WHEN it is deserialized to json THEN matching object is created`(){
        // GIVEN
        val originalOrderRequest = OrderRequest(toppings = Toppings("Cheese","Ketchup"))
        val json = jacksonObjectMapper().writeValueAsString(originalOrderRequest)

        // WHEN
        val deserializedOrderRequest = jacksonObjectMapper().readValue(json, OrderRequest::class.java)

        // THEN
        assertThat(originalOrderRequest).isEqualTo(deserializedOrderRequest)
    }
}