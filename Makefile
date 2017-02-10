docker-image:
	docker build -t ccbrown/gggtracker .

format:
	gofmt -s -w .

pre-commit: format docker-image
