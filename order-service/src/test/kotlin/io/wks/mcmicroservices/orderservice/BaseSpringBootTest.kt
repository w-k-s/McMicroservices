package io.wks.mcmicroservices.orderservice

import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.autoconfigure.jdbc.AutoConfigureTestDatabase
import org.springframework.boot.test.autoconfigure.web.servlet.AutoConfigureMockMvc
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.boot.test.util.TestPropertyValues
import org.springframework.boot.test.web.client.TestRestTemplate
import org.springframework.boot.web.server.LocalServerPort
import org.springframework.context.ApplicationContextInitializer
import org.springframework.context.ConfigurableApplicationContext
import org.springframework.test.context.ActiveProfiles
import org.springframework.test.context.ContextConfiguration
import org.testcontainers.containers.PostgreSQLContainer
import org.testcontainers.junit.jupiter.Testcontainers

@ActiveProfiles("test")
@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
@AutoConfigureTestDatabase(replace = AutoConfigureTestDatabase.Replace.NONE)
@AutoConfigureMockMvc
@ContextConfiguration(initializers = [BaseSpringBootTest.DockerPostgreDataSourceInitializer::class])
@Testcontainers
class BaseSpringBootTest {

    @Autowired
    protected lateinit var restTemplate: TestRestTemplate

    @LocalServerPort
    protected var port: Int = 0

    companion object {
        val postgreSQLContainer = PostgreSQLContainer<Nothing>("postgres:9.4")
    }

    class DockerPostgreDataSourceInitializer : ApplicationContextInitializer<ConfigurableApplicationContext> {

        override fun initialize(applicationContext: ConfigurableApplicationContext) {
            postgreSQLContainer.start()
            TestPropertyValues.of(
                "spring.datasource.url=" + postgreSQLContainer.jdbcUrl,
                "spring.datasource.username=" + postgreSQLContainer.username,
                "spring.datasource.password=" + postgreSQLContainer.password
            ).applyTo(applicationContext.environment)
        }
    }
}