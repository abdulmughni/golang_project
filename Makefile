.PHONY: start token

start:
	@lsof -nP -iTCP:8080 -sTCP:LISTEN | awk 'NR>1 {print $$2}' | xargs -r kill -9; \
	go run main.go


access-token:
	@curl --request POST \
		--url 'https://solutionpilot-prod-us1.us.auth0.com/oauth/token' \
		--header 'content-type: application/json' \
		--data '{"grant_type":"password","username":"","password":"","audience":"https://solutionpilot-prod-primeapi-us.com/endpoint","client_id":"","client_secret":"BQ4VxIP1e0w6dTj8Fvb27a3G_uJoTjQKtYrXlyv3hMyD1wH57PAhGghrblK1OwSG"}'

id-token:
	@curl --request POST \
		--url 'https://solutionpilot-prod-us1.us.auth0.com/oauth/token' \
		--header 'content-type: application/json' \
		--data '{"grant_type":"password","username":"","password":"","audience":"https://solutionpilot-prod-primeapi-us.com/endpoint","scope":"openid profile email","client_id":"","client_secret":""}' | \
		jq '.id_token'

tokens:
	@curl --request POST \
		--url 'https://solutionpilot-prod-us1.us.auth0.com/oauth/token' \
		--header 'content-type: application/json' \
		--data '{"grant_type":"password","username":"","password":"","audience":"https://solutionpilot-prod-primeapi-us.com/endpoint","scope":"openid profile email","client_id":"","client_secret":""}' | \
		jq '{access_token: .access_token, id_token: .id_token}'

scan:
	@sonar-scanner \
		-Dsonar.projectKey=solutionpilot-primeapi \
		-Dsonar.sources=. \
		-Dsonar.host.url= \
		-Dsonar.token= \
		-Dsonar.go.coverage.reportPaths=coverage.out \
		-Dsonar.test.inclusions=**/*_test.go

test:
	@go test -v \
		-coverprofile=coverage.out \
		-coverpkg=./handlers/apis/... \
		./handlers/apis/tests/
		