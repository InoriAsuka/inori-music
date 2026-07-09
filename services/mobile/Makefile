.PHONY: gen:api build:watch analyze test clean

OPENAPI_SPEC := ../../packages/api-contract/openapi/storage-admin.v1.json
GENERATED_DIR := lib/src/api/generated

## Generate Dart API client from OpenAPI spec
gen:api:
	npx @openapitools/openapi-generator-cli generate \
		-i $(OPENAPI_SPEC) \
		-g dart-dio \
		-o $(GENERATED_DIR) \
		--additional-properties="pubName=inori_api,nullSafe=true,serializationLibrary=json_serializable,pubVersion=3.0.0"
	# openapi-generator's dart-dio template hardcodes the SDK floor (currently
	# >=3.5.0) with no --additional-properties override available. Our outer
	# pubspec.yaml requires ^3.12.2, so a mismatched floor here resolves this
	# package to a lower languageVersion in .dart_tool/package_config.json —
	# which breaks part-file compilation with "language version override has to
	# be the same in the library and its part(s)" (CFE-only; flutter analyze
	# stays clean). Force it back in sync after every regeneration.
	sed -i.bak "s/^  sdk: .*/  sdk: '>=3.12.2 <4.0.0'/" $(GENERATED_DIR)/pubspec.yaml
	rm -f $(GENERATED_DIR)/pubspec.yaml.bak
	@echo "✅ API client generated at $(GENERATED_DIR)"

## Run build_runner once (generate Riverpod providers, Freezed models)
build:
	dart run build_runner build --delete-conflicting-outputs

## Watch mode for development
build:watch:
	dart run build_runner watch --delete-conflicting-outputs

## Run flutter analyze
analyze:
	flutter analyze

## Run tests
test:
	flutter test

## Clean build artifacts
clean:
	flutter clean
	rm -rf .dart_tool/build

## Install dependencies
deps:
	flutter pub get
