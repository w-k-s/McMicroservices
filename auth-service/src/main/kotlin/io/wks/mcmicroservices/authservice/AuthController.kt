package io.wks.mcmicroservices.authservice

import org.springframework.http.ResponseEntity
import org.springframework.validation.annotation.Validated
import org.springframework.web.bind.annotation.*
import org.zalando.problem.spring.web.advice.ProblemHandling
import java.net.URI
import java.util.concurrent.ConcurrentHashMap
import javax.servlet.http.HttpServletRequest
import javax.validation.constraints.Email
import javax.validation.constraints.NotBlank

data class AuthRequest(
    @field:Email(message = "validation.username.required")
    val username: String,
    @field:NotBlank(message = "validation.password.required")
    val password: String
)

data class TokenResponse(val token: String)

@RestController
@RequestMapping("/auth/api/v1")
class AuthController {

    val users = ConcurrentHashMap<String, String>()

    @PostMapping("/login")
    fun login(@Validated @RequestBody authRequest: AuthRequest,
              request: HttpServletRequest): ResponseEntity<*> {
        return when (users[authRequest.username] == authRequest.password) {
            true -> ResponseEntity.ok(TokenResponse("token"))
            false -> throw InvalidCredentialsProblem(instance = URI.create(request.requestURI))
        }
    }

    @PostMapping("/register")
    fun register(@Validated @RequestBody request: AuthRequest): ResponseEntity<*> {
        users[request.username] = request.password
        return ResponseEntity.ok(Unit)
    }
}

@ControllerAdvice
class ExceptionHandler : ProblemHandling