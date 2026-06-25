# inori_api.api.CatalogApi

## Load the API package
```dart
import 'package:inori_api/api.dart';
```

All URIs are relative to *http://127.0.0.1:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**adminListAlbumsByArtist**](CatalogApi.md#adminlistalbumsbyartist) | **GET** /api/v1/admin/catalog/artists/{id}/albums | List albums by artist (admin)
[**adminListTracksByAlbum**](CatalogApi.md#adminlisttracksbyalbum) | **GET** /api/v1/admin/catalog/albums/{id}/tracks | List tracks by album (admin)
[**adminListTracksByArtist**](CatalogApi.md#adminlisttracksbyartist) | **GET** /api/v1/admin/catalog/artists/{id}/tracks | List tracks by artist (admin)
[**apiV1AdminCatalogPlaylistsGet**](CatalogApi.md#apiv1admincatalogplaylistsget) | **GET** /api/v1/admin/catalog/playlists | List all playlists
[**apiV1AdminCatalogPlaylistsIdDelete**](CatalogApi.md#apiv1admincatalogplaylistsiddelete) | **DELETE** /api/v1/admin/catalog/playlists/{id} | Delete a playlist
[**apiV1AdminCatalogPlaylistsIdGet**](CatalogApi.md#apiv1admincatalogplaylistsidget) | **GET** /api/v1/admin/catalog/playlists/{id} | Get a playlist by ID
[**apiV1AdminCatalogPlaylistsIdPatch**](CatalogApi.md#apiv1admincatalogplaylistsidpatch) | **PATCH** /api/v1/admin/catalog/playlists/{id} | Update playlist metadata
[**apiV1AdminCatalogPlaylistsIdTracksGet**](CatalogApi.md#apiv1admincatalogplaylistsidtracksget) | **GET** /api/v1/admin/catalog/playlists/{id}/tracks | List the tracks of a playlist in playlist order
[**apiV1AdminCatalogPlaylistsIdTracksPost**](CatalogApi.md#apiv1admincatalogplaylistsidtrackspost) | **POST** /api/v1/admin/catalog/playlists/{id}/tracks | Append a track to a playlist
[**apiV1AdminCatalogPlaylistsIdTracksPut**](CatalogApi.md#apiv1admincatalogplaylistsidtracksput) | **PUT** /api/v1/admin/catalog/playlists/{id}/tracks | Replace the ordered track list of a playlist
[**apiV1AdminCatalogPlaylistsIdTracksTrackIdDelete**](CatalogApi.md#apiv1admincatalogplaylistsidtrackstrackiddelete) | **DELETE** /api/v1/admin/catalog/playlists/{id}/tracks/{trackId} | Remove first occurrence of a track from a playlist
[**apiV1AdminCatalogPlaylistsPost**](CatalogApi.md#apiv1admincatalogplaylistspost) | **POST** /api/v1/admin/catalog/playlists | Create a playlist
[**apiV1CatalogPlaylistsGet**](CatalogApi.md#apiv1catalogplaylistsget) | **GET** /api/v1/catalog/playlists | List playlists (viewer)
[**apiV1CatalogPlaylistsIdGet**](CatalogApi.md#apiv1catalogplaylistsidget) | **GET** /api/v1/catalog/playlists/{id} | Get a playlist by ID (viewer)
[**apiV1CatalogPlaylistsIdTracksGet**](CatalogApi.md#apiv1catalogplaylistsidtracksget) | **GET** /api/v1/catalog/playlists/{id}/tracks | List the tracks of a playlist in playlist order (viewer)
[**apiV1CatalogTracksIdStreamGet**](CatalogApi.md#apiv1catalogtracksidstreamget) | **GET** /api/v1/catalog/tracks/{id}/stream | Stream track audio
[**getCatalogAlbum**](CatalogApi.md#getcatalogalbum) | **GET** /api/v1/catalog/albums/{id} | Get album
[**getCatalogArtist**](CatalogApi.md#getcatalogartist) | **GET** /api/v1/catalog/artists/{id} | Get artist
[**getCatalogTrack**](CatalogApi.md#getcatalogtrack) | **GET** /api/v1/catalog/tracks/{id} | Get track
[**getTrackPlaybackDescriptor**](CatalogApi.md#gettrackplaybackdescriptor) | **GET** /api/v1/catalog/tracks/{id}/playback | Get track playback descriptor
[**getViewerCatalogAlbumStats**](CatalogApi.md#getviewercatalogalbumstats) | **GET** /api/v1/catalog/stats/albums | Get per-album stats breakdown
[**getViewerCatalogArtistStats**](CatalogApi.md#getviewercatalogartiststats) | **GET** /api/v1/catalog/stats/artists | Get per-artist stats breakdown
[**getViewerCatalogPlaylistStats**](CatalogApi.md#getviewercatalogplayliststats) | **GET** /api/v1/catalog/stats/playlists | Get per-playlist stats breakdown
[**getViewerCatalogStats**](CatalogApi.md#getviewercatalogstats) | **GET** /api/v1/catalog/stats | Get catalog stats
[**listAlbumsByArtist**](CatalogApi.md#listalbumsbyartist) | **GET** /api/v1/catalog/artists/{id}/albums | List albums by artist
[**listCatalogAlbums**](CatalogApi.md#listcatalogalbums) | **GET** /api/v1/catalog/albums | List albums
[**listCatalogArtists**](CatalogApi.md#listcatalogartists) | **GET** /api/v1/catalog/artists | List artists
[**listCatalogSearch**](CatalogApi.md#listcatalogsearch) | **GET** /api/v1/catalog/search | Search catalog
[**listCatalogTracks**](CatalogApi.md#listcatalogtracks) | **GET** /api/v1/catalog/tracks | List tracks
[**listRecentlyAddedCatalogItems**](CatalogApi.md#listrecentlyaddedcatalogitems) | **GET** /api/v1/catalog/recently-added | List recently added catalog items
[**listRecentlyUpdatedCatalogItems**](CatalogApi.md#listrecentlyupdatedcatalogitems) | **GET** /api/v1/catalog/recently-updated | List recently updated catalog items
[**listTracksByAlbum**](CatalogApi.md#listtracksbyalbum) | **GET** /api/v1/catalog/albums/{id}/tracks | List tracks by album
[**listTracksByArtist**](CatalogApi.md#listtracksbyartist) | **GET** /api/v1/catalog/artists/{id}/tracks | List tracks by artist
[**searchCatalog**](CatalogApi.md#searchcatalog) | **GET** /api/v1/admin/catalog/search | Search catalog


# **adminListAlbumsByArtist**
> ApiV1AdminCatalogAlbumsGet200Response adminListAlbumsByArtist(id, limit, offset, sortBy, sortOrder, releaseYearMin, releaseYearMax)

List albums by artist (admin)

List albums belonging to an artist. sortBy: title (default), sortTitle, releaseYear, createdAt, updatedAt.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier
final int limit = 56; // int | Max items to return (1–500, default 50).
final int offset = 56; // int | Zero-based start index.
final String sortBy = sortBy_example; // String | Field to sort by.
final String sortOrder = sortOrder_example; // String | Sort direction.
final int releaseYearMin = 56; // int | Include only albums with release year >= this value
final int releaseYearMax = 56; // int | Include only albums with release year <= this value

try {
    final response = api.adminListAlbumsByArtist(id, limit, offset, sortBy, sortOrder, releaseYearMin, releaseYearMax);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->adminListAlbumsByArtist: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **limit** | **int**| Max items to return (1–500, default 50). | [optional] [default to 50]
 **offset** | **int**| Zero-based start index. | [optional] [default to 0]
 **sortBy** | **String**| Field to sort by. | [optional] 
 **sortOrder** | **String**| Sort direction. | [optional] [default to 'asc']
 **releaseYearMin** | **int**| Include only albums with release year >= this value | [optional] 
 **releaseYearMax** | **int**| Include only albums with release year <= this value | [optional] 

### Return type

[**ApiV1AdminCatalogAlbumsGet200Response**](ApiV1AdminCatalogAlbumsGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **adminListTracksByAlbum**
> ApiV1AdminCatalogTracksGet200Response adminListTracksByAlbum(id, limit, offset, sortBy, sortOrder, genre)

List tracks by album (admin)

List tracks belonging to an album. sortBy: title (default), sortTitle, trackNumber, discNumber, durationMs, createdAt, updatedAt.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier
final int limit = 56; // int | Max items to return (1–500, default 50).
final int offset = 56; // int | Zero-based start index.
final String sortBy = sortBy_example; // String | Field to sort by.
final String sortOrder = sortOrder_example; // String | Sort direction.
final String genre = genre_example; // String | Filter tracks by genre (case-insensitive exact match)

try {
    final response = api.adminListTracksByAlbum(id, limit, offset, sortBy, sortOrder, genre);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->adminListTracksByAlbum: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **limit** | **int**| Max items to return (1–500, default 50). | [optional] [default to 50]
 **offset** | **int**| Zero-based start index. | [optional] [default to 0]
 **sortBy** | **String**| Field to sort by. | [optional] 
 **sortOrder** | **String**| Sort direction. | [optional] [default to 'asc']
 **genre** | **String**| Filter tracks by genre (case-insensitive exact match) | [optional] 

### Return type

[**ApiV1AdminCatalogTracksGet200Response**](ApiV1AdminCatalogTracksGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **adminListTracksByArtist**
> ApiV1AdminCatalogTracksGet200Response adminListTracksByArtist(id, limit, offset, sortBy, sortOrder, genre)

List tracks by artist (admin)

List tracks belonging to an artist. sortBy: title (default), sortTitle, trackNumber, discNumber, durationMs, createdAt, updatedAt.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier
final int limit = 56; // int | Max items to return (1–500, default 50).
final int offset = 56; // int | Zero-based start index.
final String sortBy = sortBy_example; // String | Field to sort by.
final String sortOrder = sortOrder_example; // String | Sort direction.
final String genre = genre_example; // String | Filter tracks by genre (case-insensitive exact match)

try {
    final response = api.adminListTracksByArtist(id, limit, offset, sortBy, sortOrder, genre);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->adminListTracksByArtist: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **limit** | **int**| Max items to return (1–500, default 50). | [optional] [default to 50]
 **offset** | **int**| Zero-based start index. | [optional] [default to 0]
 **sortBy** | **String**| Field to sort by. | [optional] 
 **sortOrder** | **String**| Sort direction. | [optional] [default to 'asc']
 **genre** | **String**| Filter tracks by genre (case-insensitive exact match) | [optional] 

### Return type

[**ApiV1AdminCatalogTracksGet200Response**](ApiV1AdminCatalogTracksGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogPlaylistsGet**
> ApiV1AdminCatalogPlaylistsGet200Response apiV1AdminCatalogPlaylistsGet(limit, offset, sortBy, sortOrder)

List all playlists

List playlists. sortBy: name (default), createdAt, updatedAt.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final int limit = 56; // int | Maximum number of items to return (1–500, default 50).
final int offset = 56; // int | Zero-based index of the first item to return.
final String sortBy = sortBy_example; // String | Field to sort by. See endpoint description for valid values.
final String sortOrder = sortOrder_example; // String | Sort direction.

try {
    final response = api.apiV1AdminCatalogPlaylistsGet(limit, offset, sortBy, sortOrder);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->apiV1AdminCatalogPlaylistsGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **limit** | **int**| Maximum number of items to return (1–500, default 50). | [optional] [default to 50]
 **offset** | **int**| Zero-based index of the first item to return. | [optional] [default to 0]
 **sortBy** | **String**| Field to sort by. See endpoint description for valid values. | [optional] 
 **sortOrder** | **String**| Sort direction. | [optional] [default to 'asc']

### Return type

[**ApiV1AdminCatalogPlaylistsGet200Response**](ApiV1AdminCatalogPlaylistsGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogPlaylistsIdDelete**
> apiV1AdminCatalogPlaylistsIdDelete(id)

Delete a playlist

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier

try {
    api.apiV1AdminCatalogPlaylistsIdDelete(id);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->apiV1AdminCatalogPlaylistsIdDelete: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 

### Return type

void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: Not defined

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogPlaylistsIdGet**
> Playlist apiV1AdminCatalogPlaylistsIdGet(id)

Get a playlist by ID

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.apiV1AdminCatalogPlaylistsIdGet(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->apiV1AdminCatalogPlaylistsIdGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 

### Return type

[**Playlist**](Playlist.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogPlaylistsIdPatch**
> Playlist apiV1AdminCatalogPlaylistsIdPatch(id, updatePlaylistRequest)

Update playlist metadata

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier
final UpdatePlaylistRequest updatePlaylistRequest = ; // UpdatePlaylistRequest | 

try {
    final response = api.apiV1AdminCatalogPlaylistsIdPatch(id, updatePlaylistRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->apiV1AdminCatalogPlaylistsIdPatch: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **updatePlaylistRequest** | [**UpdatePlaylistRequest**](UpdatePlaylistRequest.md)|  | 

### Return type

[**Playlist**](Playlist.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogPlaylistsIdTracksGet**
> ApiV1AdminCatalogPlaylistsIdTracksGet200Response apiV1AdminCatalogPlaylistsIdTracksGet(id, limit, offset)

List the tracks of a playlist in playlist order

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier
final int limit = 56; // int | Maximum number of tracks to return (1–500, default 50).
final int offset = 56; // int | Zero-based index of the first track to return.

try {
    final response = api.apiV1AdminCatalogPlaylistsIdTracksGet(id, limit, offset);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->apiV1AdminCatalogPlaylistsIdTracksGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **limit** | **int**| Maximum number of tracks to return (1–500, default 50). | [optional] [default to 50]
 **offset** | **int**| Zero-based index of the first track to return. | [optional] [default to 0]

### Return type

[**ApiV1AdminCatalogPlaylistsIdTracksGet200Response**](ApiV1AdminCatalogPlaylistsIdTracksGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogPlaylistsIdTracksPost**
> Playlist apiV1AdminCatalogPlaylistsIdTracksPost(id, addPlaylistTrackRequest)

Append a track to a playlist

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier
final AddPlaylistTrackRequest addPlaylistTrackRequest = ; // AddPlaylistTrackRequest | 

try {
    final response = api.apiV1AdminCatalogPlaylistsIdTracksPost(id, addPlaylistTrackRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->apiV1AdminCatalogPlaylistsIdTracksPost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **addPlaylistTrackRequest** | [**AddPlaylistTrackRequest**](AddPlaylistTrackRequest.md)|  | 

### Return type

[**Playlist**](Playlist.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogPlaylistsIdTracksPut**
> Playlist apiV1AdminCatalogPlaylistsIdTracksPut(id, setPlaylistTracksRequest)

Replace the ordered track list of a playlist

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier
final SetPlaylistTracksRequest setPlaylistTracksRequest = ; // SetPlaylistTracksRequest | 

try {
    final response = api.apiV1AdminCatalogPlaylistsIdTracksPut(id, setPlaylistTracksRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->apiV1AdminCatalogPlaylistsIdTracksPut: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **setPlaylistTracksRequest** | [**SetPlaylistTracksRequest**](SetPlaylistTracksRequest.md)|  | 

### Return type

[**Playlist**](Playlist.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogPlaylistsIdTracksTrackIdDelete**
> Playlist apiV1AdminCatalogPlaylistsIdTracksTrackIdDelete(trackId, id)

Remove first occurrence of a track from a playlist

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String trackId = trackId_example; // String | 
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.apiV1AdminCatalogPlaylistsIdTracksTrackIdDelete(trackId, id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->apiV1AdminCatalogPlaylistsIdTracksTrackIdDelete: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trackId** | **String**|  | 
 **id** | **String**| Catalog entity identifier | 

### Return type

[**Playlist**](Playlist.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogPlaylistsPost**
> Playlist apiV1AdminCatalogPlaylistsPost(createPlaylistRequest)

Create a playlist

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final CreatePlaylistRequest createPlaylistRequest = ; // CreatePlaylistRequest | 

try {
    final response = api.apiV1AdminCatalogPlaylistsPost(createPlaylistRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->apiV1AdminCatalogPlaylistsPost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createPlaylistRequest** | [**CreatePlaylistRequest**](CreatePlaylistRequest.md)|  | 

### Return type

[**Playlist**](Playlist.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1CatalogPlaylistsGet**
> ApiV1AdminCatalogPlaylistsGet200Response apiV1CatalogPlaylistsGet(limit, offset, sortBy, sortOrder)

List playlists (viewer)

List playlists. sortBy: name (default), createdAt, updatedAt.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final int limit = 56; // int | Maximum number of items to return (1–500, default 50).
final int offset = 56; // int | Zero-based index of the first item to return.
final String sortBy = sortBy_example; // String | Field to sort by. See endpoint description for valid values.
final String sortOrder = sortOrder_example; // String | Sort direction.

try {
    final response = api.apiV1CatalogPlaylistsGet(limit, offset, sortBy, sortOrder);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->apiV1CatalogPlaylistsGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **limit** | **int**| Maximum number of items to return (1–500, default 50). | [optional] [default to 50]
 **offset** | **int**| Zero-based index of the first item to return. | [optional] [default to 0]
 **sortBy** | **String**| Field to sort by. See endpoint description for valid values. | [optional] 
 **sortOrder** | **String**| Sort direction. | [optional] [default to 'asc']

### Return type

[**ApiV1AdminCatalogPlaylistsGet200Response**](ApiV1AdminCatalogPlaylistsGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1CatalogPlaylistsIdGet**
> Playlist apiV1CatalogPlaylistsIdGet(id)

Get a playlist by ID (viewer)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.apiV1CatalogPlaylistsIdGet(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->apiV1CatalogPlaylistsIdGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 

### Return type

[**Playlist**](Playlist.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1CatalogPlaylistsIdTracksGet**
> ApiV1AdminCatalogPlaylistsIdTracksGet200Response apiV1CatalogPlaylistsIdTracksGet(id, limit, offset)

List the tracks of a playlist in playlist order (viewer)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier
final int limit = 56; // int | Maximum number of tracks to return (1–500, default 50).
final int offset = 56; // int | Zero-based index of the first track to return.

try {
    final response = api.apiV1CatalogPlaylistsIdTracksGet(id, limit, offset);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->apiV1CatalogPlaylistsIdTracksGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **limit** | **int**| Maximum number of tracks to return (1–500, default 50). | [optional] [default to 50]
 **offset** | **int**| Zero-based index of the first track to return. | [optional] [default to 0]

### Return type

[**ApiV1AdminCatalogPlaylistsIdTracksGet200Response**](ApiV1AdminCatalogPlaylistsIdTracksGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1CatalogTracksIdStreamGet**
> apiV1CatalogTracksIdStreamGet(id, token)

Stream track audio

Proxies audio bytes with HTTP 206 Range support for filesystem-based storage backends. Authenticates via Bearer token in Authorization header or ?token= query parameter.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier
final String token = token_example; // String | Viewer bearer token (fallback when Authorization header cannot be set, e.g. <audio> src)

try {
    api.apiV1CatalogTracksIdStreamGet(id, token);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->apiV1CatalogTracksIdStreamGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **token** | **String**| Viewer bearer token (fallback when Authorization header cannot be set, e.g. <audio> src) | [optional] 

### Return type

void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getCatalogAlbum**
> CatalogAlbum getCatalogAlbum(id)

Get album

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.getCatalogAlbum(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->getCatalogAlbum: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 

### Return type

[**CatalogAlbum**](CatalogAlbum.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getCatalogArtist**
> CatalogArtist getCatalogArtist(id)

Get artist

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.getCatalogArtist(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->getCatalogArtist: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 

### Return type

[**CatalogArtist**](CatalogArtist.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getCatalogTrack**
> CatalogTrack getCatalogTrack(id)

Get track

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.getCatalogTrack(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->getCatalogTrack: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 

### Return type

[**CatalogTrack**](CatalogTrack.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getTrackPlaybackDescriptor**
> TrackPlaybackDescriptor getTrackPlaybackDescriptor(id)

Get track playback descriptor

Returns a metadata-only playback descriptor for the specified track. The linked media object must be active and have an audio asset kind (original_audio or transcoded_audio); otherwise 422 is returned.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.getTrackPlaybackDescriptor(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->getTrackPlaybackDescriptor: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 

### Return type

[**TrackPlaybackDescriptor**](TrackPlaybackDescriptor.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getViewerCatalogAlbumStats**
> CatalogAlbumStatsBreakdown getViewerCatalogAlbumStats()

Get per-album stats breakdown

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();

try {
    final response = api.getViewerCatalogAlbumStats();
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->getViewerCatalogAlbumStats: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**CatalogAlbumStatsBreakdown**](CatalogAlbumStatsBreakdown.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getViewerCatalogArtistStats**
> CatalogArtistStatsBreakdown getViewerCatalogArtistStats()

Get per-artist stats breakdown

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();

try {
    final response = api.getViewerCatalogArtistStats();
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->getViewerCatalogArtistStats: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**CatalogArtistStatsBreakdown**](CatalogArtistStatsBreakdown.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getViewerCatalogPlaylistStats**
> CatalogPlaylistStatsBreakdown getViewerCatalogPlaylistStats()

Get per-playlist stats breakdown

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();

try {
    final response = api.getViewerCatalogPlaylistStats();
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->getViewerCatalogPlaylistStats: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**CatalogPlaylistStatsBreakdown**](CatalogPlaylistStatsBreakdown.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getViewerCatalogStats**
> CatalogStats getViewerCatalogStats()

Get catalog stats

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();

try {
    final response = api.getViewerCatalogStats();
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->getViewerCatalogStats: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**CatalogStats**](CatalogStats.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listAlbumsByArtist**
> ApiV1AdminCatalogAlbumsGet200Response listAlbumsByArtist(id, limit, offset, sortBy, sortOrder, releaseYearMin, releaseYearMax)

List albums by artist

List albums belonging to an artist. sortBy: title (default), sortTitle, releaseYear, createdAt, updatedAt.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier
final int limit = 56; // int | Max items to return (1–500, default 50).
final int offset = 56; // int | Zero-based start index.
final String sortBy = sortBy_example; // String | Field to sort by.
final String sortOrder = sortOrder_example; // String | Sort direction.
final int releaseYearMin = 56; // int | Include only albums with release year >= this value
final int releaseYearMax = 56; // int | Include only albums with release year <= this value

try {
    final response = api.listAlbumsByArtist(id, limit, offset, sortBy, sortOrder, releaseYearMin, releaseYearMax);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->listAlbumsByArtist: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **limit** | **int**| Max items to return (1–500, default 50). | [optional] [default to 50]
 **offset** | **int**| Zero-based start index. | [optional] [default to 0]
 **sortBy** | **String**| Field to sort by. | [optional] 
 **sortOrder** | **String**| Sort direction. | [optional] [default to 'asc']
 **releaseYearMin** | **int**| Include only albums with release year >= this value | [optional] 
 **releaseYearMax** | **int**| Include only albums with release year <= this value | [optional] 

### Return type

[**ApiV1AdminCatalogAlbumsGet200Response**](ApiV1AdminCatalogAlbumsGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listCatalogAlbums**
> ApiV1AdminCatalogAlbumsGet200Response listCatalogAlbums(artistId, limit, offset, sortBy, sortOrder, releaseYearMin, releaseYearMax)

List albums

List albums. sortBy: title (default), sortTitle, releaseYear, createdAt, updatedAt.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String artistId = artistId_example; // String | 
final int limit = 56; // int | Maximum number of items to return (1–500, default 50).
final int offset = 56; // int | Zero-based index of the first item to return.
final String sortBy = sortBy_example; // String | Field to sort by. See endpoint description for valid values.
final String sortOrder = sortOrder_example; // String | Sort direction.
final int releaseYearMin = 56; // int | Include only albums with release year >= this value
final int releaseYearMax = 56; // int | Include only albums with release year <= this value

try {
    final response = api.listCatalogAlbums(artistId, limit, offset, sortBy, sortOrder, releaseYearMin, releaseYearMax);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->listCatalogAlbums: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **artistId** | **String**|  | [optional] 
 **limit** | **int**| Maximum number of items to return (1–500, default 50). | [optional] [default to 50]
 **offset** | **int**| Zero-based index of the first item to return. | [optional] [default to 0]
 **sortBy** | **String**| Field to sort by. See endpoint description for valid values. | [optional] 
 **sortOrder** | **String**| Sort direction. | [optional] [default to 'asc']
 **releaseYearMin** | **int**| Include only albums with release year >= this value | [optional] 
 **releaseYearMax** | **int**| Include only albums with release year <= this value | [optional] 

### Return type

[**ApiV1AdminCatalogAlbumsGet200Response**](ApiV1AdminCatalogAlbumsGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listCatalogArtists**
> ApiV1AdminCatalogArtistsGet200Response listCatalogArtists(limit, offset, sortBy, sortOrder)

List artists

List artists. sortBy: name (default), sortName, createdAt, updatedAt.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final int limit = 56; // int | Maximum number of items to return (1–500, default 50).
final int offset = 56; // int | Zero-based index of the first item to return.
final String sortBy = sortBy_example; // String | Field to sort by. See endpoint description for valid values.
final String sortOrder = sortOrder_example; // String | Sort direction.

try {
    final response = api.listCatalogArtists(limit, offset, sortBy, sortOrder);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->listCatalogArtists: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **limit** | **int**| Maximum number of items to return (1–500, default 50). | [optional] [default to 50]
 **offset** | **int**| Zero-based index of the first item to return. | [optional] [default to 0]
 **sortBy** | **String**| Field to sort by. See endpoint description for valid values. | [optional] 
 **sortOrder** | **String**| Sort direction. | [optional] [default to 'asc']

### Return type

[**ApiV1AdminCatalogArtistsGet200Response**](ApiV1AdminCatalogArtistsGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listCatalogSearch**
> CatalogSearchResult listCatalogSearch(q, types)

Search catalog

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String q = q_example; // String | Search query string.
final String types = types_example; // String | Filter search results to specific entity kinds (comma-separated: artist,album,track)

try {
    final response = api.listCatalogSearch(q, types);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->listCatalogSearch: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **q** | **String**| Search query string. | 
 **types** | **String**| Filter search results to specific entity kinds (comma-separated: artist,album,track) | [optional] 

### Return type

[**CatalogSearchResult**](CatalogSearchResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listCatalogTracks**
> ApiV1AdminCatalogTracksGet200Response listCatalogTracks(albumId, artistId, limit, offset, sortBy, sortOrder, genre)

List tracks

List tracks. sortBy: title (default), sortTitle, trackNumber, discNumber, durationMs, createdAt, updatedAt.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String albumId = albumId_example; // String | 
final String artistId = artistId_example; // String | 
final int limit = 56; // int | Maximum number of items to return (1–500, default 50).
final int offset = 56; // int | Zero-based index of the first item to return.
final String sortBy = sortBy_example; // String | Field to sort by. See endpoint description for valid values.
final String sortOrder = sortOrder_example; // String | Sort direction.
final String genre = genre_example; // String | Filter tracks by genre (case-insensitive exact match)

try {
    final response = api.listCatalogTracks(albumId, artistId, limit, offset, sortBy, sortOrder, genre);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->listCatalogTracks: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **albumId** | **String**|  | [optional] 
 **artistId** | **String**|  | [optional] 
 **limit** | **int**| Maximum number of items to return (1–500, default 50). | [optional] [default to 50]
 **offset** | **int**| Zero-based index of the first item to return. | [optional] [default to 0]
 **sortBy** | **String**| Field to sort by. See endpoint description for valid values. | [optional] 
 **sortOrder** | **String**| Sort direction. | [optional] [default to 'asc']
 **genre** | **String**| Filter tracks by genre (case-insensitive exact match) | [optional] 

### Return type

[**ApiV1AdminCatalogTracksGet200Response**](ApiV1AdminCatalogTracksGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listRecentlyAddedCatalogItems**
> RecentCatalogResult listRecentlyAddedCatalogItems(kind, limit)

List recently added catalog items

Returns a newest-first unified timeline of recently created artists, albums, tracks, and playlists. Use kind to filter to one entity type and limit to cap the response size (default 20, max 100).

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final RecentItemKind kind = ; // RecentItemKind | Optional entity kind filter.
final int limit = 56; // int | Maximum number of items to return. Defaults to 20; values above 100 are clamped.

try {
    final response = api.listRecentlyAddedCatalogItems(kind, limit);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->listRecentlyAddedCatalogItems: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kind** | [**RecentItemKind**](.md)| Optional entity kind filter. | [optional] 
 **limit** | **int**| Maximum number of items to return. Defaults to 20; values above 100 are clamped. | [optional] [default to 20]

### Return type

[**RecentCatalogResult**](RecentCatalogResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listRecentlyUpdatedCatalogItems**
> UpdatedCatalogResult listRecentlyUpdatedCatalogItems(kind, limit)

List recently updated catalog items

Returns a newest-first unified timeline of recently updated artists, albums, tracks, and playlists. Use kind to filter to one entity type and limit to cap the response size (default 20, max 100).

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final RecentItemKind kind = ; // RecentItemKind | Optional entity kind filter.
final int limit = 56; // int | Maximum number of items to return. Defaults to 20; values above 100 are clamped.

try {
    final response = api.listRecentlyUpdatedCatalogItems(kind, limit);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->listRecentlyUpdatedCatalogItems: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **kind** | [**RecentItemKind**](.md)| Optional entity kind filter. | [optional] 
 **limit** | **int**| Maximum number of items to return. Defaults to 20; values above 100 are clamped. | [optional] [default to 20]

### Return type

[**UpdatedCatalogResult**](UpdatedCatalogResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listTracksByAlbum**
> ApiV1AdminCatalogTracksGet200Response listTracksByAlbum(id, limit, offset, sortBy, sortOrder, genre)

List tracks by album

List tracks belonging to an album. sortBy: title (default), sortTitle, trackNumber, discNumber, durationMs, createdAt, updatedAt.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier
final int limit = 56; // int | Max items to return (1–500, default 50).
final int offset = 56; // int | Zero-based start index.
final String sortBy = sortBy_example; // String | Field to sort by.
final String sortOrder = sortOrder_example; // String | Sort direction.
final String genre = genre_example; // String | Filter tracks by genre (case-insensitive exact match)

try {
    final response = api.listTracksByAlbum(id, limit, offset, sortBy, sortOrder, genre);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->listTracksByAlbum: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **limit** | **int**| Max items to return (1–500, default 50). | [optional] [default to 50]
 **offset** | **int**| Zero-based start index. | [optional] [default to 0]
 **sortBy** | **String**| Field to sort by. | [optional] 
 **sortOrder** | **String**| Sort direction. | [optional] [default to 'asc']
 **genre** | **String**| Filter tracks by genre (case-insensitive exact match) | [optional] 

### Return type

[**ApiV1AdminCatalogTracksGet200Response**](ApiV1AdminCatalogTracksGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listTracksByArtist**
> ApiV1AdminCatalogTracksGet200Response listTracksByArtist(id, limit, offset, sortBy, sortOrder, genre)

List tracks by artist

List tracks belonging to an artist. sortBy: title (default), sortTitle, trackNumber, discNumber, durationMs, createdAt, updatedAt.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String id = id_example; // String | Catalog entity identifier
final int limit = 56; // int | Max items to return (1–500, default 50).
final int offset = 56; // int | Zero-based start index.
final String sortBy = sortBy_example; // String | Field to sort by.
final String sortOrder = sortOrder_example; // String | Sort direction.
final String genre = genre_example; // String | Filter tracks by genre (case-insensitive exact match)

try {
    final response = api.listTracksByArtist(id, limit, offset, sortBy, sortOrder, genre);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->listTracksByArtist: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **limit** | **int**| Max items to return (1–500, default 50). | [optional] [default to 50]
 **offset** | **int**| Zero-based start index. | [optional] [default to 0]
 **sortBy** | **String**| Field to sort by. | [optional] 
 **sortOrder** | **String**| Sort direction. | [optional] [default to 'asc']
 **genre** | **String**| Filter tracks by genre (case-insensitive exact match) | [optional] 

### Return type

[**ApiV1AdminCatalogTracksGet200Response**](ApiV1AdminCatalogTracksGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **searchCatalog**
> CatalogSearchResult searchCatalog(q, types)

Search catalog

Full-text search across artists, albums, and tracks. Returns ordered results with artist hits first, then albums, then tracks.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogApi();
final String q = q_example; // String | Search query string (tokenised using the simple text-search dictionary).
final String types = types_example; // String | Filter search results to specific entity kinds (comma-separated: artist,album,track)

try {
    final response = api.searchCatalog(q, types);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogApi->searchCatalog: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **q** | **String**| Search query string (tokenised using the simple text-search dictionary). | 
 **types** | **String**| Filter search results to specific entity kinds (comma-separated: artist,album,track) | [optional] 

### Return type

[**CatalogSearchResult**](CatalogSearchResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

