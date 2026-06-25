# inori_api.model.CatalogUpdateTrackRequest

## Load the model package
```dart
import 'package:inori_api/api.dart';
```

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**title** | **String** | Track title. Must not be empty if provided. | [optional] 
**sortTitle** | **String** | Sort key for the track title. May be empty to clear. | [optional] 
**artistId** | **String** | ID of the performing artist. Must reference an existing artist. | [optional] 
**albumId** | **String** | ID of the parent album. Empty string removes album association. | [optional] 
**trackNumber** | **int** |  | [optional] 
**discNumber** | **int** |  | [optional] 
**durationMs** | **int** | Track duration in milliseconds. | [optional] 
**genre** | **String** |  | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


