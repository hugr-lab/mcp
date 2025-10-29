FROM openjdk:17-slim

RUN apt-get update && apt-get install -y \
    curl \
    libfreetype6 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN curl -L -o synthea.jar https://github.com/synthetichealth/synthea/releases/download/master-branch-latest/synthea-with-dependencies.jar

ENV SYNTHEA_STATE=Massachusetts
ENV SYNTHEA_COUNT=100

COPY synthea.properties /app/synthea.properties

VOLUME ["/output"]

CMD ["sh", "-c", "java -Djava.awt.headless=true -jar synthea.jar -p $SYNTHEA_COUNT $SYNTHEA_STATE --exporter.csv.export true --exporter.fhir.export false && cp -r output/csv /output"]
