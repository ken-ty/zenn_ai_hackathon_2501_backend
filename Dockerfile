FROM swaggerapi/swagger-ui

COPY swagger.json /usr/share/nginx/html/swagger.json

ENV SWAGGER_JSON=/usr/share/nginx/html/swagger.json
