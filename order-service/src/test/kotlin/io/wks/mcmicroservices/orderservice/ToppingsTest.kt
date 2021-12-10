package io.wks.mcmicroservices.orderservice

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test

internal class ToppingsTest{

    @Test
    fun `GIVEN toppings WHEN it is converted to string and back THEN original list is restored`(){
        // GIVEN
        val toppings = Toppings("Zinc","Aluminium","Cheese")

        // THEN
        assertThat(Toppings.valueOf(toppings.toString())).isEqualTo(toppings)
    }

    @Test
    fun `GIVEN toppings WHEN it is converted to string THEN toppings are sorted in descending order`(){
        // GIVEN
        val toppings = Toppings("Zinc","Aluminium","Cheese")

        // THEN
        assertThat(toppings.toString()).isEqualTo("Aluminium, Cheese, Zinc")
    }
}