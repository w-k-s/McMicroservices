package io.wks.mcmicroservices.orderservice

import org.springframework.core.convert.converter.Converter
import org.springframework.data.convert.ReadingConverter
import org.springframework.data.convert.WritingConverter
import org.springframework.data.repository.CrudRepository
import org.springframework.stereotype.Repository
import org.springframework.data.jdbc.repository.query.Modifying
import org.springframework.data.jdbc.repository.query.Query
import org.springframework.data.repository.query.Param


@Repository
interface OrderRepository : CrudRepository<Order, OrderId>{

    @Modifying
    @Query("UPDATE \"order\".\"order\" SET status = 'READY' WHERE id = :id")
    fun setOrderReady(
        @Param("id") id: OrderId
    ): Boolean

    @Modifying
    @Query("UPDATE \"order\".\"order\" SET status = 'FAILED', failure_reason = :reason WHERE id = :id")
    fun setOrderFailed(
        @Param("id") id: OrderId,
        @Param("reason") reason: String
    ): Boolean

}

@WritingConverter
class ToppingsToStringConverter : Converter<Toppings, String> {
    override fun convert(source: Toppings) = source.toString()
}

@ReadingConverter
class StringToToppingsConverter : Converter<String, Toppings> {
    override fun convert(source: String) = source.let { Toppings.valueOf(it) }
}

@WritingConverter
class OrderIdToStringConverter : Converter<OrderId, String> {
    override fun convert(source: OrderId) = source.value
}

@ReadingConverter
class StringToOrderIdConverter : Converter<String, OrderId> {
    override fun convert(source: String) = OrderId(source)
}