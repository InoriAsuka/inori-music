# inori_api.api.UserPlaylistsApi

## Load the API package
```dart
import 'package:inori_api/api.dart';
```

All URIs are relative to *http://127.0.0.1:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**addUserPlaylistTrack**](UserPlaylistsApi.md#adduserplaylisttrack) | **POST** /api/v1/me/playlists/{id}/tracks | Add track to user playlist
[**createUserPlaylist**](UserPlaylistsApi.md#createuserplaylist) | **POST** /api/v1/me/playlists | Create user playlist
[**deleteUserPlaylist**](UserPlaylistsApi.md#deleteuserplaylist) | **DELETE** /api/v1/me/playlists/{id} | Delete user playlist
[**getUserPlaylist**](UserPlaylistsApi.md#getuserplaylist) | **GET** /api/v1/me/playlists/{id} | Get user playlist
[**getUserPlaylistTracks**](UserPlaylistsApi.md#getuserplaylisttracks) | **GET** /api/v1/me/playlists/{id}/tracks | Get user playlist tracks
[**listUserPlaylists**](UserPlaylistsApi.md#listuserplaylists) | **GET** /api/v1/me/playlists | List user playlists
[**removeUserPlaylistTrack**](UserPlaylistsApi.md#removeuserplaylisttrack) | **DELETE** /api/v1/me/playlists/{id}/tracks/{trackId} | Remove track from user playlist
[**setUserPlaylistTracks**](UserPlaylistsApi.md#setuserplaylisttracks) | **PUT** /api/v1/me/playlists/{id}/tracks | Replace all tracks in user playlist
[**updateUserPlaylist**](UserPlaylistsApi.md#updateuserplaylist) | **PATCH** /api/v1/me/playlists/{id} | Update user playlist metadata


# **addUserPlaylistTrack**
> UserPlaylist addUserPlaylistTrack(id, addUserPlaylistTrackRequest)

Add track to user playlist

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getUserPlaylistsApi();
final String id = id_example; // String | Catalog entity identifier
final AddUserPlaylistTrackRequest addUserPlaylistTrackRequest = ; // AddUserPlaylistTrackRequest | 

try {
    final response = api.addUserPlaylistTrack(id, addUserPlaylistTrackRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling UserPlaylistsApi->addUserPlaylistTrack: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **addUserPlaylistTrackRequest** | [**AddUserPlaylistTrackRequest**](AddUserPlaylistTrackRequest.md)|  | 

### Return type

[**UserPlaylist**](UserPlaylist.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **createUserPlaylist**
> UserPlaylist createUserPlaylist(createUserPlaylistRequest)

Create user playlist

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getUserPlaylistsApi();
final CreateUserPlaylistRequest createUserPlaylistRequest = ; // CreateUserPlaylistRequest | 

try {
    final response = api.createUserPlaylist(createUserPlaylistRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling UserPlaylistsApi->createUserPlaylist: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createUserPlaylistRequest** | [**CreateUserPlaylistRequest**](CreateUserPlaylistRequest.md)|  | 

### Return type

[**UserPlaylist**](UserPlaylist.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **deleteUserPlaylist**
> deleteUserPlaylist(id)

Delete user playlist

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getUserPlaylistsApi();
final String id = id_example; // String | Catalog entity identifier

try {
    api.deleteUserPlaylist(id);
} on DioException catch (e) {
    print('Exception when calling UserPlaylistsApi->deleteUserPlaylist: $e\n');
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
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getUserPlaylist**
> UserPlaylist getUserPlaylist(id)

Get user playlist

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getUserPlaylistsApi();
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.getUserPlaylist(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling UserPlaylistsApi->getUserPlaylist: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 

### Return type

[**UserPlaylist**](UserPlaylist.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getUserPlaylistTracks**
> GetUserPlaylistTracks200Response getUserPlaylistTracks(id)

Get user playlist tracks

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getUserPlaylistsApi();
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.getUserPlaylistTracks(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling UserPlaylistsApi->getUserPlaylistTracks: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 

### Return type

[**GetUserPlaylistTracks200Response**](GetUserPlaylistTracks200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **listUserPlaylists**
> ListUserPlaylists200Response listUserPlaylists()

List user playlists

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getUserPlaylistsApi();

try {
    final response = api.listUserPlaylists();
    print(response);
} on DioException catch (e) {
    print('Exception when calling UserPlaylistsApi->listUserPlaylists: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**ListUserPlaylists200Response**](ListUserPlaylists200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **removeUserPlaylistTrack**
> UserPlaylist removeUserPlaylistTrack(trackId, id)

Remove track from user playlist

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getUserPlaylistsApi();
final String trackId = trackId_example; // String | 
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.removeUserPlaylistTrack(trackId, id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling UserPlaylistsApi->removeUserPlaylistTrack: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **trackId** | **String**|  | 
 **id** | **String**| Catalog entity identifier | 

### Return type

[**UserPlaylist**](UserPlaylist.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **setUserPlaylistTracks**
> UserPlaylist setUserPlaylistTracks(id, setUserPlaylistTracksRequest)

Replace all tracks in user playlist

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getUserPlaylistsApi();
final String id = id_example; // String | Catalog entity identifier
final SetUserPlaylistTracksRequest setUserPlaylistTracksRequest = ; // SetUserPlaylistTracksRequest | 

try {
    final response = api.setUserPlaylistTracks(id, setUserPlaylistTracksRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling UserPlaylistsApi->setUserPlaylistTracks: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **setUserPlaylistTracksRequest** | [**SetUserPlaylistTracksRequest**](SetUserPlaylistTracksRequest.md)|  | 

### Return type

[**UserPlaylist**](UserPlaylist.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **updateUserPlaylist**
> UserPlaylist updateUserPlaylist(id, updateUserPlaylistRequest)

Update user playlist metadata

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getUserPlaylistsApi();
final String id = id_example; // String | Catalog entity identifier
final UpdateUserPlaylistRequest updateUserPlaylistRequest = ; // UpdateUserPlaylistRequest | 

try {
    final response = api.updateUserPlaylist(id, updateUserPlaylistRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling UserPlaylistsApi->updateUserPlaylist: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **updateUserPlaylistRequest** | [**UpdateUserPlaylistRequest**](UpdateUserPlaylistRequest.md)|  | 

### Return type

[**UserPlaylist**](UserPlaylist.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

