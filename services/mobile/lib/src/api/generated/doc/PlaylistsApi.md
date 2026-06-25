# inori_api.api.PlaylistsApi

## Load the API package
```dart
import 'package:inori_api/api.dart';
```

All URIs are relative to *http://127.0.0.1:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**apiV1AdminCatalogPlaylistsGet**](PlaylistsApi.md#apiv1admincatalogplaylistsget) | **GET** /api/v1/admin/catalog/playlists | List all playlists
[**apiV1AdminCatalogPlaylistsIdDelete**](PlaylistsApi.md#apiv1admincatalogplaylistsiddelete) | **DELETE** /api/v1/admin/catalog/playlists/{id} | Delete a playlist
[**apiV1AdminCatalogPlaylistsIdGet**](PlaylistsApi.md#apiv1admincatalogplaylistsidget) | **GET** /api/v1/admin/catalog/playlists/{id} | Get a playlist by ID
[**apiV1AdminCatalogPlaylistsIdPatch**](PlaylistsApi.md#apiv1admincatalogplaylistsidpatch) | **PATCH** /api/v1/admin/catalog/playlists/{id} | Update playlist metadata
[**apiV1AdminCatalogPlaylistsIdTracksGet**](PlaylistsApi.md#apiv1admincatalogplaylistsidtracksget) | **GET** /api/v1/admin/catalog/playlists/{id}/tracks | List the tracks of a playlist in playlist order
[**apiV1AdminCatalogPlaylistsIdTracksPost**](PlaylistsApi.md#apiv1admincatalogplaylistsidtrackspost) | **POST** /api/v1/admin/catalog/playlists/{id}/tracks | Append a track to a playlist
[**apiV1AdminCatalogPlaylistsIdTracksPut**](PlaylistsApi.md#apiv1admincatalogplaylistsidtracksput) | **PUT** /api/v1/admin/catalog/playlists/{id}/tracks | Replace the ordered track list of a playlist
[**apiV1AdminCatalogPlaylistsIdTracksTrackIdDelete**](PlaylistsApi.md#apiv1admincatalogplaylistsidtrackstrackiddelete) | **DELETE** /api/v1/admin/catalog/playlists/{id}/tracks/{trackId} | Remove first occurrence of a track from a playlist
[**apiV1AdminCatalogPlaylistsPost**](PlaylistsApi.md#apiv1admincatalogplaylistspost) | **POST** /api/v1/admin/catalog/playlists | Create a playlist
[**apiV1CatalogPlaylistsGet**](PlaylistsApi.md#apiv1catalogplaylistsget) | **GET** /api/v1/catalog/playlists | List playlists (viewer)
[**apiV1CatalogPlaylistsIdGet**](PlaylistsApi.md#apiv1catalogplaylistsidget) | **GET** /api/v1/catalog/playlists/{id} | Get a playlist by ID (viewer)
[**apiV1CatalogPlaylistsIdTracksGet**](PlaylistsApi.md#apiv1catalogplaylistsidtracksget) | **GET** /api/v1/catalog/playlists/{id}/tracks | List the tracks of a playlist in playlist order (viewer)


# **apiV1AdminCatalogPlaylistsGet**
> ApiV1AdminCatalogPlaylistsGet200Response apiV1AdminCatalogPlaylistsGet(limit, offset, sortBy, sortOrder)

List all playlists

List playlists. sortBy: name (default), createdAt, updatedAt.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getPlaylistsApi();
final int limit = 56; // int | Maximum number of items to return (1–500, default 50).
final int offset = 56; // int | Zero-based index of the first item to return.
final String sortBy = sortBy_example; // String | Field to sort by. See endpoint description for valid values.
final String sortOrder = sortOrder_example; // String | Sort direction.

try {
    final response = api.apiV1AdminCatalogPlaylistsGet(limit, offset, sortBy, sortOrder);
    print(response);
} on DioException catch (e) {
    print('Exception when calling PlaylistsApi->apiV1AdminCatalogPlaylistsGet: $e\n');
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

final api = InoriApi().getPlaylistsApi();
final String id = id_example; // String | Catalog entity identifier

try {
    api.apiV1AdminCatalogPlaylistsIdDelete(id);
} on DioException catch (e) {
    print('Exception when calling PlaylistsApi->apiV1AdminCatalogPlaylistsIdDelete: $e\n');
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

final api = InoriApi().getPlaylistsApi();
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.apiV1AdminCatalogPlaylistsIdGet(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling PlaylistsApi->apiV1AdminCatalogPlaylistsIdGet: $e\n');
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

final api = InoriApi().getPlaylistsApi();
final String id = id_example; // String | Catalog entity identifier
final UpdatePlaylistRequest updatePlaylistRequest = ; // UpdatePlaylistRequest | 

try {
    final response = api.apiV1AdminCatalogPlaylistsIdPatch(id, updatePlaylistRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling PlaylistsApi->apiV1AdminCatalogPlaylistsIdPatch: $e\n');
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

final api = InoriApi().getPlaylistsApi();
final String id = id_example; // String | Catalog entity identifier
final int limit = 56; // int | Maximum number of tracks to return (1–500, default 50).
final int offset = 56; // int | Zero-based index of the first track to return.

try {
    final response = api.apiV1AdminCatalogPlaylistsIdTracksGet(id, limit, offset);
    print(response);
} on DioException catch (e) {
    print('Exception when calling PlaylistsApi->apiV1AdminCatalogPlaylistsIdTracksGet: $e\n');
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

final api = InoriApi().getPlaylistsApi();
final String id = id_example; // String | Catalog entity identifier
final AddPlaylistTrackRequest addPlaylistTrackRequest = ; // AddPlaylistTrackRequest | 

try {
    final response = api.apiV1AdminCatalogPlaylistsIdTracksPost(id, addPlaylistTrackRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling PlaylistsApi->apiV1AdminCatalogPlaylistsIdTracksPost: $e\n');
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

final api = InoriApi().getPlaylistsApi();
final String id = id_example; // String | Catalog entity identifier
final SetPlaylistTracksRequest setPlaylistTracksRequest = ; // SetPlaylistTracksRequest | 

try {
    final response = api.apiV1AdminCatalogPlaylistsIdTracksPut(id, setPlaylistTracksRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling PlaylistsApi->apiV1AdminCatalogPlaylistsIdTracksPut: $e\n');
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

final api = InoriApi().getPlaylistsApi();
final String trackId = trackId_example; // String | 
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.apiV1AdminCatalogPlaylistsIdTracksTrackIdDelete(trackId, id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling PlaylistsApi->apiV1AdminCatalogPlaylistsIdTracksTrackIdDelete: $e\n');
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

final api = InoriApi().getPlaylistsApi();
final CreatePlaylistRequest createPlaylistRequest = ; // CreatePlaylistRequest | 

try {
    final response = api.apiV1AdminCatalogPlaylistsPost(createPlaylistRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling PlaylistsApi->apiV1AdminCatalogPlaylistsPost: $e\n');
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

final api = InoriApi().getPlaylistsApi();
final int limit = 56; // int | Maximum number of items to return (1–500, default 50).
final int offset = 56; // int | Zero-based index of the first item to return.
final String sortBy = sortBy_example; // String | Field to sort by. See endpoint description for valid values.
final String sortOrder = sortOrder_example; // String | Sort direction.

try {
    final response = api.apiV1CatalogPlaylistsGet(limit, offset, sortBy, sortOrder);
    print(response);
} on DioException catch (e) {
    print('Exception when calling PlaylistsApi->apiV1CatalogPlaylistsGet: $e\n');
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

final api = InoriApi().getPlaylistsApi();
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.apiV1CatalogPlaylistsIdGet(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling PlaylistsApi->apiV1CatalogPlaylistsIdGet: $e\n');
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

final api = InoriApi().getPlaylistsApi();
final String id = id_example; // String | Catalog entity identifier
final int limit = 56; // int | Maximum number of tracks to return (1–500, default 50).
final int offset = 56; // int | Zero-based index of the first track to return.

try {
    final response = api.apiV1CatalogPlaylistsIdTracksGet(id, limit, offset);
    print(response);
} on DioException catch (e) {
    print('Exception when calling PlaylistsApi->apiV1CatalogPlaylistsIdTracksGet: $e\n');
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

