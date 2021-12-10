package io.wks.mcmicroservices.authservice

import org.zalando.problem.AbstractThrowableProblem
import org.zalando.problem.Status
import org.zalando.problem.StatusType
import java.net.URI


enum class ProblemType(
    val type: URI,
    private val statusCode: Int,
    private val reasonPhrase: String
) : StatusType {
    BAD_REQUEST(URI.create("api/v1/problems/BAD_REQUEST"), 400, "Bad Request");

    override fun getStatusCode() = this.statusCode
    override fun getReasonPhrase() = this.reasonPhrase
}

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

/*
fun BindingResult.toProblem(instance: String? = null): Problem {
    require(this.hasErrors())

    return Problem.builder()
        .withStatus(ProblemType.BAD_REQUEST)
        .withType(ProblemType.BAD_REQUEST.type)
        .also { problem ->

            instance?.let {
                problem.withInstance(URI.create(it))
            }

            this.fieldErrors.forEach { fieldError ->
                problem.with(fieldError.field, fieldError.defaultMessage)
            }

            this.globalErrors
                .takeIf { it.isNotEmpty() }
                ?.map { global -> global.objectName }
                ?.joinToString(", ")
                ?.let { problem.withDetail(it) }
        }.build()
}
*/
