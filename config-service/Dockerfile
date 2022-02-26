FROM openjdk:17-alpine

ARG JAR_FILE=build/libs/*.jar
COPY ${JAR_FILE} app.jar

CMD ["java", "-jar", "app.jar"]