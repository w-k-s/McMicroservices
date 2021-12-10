package io.wks.mcmicroservices.orderservice

import org.springframework.data.annotation.CreatedDate
import org.springframework.data.annotation.Id
import org.springframework.data.annotation.LastModifiedDate
import org.springframework.data.annotation.Version
import org.springframework.data.relational.core.mapping.Table
import java.time.OffsetDateTime
import java.time.ZoneOffset
import java.util.*

data class OrderId(val value: String){
    constructor() : this(OffsetDateTime.now().withOffsetSameInstant(ZoneOffset.UTC).toInstant().toEpochMilli().toString())
}

data class Toppings(private val items: SortedSet<String>) {
    companion object {
        fun valueOf(toppings: String) = toppings
            .split(",")
            .map { it.trim() }
            .let { Toppings(it.toSortedSet()) }
    }
    constructor(vararg toppings: String) : this(sortedSetOf(*toppings))
    override fun toString() = items.joinToString(", ")
}

data class Order(
    @Id
    val id: OrderId,
    val toppings: Toppings,
    val status: Status,
    @CreatedDate
    val createdAt: OffsetDateTime = OffsetDateTime.now().withOffsetSameInstant(ZoneOffset.UTC),
    @LastModifiedDate
    val updatedAt: OffsetDateTime? = null,
    @Version
    val version: Int = 0,
    val failureReason: String? = null,
) {
    enum class Status {
        PREPARING,
        READY,
        FAILED
    }
}