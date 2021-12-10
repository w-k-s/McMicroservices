package io.wks.mcmicroservices.authservice.config

import com.fasterxml.jackson.databind.ObjectMapper
import com.fasterxml.jackson.module.kotlin.KotlinModule
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.zalando.problem.ProblemModule
import org.zalando.problem.violations.ConstraintViolationProblemModule

@Configuration
class ObjectMapperConfiguration {

    @Bean
    fun objectMapper()
    = ObjectMapper()
        .registerModule(ProblemModule())
        .registerModule(ConstraintViolationProblemModule())
        .registerModule(KotlinModule.Builder().build())
}