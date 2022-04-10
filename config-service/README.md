# Config Service

This service is a centralized configuration store i.e. it provides configurations to all the other services
(well not at the moment, I need to update the kitchen-service to use `.properties`` files)

- The configurations are stored on AWS S3 in a bucket called `com.wks.mcmicroservices.configservice`.
- The properties files are named according to the following format `${application}-${profile}.properties` e.g. `order-service-default.properties`
- The properties can be fetching using curl `http://localhost:8888/${application}/${profile}` e.g. `http://localhost:8888/order-service/default`
- If a spring boot application is configured to pick up configurations from a config server, the application will pick up its own cconfigurations using the name configured in the `spring.application.name` property. For example, a spring application with the configuration `spring.application.name=order-service` will automatically pick up the configs returned from `http://localhost:8888/order-service/default`

## Sample Configuration Files

**Local**

```properties
spring.datasource.url=jdbc:postgresql://localhost:5432/db
spring.datasource.username=admin
spring.datasource.pssword=password
```

**Zalando K8s Postgres Operator**

```
spring.datasource.url=jdbc:postgresql://{postgresql-service}.{namespace}.svc.cluster.local:5432/{db}
spring.datasource.username={user}
spring.datasource.password={password}
```