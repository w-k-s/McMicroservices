FROM openjdk:17-alpine

HEALTHCHECK --interval=30s --timeout=60s CMD curl -f http://localhost:8080/actuator/health || exit 1

ARG JAR_FILE=build/libs/*.jar
COPY ${JAR_FILE} app.jar

CMD ["java", "-jar", "app.jar"]