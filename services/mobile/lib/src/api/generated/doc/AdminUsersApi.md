# inori_api.api.AdminUsersApi

## Load the API package
```dart
import 'package:inori_api/api.dart';
```

All URIs are relative to *http://127.0.0.1:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**deleteAdminUserSessions**](AdminUsersApi.md#deleteadminusersessions) | **DELETE** /api/v1/admin/users/{id}/sessions | Revoke all active sessions for a user
[**getAdminUserSessions**](AdminUsersApi.md#getadminusersessions) | **GET** /api/v1/admin/users/{id}/sessions | List active sessions for a user


# **deleteAdminUserSessions**
> DeleteAdminUserSessions200Response deleteAdminUserSessions(id)

Revoke all active sessions for a user

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminUsersApi();
final String id = id_example; // String | User identifier

try {
    final response = api.deleteAdminUserSessions(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminUsersApi->deleteAdminUserSessions: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| User identifier | 

### Return type

[**DeleteAdminUserSessions200Response**](DeleteAdminUserSessions200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getAdminUserSessions**
> GetAdminUserSessions200Response getAdminUserSessions(id)

List active sessions for a user

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getAdminUsersApi();
final String id = id_example; // String | User identifier

try {
    final response = api.getAdminUserSessions(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling AdminUsersApi->getAdminUserSessions: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| User identifier | 

### Return type

[**GetAdminUserSessions200Response**](GetAdminUserSessions200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

