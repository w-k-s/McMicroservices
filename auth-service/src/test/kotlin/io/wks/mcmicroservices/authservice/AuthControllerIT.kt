package io.wks.mcmicroservices.authservice

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.context.SpringBootTest
import org.springframework.boot.test.web.client.TestRestTemplate
import org.springframework.boot.web.server.LocalServerPort
import org.springframework.http.MediaType
import org.zalando.problem.DefaultProblem
import org.zalando.problem.Status
import org.zalando.problem.violations.ConstraintViolationProblem
import org.zalando.problem.violations.Violation
import java.net.URI


@SpringBootTest(webEnvironment = SpringBootTest.WebEnvironment.RANDOM_PORT)
internal class AuthControllerIT {

    @Autowired
    private lateinit var controller: AuthController

    @LocalServerPort
    private var port: Int = 0

    @Autowired
    private lateinit var restTemplate: TestRestTemplate

    @Test
    fun contextLoads() {
        assertThat(controller).isNotNull
    }

    @Test
    fun `GIVEN a user registers WHEN the user logs in THEN the user receives a token`() {
        // GIVEN
        val registerRequest = AuthRequest(username = "jack.torrance@theoverlook.com", "123456")
        val registerResponse = restTemplate.postForEntity("http://localhost:$port/auth/api/v1/register", registerRequest, Unit::class.java)

        assertThat(registerResponse.statusCodeValue).isEqualTo(200)

        // WHEN
        val loginResponse = restTemplate.postForEntity("http://localhost:$port/auth/api/v1/login", registerRequest, TokenResponse::class.java)

        // THEN
        assertThat(loginResponse.statusCodeValue).isEqualTo(200)
        assertThat(loginResponse.body!!.token).isNotBlank
    }

    @Test
    fun `GIVEN a user registers WHEN the email is invalid THEN bad request is returned`() {
        // GIVEN
        val registerRequest = AuthRequest(username = "jack.torrance", "123456")
        val registerResponse = restTemplate.postForEntity("http://localhost:$port/auth/api/v1/register", registerRequest, ConstraintViolationProblem::class.java)

        // THEN
        assertThat(registerResponse.statusCodeValue).isEqualTo(400)
        assertThat(registerResponse.headers.contentType).isEqualTo(MediaType.APPLICATION_PROBLEM_JSON)
        assertThat(registerResponse.body?.title).isEqualTo("Constraint Violation")
        assertThat(registerResponse.body?.status).isEqualTo(Status.BAD_REQUEST)
        assertThat(registerResponse.body?.type).isEqualTo(URI.create("https://zalando.github.io/problem/constraint-violation"))
        assertThat(registerResponse.body?.violations?.first()).usingRecursiveComparison().isEqualTo(Violation("username","validation.username.required"))
    }

    @Test
    fun `GIVEN a user logs in WHEN the password is blank THEN bad request is returned`() {
        // GIVEN
        val loginRequest = AuthRequest(username = "jack.torrance", "")
        val loginResponse = restTemplate.postForEntity("http://localhost:$port/auth/api/v1/login", loginRequest, ConstraintViolationProblem::class.java)

        // THEN
        assertThat(loginResponse.statusCodeValue).isEqualTo(400)
        assertThat(loginResponse.headers.contentType).isEqualTo(MediaType.APPLICATION_PROBLEM_JSON)
        assertThat(loginResponse.body?.title).isEqualTo("Constraint Violation")
        assertThat(loginResponse.body?.status).isEqualTo(Status.BAD_REQUEST)
        assertThat(loginResponse.body?.type).isEqualTo(URI.create("https://zalando.github.io/problem/constraint-violation"))
        assertThat(loginResponse.body?.violations?.first()).usingRecursiveComparison().isEqualTo(Violation("password","validation.password.required"))
    }

    @Test
    fun `GIVEN a registered user WHEN the user logs in with incorrect password THEN unauthorized is returned`() {
        // GIVEN
        val registerRequest = AuthRequest(username = "jack.torrance@theoverlook.com", "123456")
        val registerResponse = restTemplate.postForEntity("http://localhost:$port/auth/api/v1/register", registerRequest, Unit::class.java)

        assertThat(registerResponse.statusCodeValue).isEqualTo(200)

        // WHEN
        val loginRequest = AuthRequest(username = "jack.torrance@theoverlook.com", "654321")
        val loginResponse = restTemplate.postForEntity("http://localhost:$port/auth/api/v1/login", loginRequest, DefaultProblem::class.java)

        // THEN
        assertThat(loginResponse.statusCodeValue).isEqualTo(401)
        assertThat(loginResponse.headers.contentType).isEqualTo(MediaType.APPLICATION_PROBLEM_JSON)
        assertThat(loginResponse.body?.title).isEqualTo("Invalid username or password")
        assertThat(loginResponse.body?.status).isEqualTo(Status.UNAUTHORIZED)
        assertThat(loginResponse.body?.type).isEqualTo(URI.create("api/v1/problems/INVALID_CREDENTIALS"))
        assertThat(loginResponse.body?.message).isEqualTo("")
    }
}