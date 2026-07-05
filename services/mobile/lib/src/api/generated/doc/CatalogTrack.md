# inori_api.model.CatalogTrack

## Load the model package
```dart
import 'package:inori_api/api.dart';
```

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**albumId** | **String** |  | [optional] 
**artistId** | **String** |  | 
**createdAt** | [**DateTime**](DateTime.md) |  | 
**discNumber** | **int** |  | [optional] 
**durationMs** | **int** |  | [optional] 
**id** | **String** |  | 
**mediaObjectId** | **String** |  | 
**sortTitle** | **String** |  | [optional] 
**title** | **String** |  | 
**trackNumber** | **int** |  | [optional] 
**updatedAt** | [**DateTime**](DateTime.md) |  | 
**genre** | **String** | Optional genre tag for the track (e.g. Rock, Jazz, Classical) | [optional] 
**isFavorite** | **bool** | Whether the authenticated viewer has favorited this track. Always false on admin endpoints. | [optional] [default to false]
**lyricsMediaObjectId** | **String** | ID of the media object used as lyrics (optional) | [optional] 
**replayGainDb** | **double** | ReplayGain normalization value in dB (null = not analyzed) | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


