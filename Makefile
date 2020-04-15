
# These values should be set for running the entire test suite
# all must be valid
Portal="1.1.1.1"
Target="iqn.1992-04.com.emc:600009700bcbb70e3287017400000000"


all:check int-test

mock-test:
	go clean -cache
	go test -v -coverprofile=c.out --run=TestMock

int-test: 
	GOISCSI_PORTAL=$(Portal) GOISCSI_TARGET=$(Target)  \
		 go test -v -timeout 20m -coverprofile=c.out -coverpkg ./...

gocover:
	go tool cover -html=c.out

check:
	gofmt -d .
	golint -set_exit_status
	go vet
