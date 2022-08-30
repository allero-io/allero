build:
	go build -ldflags="-X github.com/allero-io/allero/cmd.CliVersion=0.0.2" -o allero

build-prod:
	go build -tags=production -ldflags="-X github.com/allero-io/allero/cmd.CliVersion=0.0.2" -o allero

run: 
	go run main.go

validate:
	go run -ldflags="-X github.com/allero-io/allero/cmd.CliVersion=test" main.go validate

create-bin:
	goreleaser --snapshot --skip-publish --rm-dist

fetch: 
	# go run main.go fetch github supran2811/familyApp
	go run main.go fetch github curbengh/hexo-yam
