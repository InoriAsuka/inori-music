# inori_api.model.CatalogUpdateAlbumRequest

## Load the model package
```dart
import 'package:inori_api/api.dart';
```

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**title** | **String** | Album title. Must not be empty if provided. | [optional] 
**sortTitle** | **String** | Sort key for the album title. May be empty to clear. | [optional] 
**artistId** | **String** | ID of the owning artist. Must reference an existing artist. | [optional] 
**releaseYear** | **int** | Release year. 0 clears the value. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


