generate:
	erb Dockerfile.erb > Dockerfile
	erb pipeline.json.erb > pipeline.json

build: generate
	brocket build

release:
	brocket release
