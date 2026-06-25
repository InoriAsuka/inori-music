# inori_api.api.AdminHistoryApi

## Load the API package
```dart
import 'package:inori_api/api.dart';
```

All URIs are relative to *http://127.0.0.1:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**getAdminHistoryStats**](AdminHistoryApi.md#getadminhistorystats) | **GET** /api/v1/admin/history/stats | Get history aggregate stats
[**getAdminHistoryTimeline**](AdminHistoryApi.md#getadminhistorytimeline) | **GET** /api/v1/admin/history/timeline | Get play history grouped by time bucket (admin)
[**getAdminTopTracks**](AdminHistoryApi.md#getadmintoptracks) | **GET** /api/v1/admin/history/top-tracks | Get most-played tracks
[**getAdminTopUsers**](AdminHistoryApi.md#getadmintopusers) | **GET** /api/v1/admin/history/top-users | Get most-active users
[**getAdminTrackStats**](AdminHistoryApi.md#getadmintrackstats) | **GET** /api/v1/admin/history/tracks/{trackId}/stats | Get play history stats for a specific track (admin)
[**getAdminTrackTopListeners**](AdminHistoryApi.md#getadmintracktoplisteners) | **GET** /api/v1/admin/history/tracks/{trackId}/top-listeners | Get top listeners for a specific track (admin)
[**getAdminUserStats**](AdminHistoryApi.md#getadminuserstats) | **GET** /api/v1/admin/history/users/{userId}/stats | Get play history stats for a specific user (admin)
[**getAdminUserTopTracks**](AdminHistoryApi.md#getadminusertoptracks) | **GET** /api/v1/admin/history/users/{userId}/top-tracks | Get top tracks for a specific user (admin)


# **getAdminHistoryStats**
> HistoryStats getAdminHistoryStats(since, until)

Get history aggregate stats

Returns system-wide playback aggregate counts (admin only).

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminHistoryApi();
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Restrict results to events on or after this RFC3339 timestamp.
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Restrict results to events before this RFC3339 timestamp (exclusive).

try {
    final response = api.getAdminHistoryStats(since, until);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminHistoryApi->getAdminHistoryStats: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **since** | **DateTime**| Restrict results to events on or after this RFC3339 timestamp. | [optional] 
 **until** | **DateTime**| Restrict results to events before this RFC3339 timestamp (exclusive). | [optional] 

### Return type

[**HistoryStats**](HistoryStats.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAdminHistoryTimeline**
> TimelineResult getAdminHistoryTimeline(since, until, granularity, userId, trackId)

Get play history grouped by time bucket (admin)

Returns play event counts grouped by day, week, or month within the specified time window. Both since and until are required. Admin only.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminHistoryApi();
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Start of the time window (RFC3339, inclusive).
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | End of the time window (RFC3339, exclusive).
final String granularity = granularity_example; // String | Bucket size: day (default), week, or month.
final String userId = userId_example; // String | Optional: restrict to a specific user's events.
final String trackId = trackId_example; // String | Optional: restrict to events for a specific track.

try {
    final response = api.getAdminHistoryTimeline(since, until, granularity, userId, trackId);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminHistoryApi->getAdminHistoryTimeline: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **since** | **DateTime**| Start of the time window (RFC3339, inclusive). | 
 **until** | **DateTime**| End of the time window (RFC3339, exclusive). | 
 **granularity** | **String**| Bucket size: day (default), week, or month. | [optional] [default to 'day']
 **userId** | **String**| Optional: restrict to a specific user's events. | [optional] 
 **trackId** | **String**| Optional: restrict to events for a specific track. | [optional] 

### Return type

[**TimelineResult**](TimelineResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAdminTopTracks**
> TopTracksResult getAdminTopTracks(limit, since, until)

Get most-played tracks

Returns tracks with the highest total play count across all users (admin only).

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminHistoryApi();
final int limit = 56; // int | Maximum number of results (default 10, max 100).
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Restrict results to events on or after this RFC3339 timestamp.
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Restrict results to events before this RFC3339 timestamp (exclusive).

try {
    final response = api.getAdminTopTracks(limit, since, until);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminHistoryApi->getAdminTopTracks: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **limit** | **int**| Maximum number of results (default 10, max 100). | [optional] [default to 10]
 **since** | **DateTime**| Restrict results to events on or after this RFC3339 timestamp. | [optional] 
 **until** | **DateTime**| Restrict results to events before this RFC3339 timestamp (exclusive). | [optional] 

### Return type

[**TopTracksResult**](TopTracksResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAdminTopUsers**
> TopUsersResult getAdminTopUsers(limit, since, until)

Get most-active users

Returns users with the highest total play event count (admin only).

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminHistoryApi();
final int limit = 56; // int | Maximum number of results (default 10, max 100).
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Restrict results to events on or after this RFC3339 timestamp.
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Restrict results to events before this RFC3339 timestamp (exclusive).

try {
    final response = api.getAdminTopUsers(limit, since, until);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminHistoryApi->getAdminTopUsers: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **limit** | **int**| Maximum number of results (default 10, max 100). | [optional] [default to 10]
 **since** | **DateTime**| Restrict results to events on or after this RFC3339 timestamp. | [optional] 
 **until** | **DateTime**| Restrict results to events before this RFC3339 timestamp (exclusive). | [optional] 

### Return type

[**TopUsersResult**](TopUsersResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAdminTrackStats**
> TrackHistoryStats getAdminTrackStats(trackId, since, until)

Get play history stats for a specific track (admin)

Returns aggregate play counts (total events, unique listeners) for any track. Admin only.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminHistoryApi();
final String trackId = trackId_example; // String | ID of the track whose stats to retrieve.
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Restrict stats to events on or after this RFC3339 timestamp.
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Restrict stats to events before this RFC3339 timestamp (exclusive).

try {
    final response = api.getAdminTrackStats(trackId, since, until);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminHistoryApi->getAdminTrackStats: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trackId** | **String**| ID of the track whose stats to retrieve. | 
 **since** | **DateTime**| Restrict stats to events on or after this RFC3339 timestamp. | [optional] 
 **until** | **DateTime**| Restrict stats to events before this RFC3339 timestamp (exclusive). | [optional] 

### Return type

[**TrackHistoryStats**](TrackHistoryStats.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAdminTrackTopListeners**
> TopUsersResult getAdminTrackTopListeners(trackId, limit, since, until)

Get top listeners for a specific track (admin)

Returns the users who have played a track the most. Admin only.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminHistoryApi();
final String trackId = trackId_example; // String | ID of the track whose top listeners to retrieve.
final int limit = 56; // int | Maximum number of results (default 10, max 100).
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Restrict results to events on or after this RFC3339 timestamp.
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Restrict results to events before this RFC3339 timestamp (exclusive).

try {
    final response = api.getAdminTrackTopListeners(trackId, limit, since, until);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminHistoryApi->getAdminTrackTopListeners: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trackId** | **String**| ID of the track whose top listeners to retrieve. | 
 **limit** | **int**| Maximum number of results (default 10, max 100). | [optional] [default to 10]
 **since** | **DateTime**| Restrict results to events on or after this RFC3339 timestamp. | [optional] 
 **until** | **DateTime**| Restrict results to events before this RFC3339 timestamp (exclusive). | [optional] 

### Return type

[**TopUsersResult**](TopUsersResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAdminUserStats**
> UserHistoryStats getAdminUserStats(userId, since, until)

Get play history stats for a specific user (admin)

Returns aggregate play counts (total events, unique tracks) for any user. Admin only.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminHistoryApi();
final String userId = userId_example; // String | ID of the user whose stats to retrieve.
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Restrict stats to events on or after this RFC3339 timestamp.
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Restrict stats to events before this RFC3339 timestamp (exclusive).

try {
    final response = api.getAdminUserStats(userId, since, until);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminHistoryApi->getAdminUserStats: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **String**| ID of the user whose stats to retrieve. | 
 **since** | **DateTime**| Restrict stats to events on or after this RFC3339 timestamp. | [optional] 
 **until** | **DateTime**| Restrict stats to events before this RFC3339 timestamp (exclusive). | [optional] 

### Return type

[**UserHistoryStats**](UserHistoryStats.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAdminUserTopTracks**
> TopTracksResult getAdminUserTopTracks(userId, limit, since, until)

Get top tracks for a specific user (admin)

Returns the tracks most frequently played by any user. Admin only.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminHistoryApi();
final String userId = userId_example; // String | ID of the user whose top tracks to retrieve.
final int limit = 56; // int | Maximum number of results (default 10, max 100).
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | Restrict results to events on or after this RFC3339 timestamp.
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | Restrict results to events before this RFC3339 timestamp (exclusive).

try {
    final response = api.getAdminUserTopTracks(userId, limit, since, until);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminHistoryApi->getAdminUserTopTracks: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **userId** | **String**| ID of the user whose top tracks to retrieve. | 
 **limit** | **int**| Maximum number of results (default 10, max 100). | [optional] [default to 10]
 **since** | **DateTime**| Restrict results to events on or after this RFC3339 timestamp. | [optional] 
 **until** | **DateTime**| Restrict results to events before this RFC3339 timestamp (exclusive). | [optional] 

### Return type

[**TopTracksResult**](TopTracksResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

