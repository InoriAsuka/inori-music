# inori_api.api.MeApi

## Load the API package
```dart
import 'package:inori_api/api.dart';
```

All URIs are relative to *http://127.0.0.1:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**getMyActiveSessions**](MeApi.md#getmyactivesessions) | **GET** /api/v1/me/sessions | List active sessions for the authenticated user
[**revokeMyOtherSessions**](MeApi.md#revokemyothersessions) | **POST** /api/v1/me/sessions/revoke-all | Revoke all sessions except the current one


# **getMyActiveSessions**
> GetAdminUserSessions200Response getMyActiveSessions()

List active sessions for the authenticated user

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getMeApi();

try {
    final response = api.getMyActiveSessions();
    print(response);
} on DioException catch (e) {
    print('Exception when calling MeApi->getMyActiveSessions: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**GetAdminUserSessions200Response**](GetAdminUserSessions200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **revokeMyOtherSessions**
> DeleteAdminUserSessions200Response revokeMyOtherSessions()

Revoke all sessions except the current one

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getMeApi();

try {
    final response = api.revokeMyOtherSessions();
    print(response);
} on DioException catch (e) {
    print('Exception when calling MeApi->revokeMyOtherSessions: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**DeleteAdminUserSessions200Response**](DeleteAdminUserSessions200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

