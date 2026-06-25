# inori_api.api.AdminApi

## Load the API package
```dart
import 'package:inori_api/api.dart';
```

All URIs are relative to *http://127.0.0.1:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**adminClearUserFavorites**](AdminApi.md#adminclearuserfavorites) | **DELETE** /api/v1/admin/favorites/users/{userId}/tracks | Clear all favorites for a user (admin)
[**adminListUserFavorites**](AdminApi.md#adminlistuserfavorites) | **GET** /api/v1/admin/favorites/users/{userId}/tracks | List a user&#39;s favorites (admin)
[**adminRemoveUserFavoriteTrack**](AdminApi.md#adminremoveuserfavoritetrack) | **DELETE** /api/v1/admin/favorites/users/{userId}/tracks/{trackId} | Remove a single user-track favorite (admin)
[**apiV1AdminCatalogPlaylistsGet**](AdminApi.md#apiv1admincatalogplaylistsget) | **GET** /api/v1/admin/catalog/playlists | List all playlists
[**apiV1AdminCatalogPlaylistsIdDelete**](AdminApi.md#apiv1admincatalogplaylistsiddelete) | **DELETE** /api/v1/admin/catalog/playlists/{id} | Delete a playlist
[**apiV1AdminCatalogPlaylistsIdGet**](AdminApi.md#apiv1admincatalogplaylistsidget) | **GET** /api/v1/admin/catalog/playlists/{id} | Get a playlist by ID
[**apiV1AdminCatalogPlaylistsIdPatch**](AdminApi.md#apiv1admincatalogplaylistsidpatch) | **PATCH** /api/v1/admin/catalog/playlists/{id} | Update playlist metadata
[**apiV1AdminCatalogPlaylistsIdTracksGet**](AdminApi.md#apiv1admincatalogplaylistsidtracksget) | **GET** /api/v1/admin/catalog/playlists/{id}/tracks | List the tracks of a playlist in playlist order
[**apiV1AdminCatalogPlaylistsIdTracksPost**](AdminApi.md#apiv1admincatalogplaylistsidtrackspost) | **POST** /api/v1/admin/catalog/playlists/{id}/tracks | Append a track to a playlist
[**apiV1AdminCatalogPlaylistsIdTracksPut**](AdminApi.md#apiv1admincatalogplaylistsidtracksput) | **PUT** /api/v1/admin/catalog/playlists/{id}/tracks | Replace the ordered track list of a playlist
[**apiV1AdminCatalogPlaylistsIdTracksTrackIdDelete**](AdminApi.md#apiv1admincatalogplaylistsidtrackstrackiddelete) | **DELETE** /api/v1/admin/catalog/playlists/{id}/tracks/{trackId} | Remove first occurrence of a track from a playlist
[**apiV1AdminCatalogPlaylistsPost**](AdminApi.md#apiv1admincatalogplaylistspost) | **POST** /api/v1/admin/catalog/playlists | Create a playlist
[**apiV1AdminHistoryBatchDeletePost**](AdminApi.md#apiv1adminhistorybatchdeletepost) | **POST** /api/v1/admin/history/batch-delete | Batch-delete play events by IDs (admin)
[**apiV1AdminHistoryDelete**](AdminApi.md#apiv1adminhistorydelete) | **DELETE** /api/v1/admin/history | Bulk-delete play events within a time window (admin)
[**apiV1AdminHistoryEventIdDelete**](AdminApi.md#apiv1adminhistoryeventiddelete) | **DELETE** /api/v1/admin/history/{eventId} | Delete a single play event by ID (admin)
[**apiV1AdminHistoryEventIdGet**](AdminApi.md#apiv1adminhistoryeventidget) | **GET** /api/v1/admin/history/{eventId} | Get a single play event by ID (admin)
[**apiV1AdminHistoryEventIdPatch**](AdminApi.md#apiv1adminhistoryeventidpatch) | **PATCH** /api/v1/admin/history/{eventId} | Update a play event&#39;s playedAt timestamp (admin)
[**apiV1AdminHistoryGet**](AdminApi.md#apiv1adminhistoryget) | **GET** /api/v1/admin/history | List all play events (admin)
[**apiV1AdminHistoryTracksTrackIdDelete**](AdminApi.md#apiv1adminhistorytrackstrackiddelete) | **DELETE** /api/v1/admin/history/tracks/{trackId} | Delete all play events for a specific track across all users (admin)
[**apiV1AdminHistoryTracksTrackIdGet**](AdminApi.md#apiv1adminhistorytrackstrackidget) | **GET** /api/v1/admin/history/tracks/{trackId} | Get play history for a specific track across all users (admin)
[**apiV1AdminHistoryUsersUserIdDelete**](AdminApi.md#apiv1adminhistoryusersuseriddelete) | **DELETE** /api/v1/admin/history/users/{userId} | Delete all play events for a specific user (admin)
[**apiV1AdminHistoryUsersUserIdGet**](AdminApi.md#apiv1adminhistoryusersuseridget) | **GET** /api/v1/admin/history/users/{userId} | Get play history for a specific user (admin)
[**getAdminHistorySummary**](AdminApi.md#getadminhistorysummary) | **GET** /api/v1/admin/history/summary | Get global history summary


# **adminClearUserFavorites**
> adminClearUserFavorites(userId)

Clear all favorites for a user (admin)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminApi();
final String userId = userId_example; // String | 

try {
    api.adminClearUserFavorites(userId);
} on DioException catch (e) {
    print('Exception when calling AdminApi->adminClearUserFavorites: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **String**|  | 

### Return type

void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **adminListUserFavorites**
> AdminListUserFavorites200Response adminListUserFavorites(userId, limit, offset)

List a user's favorites (admin)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminApi();
final String userId = userId_example; // String | 
final int limit = 56; // int | 
final int offset = 56; // int | 

try {
    final response = api.adminListUserFavorites(userId, limit, offset);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminApi->adminListUserFavorites: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **String**|  | 
 **limit** | **int**|  | [optional] 
 **offset** | **int**|  | [optional] 

### Return type

[**AdminListUserFavorites200Response**](AdminListUserFavorites200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **adminRemoveUserFavoriteTrack**
> adminRemoveUserFavoriteTrack(userId, trackId)

Remove a single user-track favorite (admin)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminApi();
final String userId = userId_example; // String | 
final String trackId = trackId_example; // String | 

try {
    api.adminRemoveUserFavoriteTrack(userId, trackId);
} on DioException catch (e) {
    print('Exception when calling AdminApi->adminRemoveUserFavoriteTrack: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **String**|  | 
 **trackId** | **String**|  | 

### Return type

void (empty response body)

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

final api = InoriApi().getAdminApi();
final int limit = 56; // int | Maximum number of items to return (1–500, default 50).
final int offset = 56; // int | Zero-based index of the first item to return.
final String sortBy = sortBy_example; // String | Field to sort by. See endpoint description for valid values.
final String sortOrder = sortOrder_example; // String | Sort direction.

try {
    final response = api.apiV1AdminCatalogPlaylistsGet(limit, offset, sortBy, sortOrder);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminCatalogPlaylistsGet: $e\n');
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

final api = InoriApi().getAdminApi();
final String id = id_example; // String | Catalog entity identifier

try {
    api.apiV1AdminCatalogPlaylistsIdDelete(id);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminCatalogPlaylistsIdDelete: $e\n');
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

final api = InoriApi().getAdminApi();
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.apiV1AdminCatalogPlaylistsIdGet(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminCatalogPlaylistsIdGet: $e\n');
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

final api = InoriApi().getAdminApi();
final String id = id_example; // String | Catalog entity identifier
final UpdatePlaylistRequest updatePlaylistRequest = ; // UpdatePlaylistRequest | 

try {
    final response = api.apiV1AdminCatalogPlaylistsIdPatch(id, updatePlaylistRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminCatalogPlaylistsIdPatch: $e\n');
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

final api = InoriApi().getAdminApi();
final String id = id_example; // String | Catalog entity identifier
final int limit = 56; // int | Maximum number of tracks to return (1–500, default 50).
final int offset = 56; // int | Zero-based index of the first track to return.

try {
    final response = api.apiV1AdminCatalogPlaylistsIdTracksGet(id, limit, offset);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminCatalogPlaylistsIdTracksGet: $e\n');
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

final api = InoriApi().getAdminApi();
final String id = id_example; // String | Catalog entity identifier
final AddPlaylistTrackRequest addPlaylistTrackRequest = ; // AddPlaylistTrackRequest | 

try {
    final response = api.apiV1AdminCatalogPlaylistsIdTracksPost(id, addPlaylistTrackRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminCatalogPlaylistsIdTracksPost: $e\n');
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

final api = InoriApi().getAdminApi();
final String id = id_example; // String | Catalog entity identifier
final SetPlaylistTracksRequest setPlaylistTracksRequest = ; // SetPlaylistTracksRequest | 

try {
    final response = api.apiV1AdminCatalogPlaylistsIdTracksPut(id, setPlaylistTracksRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminCatalogPlaylistsIdTracksPut: $e\n');
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

final api = InoriApi().getAdminApi();
final String trackId = trackId_example; // String | 
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.apiV1AdminCatalogPlaylistsIdTracksTrackIdDelete(trackId, id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminCatalogPlaylistsIdTracksTrackIdDelete: $e\n');
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

final api = InoriApi().getAdminApi();
final CreatePlaylistRequest createPlaylistRequest = ; // CreatePlaylistRequest | 

try {
    final response = api.apiV1AdminCatalogPlaylistsPost(createPlaylistRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminCatalogPlaylistsPost: $e\n');
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

# **apiV1AdminHistoryBatchDeletePost**
> BatchDeleteResult apiV1AdminHistoryBatchDeletePost(batchDeleteRequest)

Batch-delete play events by IDs (admin)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminApi();
final BatchDeleteRequest batchDeleteRequest = ; // BatchDeleteRequest | 

try {
    final response = api.apiV1AdminHistoryBatchDeletePost(batchDeleteRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminHistoryBatchDeletePost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **batchDeleteRequest** | [**BatchDeleteRequest**](BatchDeleteRequest.md)|  | 

### Return type

[**BatchDeleteResult**](BatchDeleteResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminHistoryDelete**
> apiV1AdminHistoryDelete(since, until)

Bulk-delete play events within a time window (admin)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminApi();
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Delete events at or after this time (RFC3339). At least one of since or until is required.
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Delete events before this time (RFC3339, exclusive). At least one of since or until is required.

try {
    api.apiV1AdminHistoryDelete(since, until);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminHistoryDelete: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **since** | **DateTime**| Delete events at or after this time (RFC3339). At least one of since or until is required. | [optional] 
 **until** | **DateTime**| Delete events before this time (RFC3339, exclusive). At least one of since or until is required. | [optional] 

### Return type

void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminHistoryEventIdDelete**
> apiV1AdminHistoryEventIdDelete(eventId)

Delete a single play event by ID (admin)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminApi();
final String eventId = eventId_example; // String | The play event ID.

try {
    api.apiV1AdminHistoryEventIdDelete(eventId);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminHistoryEventIdDelete: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **eventId** | **String**| The play event ID. | 

### Return type

void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminHistoryEventIdGet**
> PlayEvent apiV1AdminHistoryEventIdGet(eventId)

Get a single play event by ID (admin)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminApi();
final String eventId = eventId_example; // String | The play event ID.

try {
    final response = api.apiV1AdminHistoryEventIdGet(eventId);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminHistoryEventIdGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **eventId** | **String**| The play event ID. | 

### Return type

[**PlayEvent**](PlayEvent.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminHistoryEventIdPatch**
> PlayEvent apiV1AdminHistoryEventIdPatch(eventId, updatePlayEventRequest)

Update a play event's playedAt timestamp (admin)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminApi();
final String eventId = eventId_example; // String | The play event ID.
final UpdatePlayEventRequest updatePlayEventRequest = ; // UpdatePlayEventRequest | Fields to update on the play event.

try {
    final response = api.apiV1AdminHistoryEventIdPatch(eventId, updatePlayEventRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminHistoryEventIdPatch: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **eventId** | **String**| The play event ID. | 
 **updatePlayEventRequest** | [**UpdatePlayEventRequest**](UpdatePlayEventRequest.md)| Fields to update on the play event. | 

### Return type

[**PlayEvent**](PlayEvent.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminHistoryGet**
> PlayEventList apiV1AdminHistoryGet(userId, trackId, since, until, limit, offset, order)

List all play events (admin)

Returns a paginated list of all play events across every user and track. Supports optional filtering by userId, trackId, since, and until.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminApi();
final String userId = userId_example; // String | Optional: filter to events recorded by this user.
final String trackId = trackId_example; // String | Optional: filter to events recorded for this track.
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Optional: include only events at or after this time (RFC3339, inclusive).
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Optional: include only events before this time (RFC3339, exclusive).
final int limit = 56; // int | Maximum number of events to return (1-500, default 50).
final int offset = 56; // int | Number of events to skip.
final String order = order_example; // String | Sort direction for results. \"desc\" (default) returns newest first; \"asc\" returns oldest first.

try {
    final response = api.apiV1AdminHistoryGet(userId, trackId, since, until, limit, offset, order);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminHistoryGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **String**| Optional: filter to events recorded by this user. | [optional] 
 **trackId** | **String**| Optional: filter to events recorded for this track. | [optional] 
 **since** | **DateTime**| Optional: include only events at or after this time (RFC3339, inclusive). | [optional] 
 **until** | **DateTime**| Optional: include only events before this time (RFC3339, exclusive). | [optional] 
 **limit** | **int**| Maximum number of events to return (1-500, default 50). | [optional] 
 **offset** | **int**| Number of events to skip. | [optional] 
 **order** | **String**| Sort direction for results. \"desc\" (default) returns newest first; \"asc\" returns oldest first. | [optional] [default to 'desc']

### Return type

[**PlayEventList**](PlayEventList.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminHistoryTracksTrackIdDelete**
> apiV1AdminHistoryTracksTrackIdDelete(trackId)

Delete all play events for a specific track across all users (admin)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminApi();
final String trackId = trackId_example; // String | ID of the track whose history to delete.

try {
    api.apiV1AdminHistoryTracksTrackIdDelete(trackId);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminHistoryTracksTrackIdDelete: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trackId** | **String**| ID of the track whose history to delete. | 

### Return type

void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminHistoryTracksTrackIdGet**
> PlayEventList apiV1AdminHistoryTracksTrackIdGet(trackId, userId, limit, offset, order, since, until)

Get play history for a specific track across all users (admin)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminApi();
final String trackId = trackId_example; // String | ID of the track whose history to retrieve.
final String userId = userId_example; // String | Optional: filter to a specific user.
final int limit = 56; // int | Maximum number of events to return (1–500, default 50).
final int offset = 56; // int | Number of events to skip.
final String order = order_example; // String | Sort direction for results. \"desc\" (default) returns newest first; \"asc\" returns oldest first.
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Include only events at or after this time (RFC3339, inclusive).
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Include only events before this time (RFC3339, exclusive).

try {
    final response = api.apiV1AdminHistoryTracksTrackIdGet(trackId, userId, limit, offset, order, since, until);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminHistoryTracksTrackIdGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trackId** | **String**| ID of the track whose history to retrieve. | 
 **userId** | **String**| Optional: filter to a specific user. | [optional] 
 **limit** | **int**| Maximum number of events to return (1–500, default 50). | [optional] 
 **offset** | **int**| Number of events to skip. | [optional] 
 **order** | **String**| Sort direction for results. \"desc\" (default) returns newest first; \"asc\" returns oldest first. | [optional] [default to 'desc']
 **since** | **DateTime**| Include only events at or after this time (RFC3339, inclusive). | [optional] 
 **until** | **DateTime**| Include only events before this time (RFC3339, exclusive). | [optional] 

### Return type

[**PlayEventList**](PlayEventList.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminHistoryUsersUserIdDelete**
> apiV1AdminHistoryUsersUserIdDelete(userId)

Delete all play events for a specific user (admin)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminApi();
final String userId = userId_example; // String | ID of the user whose history to delete.

try {
    api.apiV1AdminHistoryUsersUserIdDelete(userId);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminHistoryUsersUserIdDelete: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **String**| ID of the user whose history to delete. | 

### Return type

void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminHistoryUsersUserIdGet**
> PlayEventList apiV1AdminHistoryUsersUserIdGet(userId, trackId, limit, offset, order, since, until)

Get play history for a specific user (admin)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminApi();
final String userId = userId_example; // String | ID of the user whose history to retrieve.
final String trackId = trackId_example; // String | Optional: filter to a specific track.
final int limit = 56; // int | Maximum number of events to return (1–500, default 50).
final int offset = 56; // int | Number of events to skip.
final String order = order_example; // String | Sort direction for results. \"desc\" (default) returns newest first; \"asc\" returns oldest first.
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Include only events at or after this time (RFC3339, inclusive).
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Include only events before this time (RFC3339, exclusive).

try {
    final response = api.apiV1AdminHistoryUsersUserIdGet(userId, trackId, limit, offset, order, since, until);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminApi->apiV1AdminHistoryUsersUserIdGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **String**| ID of the user whose history to retrieve. | 
 **trackId** | **String**| Optional: filter to a specific track. | [optional] 
 **limit** | **int**| Maximum number of events to return (1–500, default 50). | [optional] 
 **offset** | **int**| Number of events to skip. | [optional] 
 **order** | **String**| Sort direction for results. \"desc\" (default) returns newest first; \"asc\" returns oldest first. | [optional] [default to 'desc']
 **since** | **DateTime**| Include only events at or after this time (RFC3339, inclusive). | [optional] 
 **until** | **DateTime**| Include only events before this time (RFC3339, exclusive). | [optional] 

### Return type

[**PlayEventList**](PlayEventList.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAdminHistorySummary**
> GlobalHistorySummary getAdminHistorySummary(since, until, limit)

Get global history summary

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminApi();
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | 
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | 
final int limit = 56; // int | 

try {
    final response = api.getAdminHistorySummary(since, until, limit);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminApi->getAdminHistorySummary: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **since** | **DateTime**|  | [optional] 
 **until** | **DateTime**|  | [optional] 
 **limit** | **int**|  | [optional] [default to 10]

### Return type

[**GlobalHistorySummary**](GlobalHistorySummary.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

