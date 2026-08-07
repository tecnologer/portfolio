.PHONY: site extract build test golden serve clean

# Render data/resume.yaml → page/index.html
site:
	go run . build

# Regenerate data/resume.yaml from page/reydavid_experience.md
extract:
	go run . extract

# Compile the profilegen binary
build:
	go build -o bin/profilegen .

# Run the test suite
test:
	go test ./...

# Regenerate the drift fixtures after a DELIBERATE parser change (AD-12)
golden:
	go test ./internal/portfolio -run TestDrift -update

# Preview the generated site at http://localhost:$(PORT)
PORT ?= 8081
serve: site
	@cd page && python3 -m http.server $(PORT)

clean:
	rm -rf bin page/index.html
