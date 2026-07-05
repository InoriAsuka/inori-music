# inori_api.model.SearchResultItem

## Load the model package
```dart
import 'package:inori_api/api.dart';
```

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**kind** | [**SearchResultKind**](SearchResultKind.md) |  | 
**artist** | [**CatalogArtist**](CatalogArtist.md) |  | [optional] 
**album** | [**CatalogAlbum**](CatalogAlbum.md) |  | [optional] 
**track** | [**CatalogTrack**](CatalogTrack.md) |  | [optional] 
**highlight** | **String** | HTML snippet (with <mark> tags around matched terms) for this item's matched field, when the search backend supports highlighting. Empty when the backend is degraded to PostgreSQL full-text search. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


