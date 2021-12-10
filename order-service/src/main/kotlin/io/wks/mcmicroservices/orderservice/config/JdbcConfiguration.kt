package io.wks.mcmicroservices.orderservice.config

import io.wks.mcmicroservices.orderservice.*
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.data.jdbc.repository.config.AbstractJdbcConfiguration
import org.springframework.data.relational.core.mapping.NamingStrategy

@Configuration
class JdbcConfiguration : AbstractJdbcConfiguration() {

    override fun userConverters() = mutableListOf(
        ToppingsToStringConverter(),
        StringToToppingsConverter(),
        OrderIdToStringConverter(),
        StringToOrderIdConverter()
    )
}

@Configuration
class DatabaseConfiguration{

    @Bean
    fun namingStrategy(): NamingStrategy = object: NamingStrategy{
        override fun getSchema() = "order"
    }
}
