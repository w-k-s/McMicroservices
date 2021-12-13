package io.wks.mcmicroservices.orderservice

import io.konform.validation.Validation
import io.konform.validation.jsonschema.minItems
import org.springframework.http.ResponseEntity
import org.springframework.http.converter.HttpMessageNotReadableException
import org.springframework.web.bind.annotation.*
import org.springframework.web.context.request.NativeWebRequest
import org.zalando.problem.Problem
import org.zalando.problem.spring.web.advice.ProblemHandling
import org.zalando.problem.violations.ConstraintViolationProblem
import javax.validation.ConstraintViolationException

data class OrderRequest(val toppings: Toppings) {
    init {
        Validation<OrderRequest>{
            OrderRequest::toppings {
                minItems(1) hint "Toppings are required"
            }
        }(this).raiseProblem()
    }

    class Builder(
        val toppings: Toppings? = null
    ) {
        init {
            Validation<Builder>{
                Builder::toppings required {}
            }(this).raiseProblem()
        }

        fun build() = OrderRequest(requireNotNull(toppings))
    }
}

data class OrdersResponse(
    val orders: List<Order>
)

@RestController
@RequestMapping("orders/api/v1/orders")
class OrderController(private val orderService: OrderService) {

    @PostMapping
    fun createOrder(@RequestBody request: OrderRequest.Builder): ResponseEntity<*> {
        return ResponseEntity.ok(orderService.createOrder(request.build()))
    }

    @GetMapping
    fun getOrders(): ResponseEntity<*> {
        return ResponseEntity.ok(OrdersResponse(orderService.getOrders()))
    }
}

@ControllerAdvice
class ExceptionHandler : ProblemHandling{
    override fun handleMessageNotReadableException(
        exception: HttpMessageNotReadableException,
        request: NativeWebRequest
    ): ResponseEntity<Problem> {
        return when(val cause = exception.mostSpecificCause){
            is ConstraintViolationProblem -> create(cause, cause, request)
            else -> super.handleMessageNotReadableException(exception, request)
        }
    }
}