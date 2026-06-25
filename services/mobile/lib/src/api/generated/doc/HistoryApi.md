# inori_api.api.HistoryApi

## Load the API package
```dart
import 'package:inori_api/api.dart';
```

All URIs are relative to *http://127.0.0.1:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**addFavoriteTrack**](HistoryApi.md#addfavoritetrack) | **POST** /api/v1/me/favorites/tracks/{trackId} | Add a track to favorites (idempotent)
[**adminClearUserFavorites**](HistoryApi.md#adminclearuserfavorites) | **DELETE** /api/v1/admin/favorites/users/{userId}/tracks | Clear all favorites for a user (admin)
[**adminListUserFavorites**](HistoryApi.md#adminlistuserfavorites) | **GET** /api/v1/admin/favorites/users/{userId}/tracks | List a user&#39;s favorites (admin)
[**adminRemoveUserFavoriteTrack**](HistoryApi.md#adminremoveuserfavoritetrack) | **DELETE** /api/v1/admin/favorites/users/{userId}/tracks/{trackId} | Remove a single user-track favorite (admin)
[**apiV1AdminHistoryBatchDeletePost**](HistoryApi.md#apiv1adminhistorybatchdeletepost) | **POST** /api/v1/admin/history/batch-delete | Batch-delete play events by IDs (admin)
[**apiV1AdminHistoryDelete**](HistoryApi.md#apiv1adminhistorydelete) | **DELETE** /api/v1/admin/history | Bulk-delete play events within a time window (admin)
[**apiV1AdminHistoryEventIdDelete**](HistoryApi.md#apiv1adminhistoryeventiddelete) | **DELETE** /api/v1/admin/history/{eventId} | Delete a single play event by ID (admin)
[**apiV1AdminHistoryEventIdGet**](HistoryApi.md#apiv1adminhistoryeventidget) | **GET** /api/v1/admin/history/{eventId} | Get a single play event by ID (admin)
[**apiV1AdminHistoryEventIdPatch**](HistoryApi.md#apiv1adminhistoryeventidpatch) | **PATCH** /api/v1/admin/history/{eventId} | Update a play event&#39;s playedAt timestamp (admin)
[**apiV1AdminHistoryGet**](HistoryApi.md#apiv1adminhistoryget) | **GET** /api/v1/admin/history | List all play events (admin)
[**apiV1AdminHistoryTracksTrackIdDelete**](HistoryApi.md#apiv1adminhistorytrackstrackiddelete) | **DELETE** /api/v1/admin/history/tracks/{trackId} | Delete all play events for a specific track across all users (admin)
[**apiV1AdminHistoryTracksTrackIdGet**](HistoryApi.md#apiv1adminhistorytrackstrackidget) | **GET** /api/v1/admin/history/tracks/{trackId} | Get play history for a specific track across all users (admin)
[**apiV1AdminHistoryUsersUserIdDelete**](HistoryApi.md#apiv1adminhistoryusersuseriddelete) | **DELETE** /api/v1/admin/history/users/{userId} | Delete all play events for a specific user (admin)
[**apiV1AdminHistoryUsersUserIdGet**](HistoryApi.md#apiv1adminhistoryusersuseridget) | **GET** /api/v1/admin/history/users/{userId} | Get play history for a specific user (admin)
[**apiV1MeHistoryBatchDeletePost**](HistoryApi.md#apiv1mehistorybatchdeletepost) | **POST** /api/v1/me/history/batch-delete | Batch-delete own play events by IDs (viewer)
[**apiV1MeHistoryEventIdDelete**](HistoryApi.md#apiv1mehistoryeventiddelete) | **DELETE** /api/v1/me/history/{eventId} | Delete a single play event by ID (viewer, own events only)
[**apiV1MeHistoryEventIdGet**](HistoryApi.md#apiv1mehistoryeventidget) | **GET** /api/v1/me/history/{eventId} | Get a single play event by ID (viewer, own events only)
[**apiV1MeHistoryEventIdPatch**](HistoryApi.md#apiv1mehistoryeventidpatch) | **PATCH** /api/v1/me/history/{eventId} | Update a play event&#39;s playedAt timestamp (viewer, own events only)
[**apiV1MeHistoryStatsGet**](HistoryApi.md#apiv1mehistorystatsget) | **GET** /api/v1/me/history/stats | Get personal play history statistics for the authenticated user
[**apiV1MeHistoryTopTracksGet**](HistoryApi.md#apiv1mehistorytoptracksget) | **GET** /api/v1/me/history/top-tracks | Get top tracks for the authenticated user by play count
[**clearPlayHistory**](HistoryApi.md#clearplayhistory) | **DELETE** /api/v1/me/history | Clear play history
[**getAdminHistorySummary**](HistoryApi.md#getadminhistorysummary) | **GET** /api/v1/admin/history/summary | Get global history summary
[**getAdminTrackHistorySummary**](HistoryApi.md#getadmintrackhistorysummary) | **GET** /api/v1/admin/history/tracks/{trackId}/history-summary | Get track history summary
[**getAdminTrackTimeline**](HistoryApi.md#getadmintracktimeline) | **GET** /api/v1/admin/history/tracks/{trackId}/timeline | Track play timeline
[**getAdminUserHistorySummary**](HistoryApi.md#getadminuserhistorysummary) | **GET** /api/v1/admin/history/users/{userId}/history-summary | Get user history summary
[**getAdminUserTimeline**](HistoryApi.md#getadminusertimeline) | **GET** /api/v1/admin/history/users/{userId}/timeline | User play timeline
[**getMyHistorySummary**](HistoryApi.md#getmyhistorysummary) | **GET** /api/v1/me/history/summary | Get viewer history summary
[**getMyHistoryTimeline**](HistoryApi.md#getmyhistorytimeline) | **GET** /api/v1/me/history/timeline | Get the authenticated user&#39;s play history grouped by time bucket
[**getMyTrackStats**](HistoryApi.md#getmytrackstats) | **GET** /api/v1/me/history/tracks/{trackId}/stats | Get play stats for a track
[**getMyTrackSummary**](HistoryApi.md#getmytracksummary) | **GET** /api/v1/me/history/tracks/{trackId}/summary | Get viewer track summary
[**getMyTrackTimeline**](HistoryApi.md#getmytracktimeline) | **GET** /api/v1/me/history/tracks/{trackId}/timeline | Get viewer track play timeline
[**listFavoriteTracks**](HistoryApi.md#listfavoritetracks) | **GET** /api/v1/me/favorites/tracks | List favorite tracks (paginated)
[**listMyTrackHistory**](HistoryApi.md#listmytrackhistory) | **GET** /api/v1/me/history/tracks/{trackId} | List play events for a track
[**listPlayEvents**](HistoryApi.md#listplayevents) | **GET** /api/v1/me/history | List play events
[**recordPlayEvent**](HistoryApi.md#recordplayevent) | **POST** /api/v1/me/history | Record a play event
[**removeFavoriteTrack**](HistoryApi.md#removefavoritetrack) | **DELETE** /api/v1/me/favorites/tracks/{trackId} | Remove a track from favorites (idempotent)


# **addFavoriteTrack**
> addFavoriteTrack(trackId)

Add a track to favorites (idempotent)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final String trackId = trackId_example; // String | 

try {
    api.addFavoriteTrack(trackId);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->addFavoriteTrack: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trackId** | **String**|  | 

### Return type

void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **adminClearUserFavorites**
> adminClearUserFavorites(userId)

Clear all favorites for a user (admin)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final String userId = userId_example; // String | 

try {
    api.adminClearUserFavorites(userId);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->adminClearUserFavorites: $e\n');
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

final api = InoriApi().getHistoryApi();
final String userId = userId_example; // String | 
final int limit = 56; // int | 
final int offset = 56; // int | 

try {
    final response = api.adminListUserFavorites(userId, limit, offset);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->adminListUserFavorites: $e\n');
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

final api = InoriApi().getHistoryApi();
final String userId = userId_example; // String | 
final String trackId = trackId_example; // String | 

try {
    api.adminRemoveUserFavoriteTrack(userId, trackId);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->adminRemoveUserFavoriteTrack: $e\n');
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

# **apiV1AdminHistoryBatchDeletePost**
> BatchDeleteResult apiV1AdminHistoryBatchDeletePost(batchDeleteRequest)

Batch-delete play events by IDs (admin)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final BatchDeleteRequest batchDeleteRequest = ; // BatchDeleteRequest | 

try {
    final response = api.apiV1AdminHistoryBatchDeletePost(batchDeleteRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->apiV1AdminHistoryBatchDeletePost: $e\n');
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

final api = InoriApi().getHistoryApi();
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Delete events at or after this time (RFC3339). At least one of since or until is required.
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Delete events before this time (RFC3339, exclusive). At least one of since or until is required.

try {
    api.apiV1AdminHistoryDelete(since, until);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->apiV1AdminHistoryDelete: $e\n');
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

final api = InoriApi().getHistoryApi();
final String eventId = eventId_example; // String | The play event ID.

try {
    api.apiV1AdminHistoryEventIdDelete(eventId);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->apiV1AdminHistoryEventIdDelete: $e\n');
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

final api = InoriApi().getHistoryApi();
final String eventId = eventId_example; // String | The play event ID.

try {
    final response = api.apiV1AdminHistoryEventIdGet(eventId);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->apiV1AdminHistoryEventIdGet: $e\n');
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

final api = InoriApi().getHistoryApi();
final String eventId = eventId_example; // String | The play event ID.
final UpdatePlayEventRequest updatePlayEventRequest = ; // UpdatePlayEventRequest | Fields to update on the play event.

try {
    final response = api.apiV1AdminHistoryEventIdPatch(eventId, updatePlayEventRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->apiV1AdminHistoryEventIdPatch: $e\n');
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

final api = InoriApi().getHistoryApi();
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
    print('Exception when calling HistoryApi->apiV1AdminHistoryGet: $e\n');
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

final api = InoriApi().getHistoryApi();
final String trackId = trackId_example; // String | ID of the track whose history to delete.

try {
    api.apiV1AdminHistoryTracksTrackIdDelete(trackId);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->apiV1AdminHistoryTracksTrackIdDelete: $e\n');
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

final api = InoriApi().getHistoryApi();
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
    print('Exception when calling HistoryApi->apiV1AdminHistoryTracksTrackIdGet: $e\n');
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

final api = InoriApi().getHistoryApi();
final String userId = userId_example; // String | ID of the user whose history to delete.

try {
    api.apiV1AdminHistoryUsersUserIdDelete(userId);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->apiV1AdminHistoryUsersUserIdDelete: $e\n');
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

final api = InoriApi().getHistoryApi();
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
    print('Exception when calling HistoryApi->apiV1AdminHistoryUsersUserIdGet: $e\n');
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

# **apiV1MeHistoryBatchDeletePost**
> BatchDeleteResult apiV1MeHistoryBatchDeletePost(batchDeleteRequest)

Batch-delete own play events by IDs (viewer)

Deletes only events owned by the authenticated viewer. IDs belonging to other users are silently skipped.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final BatchDeleteRequest batchDeleteRequest = ; // BatchDeleteRequest | 

try {
    final response = api.apiV1MeHistoryBatchDeletePost(batchDeleteRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->apiV1MeHistoryBatchDeletePost: $e\n');
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

# **apiV1MeHistoryEventIdDelete**
> apiV1MeHistoryEventIdDelete(eventId)

Delete a single play event by ID (viewer, own events only)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final String eventId = eventId_example; // String | The play event ID.

try {
    api.apiV1MeHistoryEventIdDelete(eventId);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->apiV1MeHistoryEventIdDelete: $e\n');
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

# **apiV1MeHistoryEventIdGet**
> PlayEvent apiV1MeHistoryEventIdGet(eventId)

Get a single play event by ID (viewer, own events only)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final String eventId = eventId_example; // String | The play event ID.

try {
    final response = api.apiV1MeHistoryEventIdGet(eventId);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->apiV1MeHistoryEventIdGet: $e\n');
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

# **apiV1MeHistoryEventIdPatch**
> PlayEvent apiV1MeHistoryEventIdPatch(eventId, updatePlayEventRequest)

Update a play event's playedAt timestamp (viewer, own events only)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final String eventId = eventId_example; // String | The play event ID.
final UpdatePlayEventRequest updatePlayEventRequest = ; // UpdatePlayEventRequest | Fields to update on the play event.

try {
    final response = api.apiV1MeHistoryEventIdPatch(eventId, updatePlayEventRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->apiV1MeHistoryEventIdPatch: $e\n');
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

# **apiV1MeHistoryStatsGet**
> UserHistoryStats apiV1MeHistoryStatsGet(since, until)

Get personal play history statistics for the authenticated user

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Count events at or after this time (RFC3339).
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Count events before this time (RFC3339, exclusive).

try {
    final response = api.apiV1MeHistoryStatsGet(since, until);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->apiV1MeHistoryStatsGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **since** | **DateTime**| Count events at or after this time (RFC3339). | [optional] 
 **until** | **DateTime**| Count events before this time (RFC3339, exclusive). | [optional] 

### Return type

[**UserHistoryStats**](UserHistoryStats.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1MeHistoryTopTracksGet**
> ApiV1MeHistoryTopTracksGet200Response apiV1MeHistoryTopTracksGet(limit, since, until)

Get top tracks for the authenticated user by play count

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final int limit = 56; // int | Maximum number of tracks to return (1–100, default 10).
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Count events at or after this time (RFC3339).
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Count events before this time (RFC3339, exclusive).

try {
    final response = api.apiV1MeHistoryTopTracksGet(limit, since, until);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->apiV1MeHistoryTopTracksGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **limit** | **int**| Maximum number of tracks to return (1–100, default 10). | [optional] 
 **since** | **DateTime**| Count events at or after this time (RFC3339). | [optional] 
 **until** | **DateTime**| Count events before this time (RFC3339, exclusive). | [optional] 

### Return type

[**ApiV1MeHistoryTopTracksGet200Response**](ApiV1MeHistoryTopTracksGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **clearPlayHistory**
> clearPlayHistory()

Clear play history

Deletes all play events for the calling user.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();

try {
    api.clearPlayHistory();
} on DioException catch (e) {
    print('Exception when calling HistoryApi->clearPlayHistory: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

void (empty response body)

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

final api = InoriApi().getHistoryApi();
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | 
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | 
final int limit = 56; // int | 

try {
    final response = api.getAdminHistorySummary(since, until, limit);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->getAdminHistorySummary: $e\n');
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

# **getAdminTrackHistorySummary**
> TrackHistorySummary getAdminTrackHistorySummary(trackId, since, until, limit)

Get track history summary

Returns combined aggregate stats and top-listeners for a specific track in a single response. Accepts optional `since`/`until` bounds and a `limit` for the top-listeners list.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final String trackId = trackId_example; // String | The track ID.
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Include only events at or after this time (RFC3339, inclusive).
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Include only events before this time (RFC3339, exclusive).
final int limit = 56; // int | Maximum number of top-listeners entries to return (1–100; default 10).

try {
    final response = api.getAdminTrackHistorySummary(trackId, since, until, limit);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->getAdminTrackHistorySummary: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trackId** | **String**| The track ID. | 
 **since** | **DateTime**| Include only events at or after this time (RFC3339, inclusive). | [optional] 
 **until** | **DateTime**| Include only events before this time (RFC3339, exclusive). | [optional] 
 **limit** | **int**| Maximum number of top-listeners entries to return (1–100; default 10). | [optional] [default to 10]

### Return type

[**TrackHistorySummary**](TrackHistorySummary.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAdminTrackTimeline**
> TimelineResult getAdminTrackTimeline(trackId, since, until, granularity)

Track play timeline

Returns play-event counts for a specific track grouped by time bucket (day, week, or month).

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final String trackId = trackId_example; // String | The track ID.
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Start of the timeline window (RFC3339, inclusive).
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | End of the timeline window (RFC3339, exclusive).
final String granularity = granularity_example; // String | Bucket size: day (default), week, or month.

try {
    final response = api.getAdminTrackTimeline(trackId, since, until, granularity);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->getAdminTrackTimeline: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trackId** | **String**| The track ID. | 
 **since** | **DateTime**| Start of the timeline window (RFC3339, inclusive). | 
 **until** | **DateTime**| End of the timeline window (RFC3339, exclusive). | 
 **granularity** | **String**| Bucket size: day (default), week, or month. | [optional] [default to 'day']

### Return type

[**TimelineResult**](TimelineResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAdminUserHistorySummary**
> UserHistorySummary getAdminUserHistorySummary(userId, since, until, limit)

Get user history summary

Returns combined aggregate stats and top-tracks for a specific user in a single response. Accepts optional `since`/`until` bounds and a `limit` for the top-tracks list.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final String userId = userId_example; // String | The user ID.
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Include only events at or after this time (RFC3339, inclusive).
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Include only events before this time (RFC3339, exclusive).
final int limit = 56; // int | Maximum number of top-tracks entries to return (1–100; default 10).

try {
    final response = api.getAdminUserHistorySummary(userId, since, until, limit);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->getAdminUserHistorySummary: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **String**| The user ID. | 
 **since** | **DateTime**| Include only events at or after this time (RFC3339, inclusive). | [optional] 
 **until** | **DateTime**| Include only events before this time (RFC3339, exclusive). | [optional] 
 **limit** | **int**| Maximum number of top-tracks entries to return (1–100; default 10). | [optional] [default to 10]

### Return type

[**UserHistorySummary**](UserHistorySummary.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAdminUserTimeline**
> TimelineResult getAdminUserTimeline(userId, since, until, granularity)

User play timeline

Returns play-event counts for a specific user grouped by time bucket (day, week, or month).

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final String userId = userId_example; // String | The user ID.
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Start of the timeline window (RFC3339, inclusive).
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | End of the timeline window (RFC3339, exclusive).
final String granularity = granularity_example; // String | Bucket size: day (default), week, or month.

try {
    final response = api.getAdminUserTimeline(userId, since, until, granularity);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->getAdminUserTimeline: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **String**| The user ID. | 
 **since** | **DateTime**| Start of the timeline window (RFC3339, inclusive). | 
 **until** | **DateTime**| End of the timeline window (RFC3339, exclusive). | 
 **granularity** | **String**| Bucket size: day (default), week, or month. | [optional] [default to 'day']

### Return type

[**TimelineResult**](TimelineResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getMyHistorySummary**
> GetMyHistorySummary200Response getMyHistorySummary(since, until, limit)

Get viewer history summary

Returns combined aggregate stats and top-tracks for the authenticated viewer in a single response. Accepts optional `since`/`until` bounds and a `limit` for the top-tracks list.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Include only events at or after this time (RFC3339, inclusive).
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Include only events before this time (RFC3339, exclusive).
final int limit = 56; // int | Maximum number of top-tracks entries to return (1–100; default 10).

try {
    final response = api.getMyHistorySummary(since, until, limit);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->getMyHistorySummary: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **since** | **DateTime**| Include only events at or after this time (RFC3339, inclusive). | [optional] 
 **until** | **DateTime**| Include only events before this time (RFC3339, exclusive). | [optional] 
 **limit** | **int**| Maximum number of top-tracks entries to return (1–100; default 10). | [optional] [default to 10]

### Return type

[**GetMyHistorySummary200Response**](GetMyHistorySummary200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getMyHistoryTimeline**
> TimelineResult getMyHistoryTimeline(since, until, granularity, trackId)

Get the authenticated user's play history grouped by time bucket

Returns the viewer's own play event counts grouped by day, week, or month within the specified time window. Both since and until are required.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Start of the time window (RFC3339, inclusive).
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | End of the time window (RFC3339, exclusive).
final String granularity = granularity_example; // String | Bucket size: day (default), week, or month.
final String trackId = trackId_example; // String | Optional: restrict to events for a specific track.

try {
    final response = api.getMyHistoryTimeline(since, until, granularity, trackId);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->getMyHistoryTimeline: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **since** | **DateTime**| Start of the time window (RFC3339, inclusive). | 
 **until** | **DateTime**| End of the time window (RFC3339, exclusive). | 
 **granularity** | **String**| Bucket size: day (default), week, or month. | [optional] [default to 'day']
 **trackId** | **String**| Optional: restrict to events for a specific track. | [optional] 

### Return type

[**TimelineResult**](TimelineResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getMyTrackStats**
> UserTrackStats getMyTrackStats(trackId, since, until)

Get play stats for a track

Returns the calling user's play statistics for a specific track: total plays, first played time, and last played time.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final String trackId = trackId_example; // String | The track ID.
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Include only events at or after this time (RFC3339, inclusive).
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Include only events before this time (RFC3339, exclusive).

try {
    final response = api.getMyTrackStats(trackId, since, until);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->getMyTrackStats: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trackId** | **String**| The track ID. | 
 **since** | **DateTime**| Include only events at or after this time (RFC3339, inclusive). | [optional] 
 **until** | **DateTime**| Include only events before this time (RFC3339, exclusive). | [optional] 

### Return type

[**UserTrackStats**](UserTrackStats.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getMyTrackSummary**
> MyTrackSummary getMyTrackSummary(trackId, since, until, limit)

Get viewer track summary

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final String trackId = trackId_example; // String | 
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | 
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | 
final int limit = 56; // int | 

try {
    final response = api.getMyTrackSummary(trackId, since, until, limit);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->getMyTrackSummary: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trackId** | **String**|  | 
 **since** | **DateTime**|  | [optional] 
 **until** | **DateTime**|  | [optional] 
 **limit** | **int**|  | [optional] [default to 10]

### Return type

[**MyTrackSummary**](MyTrackSummary.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getMyTrackTimeline**
> TimelineResult getMyTrackTimeline(trackId, since, until, granularity)

Get viewer track play timeline

Returns the calling user's play-event counts for a specific track grouped by time bucket (day/week/month). Both `since` and `until` are required.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final String trackId = trackId_example; // String | The track ID.
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Start of the time window (RFC3339, inclusive).
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | End of the time window (RFC3339, exclusive).
final String granularity = granularity_example; // String | Bucket size: day (default), week, or month.

try {
    final response = api.getMyTrackTimeline(trackId, since, until, granularity);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->getMyTrackTimeline: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trackId** | **String**| The track ID. | 
 **since** | **DateTime**| Start of the time window (RFC3339, inclusive). | 
 **until** | **DateTime**| End of the time window (RFC3339, exclusive). | 
 **granularity** | **String**| Bucket size: day (default), week, or month. | [optional] [default to 'day']

### Return type

[**TimelineResult**](TimelineResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listFavoriteTracks**
> FavoritesPage listFavoriteTracks(limit, offset)

List favorite tracks (paginated)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final int limit = 56; // int | 
final int offset = 56; // int | 

try {
    final response = api.listFavoriteTracks(limit, offset);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->listFavoriteTracks: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **limit** | **int**|  | [optional] 
 **offset** | **int**|  | [optional] 

### Return type

[**FavoritesPage**](FavoritesPage.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listMyTrackHistory**
> ListMyTrackHistory200Response listMyTrackHistory(trackId, limit, offset, order, since, until)

List play events for a track

Returns the calling user's play history for a specific track, newest first.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final String trackId = trackId_example; // String | The track ID.
final int limit = 56; // int | 
final int offset = 56; // int | 
final String order = order_example; // String | Sort direction. \"desc\" (default) returns newest first; \"asc\" returns oldest first.
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Include only events at or after this time (RFC3339, inclusive).
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Include only events before this time (RFC3339, exclusive).

try {
    final response = api.listMyTrackHistory(trackId, limit, offset, order, since, until);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->listMyTrackHistory: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trackId** | **String**| The track ID. | 
 **limit** | **int**|  | [optional] [default to 50]
 **offset** | **int**|  | [optional] [default to 0]
 **order** | **String**| Sort direction. \"desc\" (default) returns newest first; \"asc\" returns oldest first. | [optional] [default to 'desc']
 **since** | **DateTime**| Include only events at or after this time (RFC3339, inclusive). | [optional] 
 **until** | **DateTime**| Include only events before this time (RFC3339, exclusive). | [optional] 

### Return type

[**ListMyTrackHistory200Response**](ListMyTrackHistory200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listPlayEvents**
> PlayEventList listPlayEvents(trackId, limit, offset, order, since, until)

List play events

Returns the calling user's play history, newest first.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final String trackId = trackId_example; // String | Filter to a single track.
final int limit = 56; // int | 
final int offset = 56; // int | 
final String order = order_example; // String | Sort direction for results. \"desc\" (default) returns newest first; \"asc\" returns oldest first.
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Include only events at or after this time (RFC3339, inclusive).
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Include only events before this time (RFC3339, exclusive).

try {
    final response = api.listPlayEvents(trackId, limit, offset, order, since, until);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->listPlayEvents: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trackId** | **String**| Filter to a single track. | [optional] 
 **limit** | **int**|  | [optional] [default to 50]
 **offset** | **int**|  | [optional] [default to 0]
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

# **recordPlayEvent**
> PlayEvent recordPlayEvent(recordPlayEventRequest)

Record a play event

Records that the authenticated user played a track.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final RecordPlayEventRequest recordPlayEventRequest = ; // RecordPlayEventRequest | 

try {
    final response = api.recordPlayEvent(recordPlayEventRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->recordPlayEvent: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **recordPlayEventRequest** | [**RecordPlayEventRequest**](RecordPlayEventRequest.md)|  | 

### Return type

[**PlayEvent**](PlayEvent.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **removeFavoriteTrack**
> removeFavoriteTrack(trackId)

Remove a track from favorites (idempotent)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getHistoryApi();
final String trackId = trackId_example; // String | 

try {
    api.removeFavoriteTrack(trackId);
} on DioException catch (e) {
    print('Exception when calling HistoryApi->removeFavoriteTrack: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trackId** | **String**|  | 

### Return type

void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

