# inori_api.api.ViewerApi

## Load the API package
```dart
import 'package:inori_api/api.dart';
```

All URIs are relative to *http://127.0.0.1:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**apiV1CatalogTracksIdStreamGet**](ViewerApi.md#apiv1catalogtracksidstreamget) | **GET** /api/v1/catalog/tracks/{id}/stream | Stream track audio
[**getMyTrackSummary**](ViewerApi.md#getmytracksummary) | **GET** /api/v1/me/history/tracks/{trackId}/summary | Get viewer track summary


# **apiV1CatalogTracksIdStreamGet**
> apiV1CatalogTracksIdStreamGet(id, token)

Stream track audio

Proxies audio bytes with HTTP 206 Range support for filesystem-based storage backends. Authenticates via Bearer token in Authorization header or ?token= query parameter.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getViewerApi();
final String id = id_example; // String | Catalog entity identifier
final String token = token_example; // String | Viewer bearer token (fallback when Authorization header cannot be set, e.g. <audio> src)

try {
    api.apiV1CatalogTracksIdStreamGet(id, token);
} on DioException catch (e) {
    print('Exception when calling ViewerApi->apiV1CatalogTracksIdStreamGet: $e\n');
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

# **getMyTrackSummary**
> MyTrackSummary getMyTrackSummary(trackId, since, until, limit)

Get viewer track summary

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getViewerApi();
final String trackId = trackId_example; // String | 
final DateTime since = 2013-10-20T19:20:30+01:00; // DateTime | 
final DateTime until = 2013-10-20T19:20:30+01:00; // DateTime | 
final int limit = 56; // int | 

try {
    final response = api.getMyTrackSummary(trackId, since, until, limit);
    print(response);
} on DioException catch (e) {
    print('Exception when calling ViewerApi->getMyTrackSummary: $e\n');
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

