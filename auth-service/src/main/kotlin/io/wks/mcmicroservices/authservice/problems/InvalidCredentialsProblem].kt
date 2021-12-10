package io.wks.mcmicroservices.authservice

import org.zalando.problem.AbstractThrowableProblem
import org.zalando.problem.Status
import java.net.URI

class InvalidCredentialsProblem(detail: String? = null, instance: URI? = null) :
    AbstractThrowableProblem(
        TYPE,
        "Invalid username or password",
        Status.UNAUTHORIZED,
        detail,
        instance
    ) {
    companion object {
        val TYPE: URI = URI.create("api/v1/problems/INVALID_CREDENTIALS")
    }

    override fun getCause() = null
}