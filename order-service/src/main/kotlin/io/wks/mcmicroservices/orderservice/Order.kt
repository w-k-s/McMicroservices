package io.wks.mcmicroservices.orderservice

import com.fasterxml.jackson.annotation.JsonCreator
import com.fasterxml.jackson.annotation.JsonValue
import org.springframework.data.annotation.CreatedDate
import org.springframework.data.annotation.Id
import org.springframework.data.annotation.LastModifiedDate
import org.springframework.data.annotation.Version
import java.time.OffsetDateTime
import java.time.ZoneOffset
import java.util.*
import javax.validation.Constraint
import javax.validation.ConstraintValidator
import javax.validation.ConstraintValidatorContext
import javax.validation.Payload
import kotlin.reflect.KClass

// TODO: Why does deserialization fail if I annotate the constructor with @JsonCreator, yet work when i create and annotate a factory method
data class OrderId constructor(@JsonValue val value: String) {
    constructor() : this(
        OffsetDateTime.now().withOffsetSameInstant(ZoneOffset.UTC).toInstant().toEpochMilli().toString()
    )
    companion object {
        @JsonCreator
        @JvmStatic
        fun of(value: String) = OrderId(value)
    }
}

data class Toppings constructor(
    @JsonValue
    private val items: SortedSet<String>
): Iterable<String> {

    companion object {
        fun valueOf(toppings: String) = toppings
            .split(",")
            .map { it.trim() }
            .let { Toppings(it.toSortedSet()) }
    }

    constructor(vararg toppings: String) : this(sortedSetOf(*toppings))

    override fun toString() = items.joinToString(", ")
    override fun iterator() = items.iterator()
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

@Constraint(validatedBy = [ToppingsValidator::class])
@Retention(AnnotationRetention.RUNTIME)
annotation class ValidToppings(
    val message: String = "Toppings can not be empty.",
    val groups: Array<KClass<*>> = [],
    val payload: Array<KClass<out Payload>> = []
)

class ToppingsValidator :
    ConstraintValidator<ValidToppings, Toppings> {
    override fun isValid(value: Toppings, context: ConstraintValidatorContext?): Boolean {
        return value.map { it.length }.sum() > 0
    }
}