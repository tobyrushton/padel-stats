gen:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g docs.go -d pkg/api/handlers,libs/auth -o openapi
	npx swagger2openapi openapi/swagger.json --yaml --outfile openapi/openapi.yaml
	npx openapi-typescript openapi/openapi.yaml -o pkg/www/src/types/api.ts