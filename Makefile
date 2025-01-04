export OUTPUT := simple
export YAML_PATH := docs/step.yaml

build:
	go build -o bin/ ./cmd/sheryl/;
exec: build
	bin/sheryl $(wordlist 2, $(words $(MAKECMDGOALS) - 2), $(MAKECMDGOALS)) -c $(YAML_PATH) -o $(OUTPUT);
run: build
	bin/sheryl run -c $(YAML_PATH) -o $(OUTPUT);
