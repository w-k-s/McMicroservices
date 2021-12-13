package io.wks.mcmicroservices.orderservice

import io.konform.validation.ValidationResult
import org.zalando.problem.Status
import org.zalando.problem.violations.ConstraintViolationProblem
import org.zalando.problem.violations.Violation

inline fun <reified T> ValidationResult<T>.raiseProblem() {
    if (this.errors.isNotEmpty()) {
        throw ConstraintViolationProblem(
            Status.BAD_REQUEST,
            this.errors.map {
                Violation(
                    it.dataPath.substring(1),
                    it.message
                )
            }
        )
    }
}