FROM golang:latest as build

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 go test /app/pkg/tax
RUN CGO_ENABLED=0 go test /app/pkg/deduction
RUN CGO_ENABLED=0 go test /app/pkg/middleware/auth

RUN CGO_ENABLED=0 go build -o /app/tax-api .

#--------------------- Run -------------------#
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/tax-api /app/tax-api

ENV PORT=8080
ENV DATABASE_URL=host=localhost port=5432 user=postgres password=postgres dbname=ktaxes sslmode=disable
ENV ADMIN_USERNAME=adminTax
ENV ADMIN_PASSWORD=admin!

EXPOSE 8080

CMD ["/app/tax-api"]