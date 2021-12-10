package io.wks.mcmicroservices.authservice

import org.springframework.boot.test.web.client.TestRestTemplate
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration
import org.springframework.context.annotation.Profile
import org.springframework.http.client.HttpComponentsClientHttpRequestFactory

/*@Configuration
@Profile("test")
class TestConfiguration {

    @Bean
    fun restTemplate(template: TestRestTemplate): TestRestTemplate{
        return template.apply {
            this.restTemplate.requestFactory = HttpComponentsClientHttpRequestFactory()
        }
    }
}*/