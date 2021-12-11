package io.wks.mcmicroservices.orderservice

import org.springframework.http.ResponseEntity
import org.springframework.web.bind.annotation.*
import org.zalando.problem.spring.web.advice.ProblemHandling
import javax.validation.*

data class OrderRequest(
    /*
     * TODO: I'd prefer this validation to go in the constructor.
     * When I tried, the Problem library was throwing a jackson error (failed to deserialize request json)
     * rather than a validation problem.
     */
    @field:ValidToppings(message = "validation.toppings.required")
    val toppings: Toppings
)

data class OrdersResponse(
    val orders: List<Order>
)

@RestController
@RequestMapping("orders/api/v1/orders")
class OrderController(private val orderService: OrderService) {

    @PostMapping
    fun createOrder(@Valid @RequestBody request: OrderRequest): ResponseEntity<*> {
        return ResponseEntity.ok(orderService.createOrder(request))
    }

    @GetMapping
    fun getOrders(): ResponseEntity<*> {
        return ResponseEntity.ok(OrdersResponse(orderService.getOrders()))
    }
}

@ControllerAdvice
class ExceptionHandler : ProblemHandling