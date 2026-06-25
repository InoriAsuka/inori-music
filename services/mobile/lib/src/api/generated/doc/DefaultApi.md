# inori_api.api.DefaultApi

## Load the API package
```dart
import 'package:inori_api/api.dart';
```

All URIs are relative to *http://127.0.0.1:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**apiV1AdminCatalogAlbumsGet**](DefaultApi.md#apiv1admincatalogalbumsget) | **GET** /api/v1/admin/catalog/albums | List albums
[**apiV1AdminCatalogAlbumsIdDelete**](DefaultApi.md#apiv1admincatalogalbumsiddelete) | **DELETE** /api/v1/admin/catalog/albums/{id} | Delete album
[**apiV1AdminCatalogAlbumsIdGet**](DefaultApi.md#apiv1admincatalogalbumsidget) | **GET** /api/v1/admin/catalog/albums/{id} | Get album
[**apiV1AdminCatalogAlbumsIdPatch**](DefaultApi.md#apiv1admincatalogalbumsidpatch) | **PATCH** /api/v1/admin/catalog/albums/{id} | Update album metadata
[**apiV1AdminCatalogAlbumsPost**](DefaultApi.md#apiv1admincatalogalbumspost) | **POST** /api/v1/admin/catalog/albums | Create album
[**apiV1AdminCatalogArtistsGet**](DefaultApi.md#apiv1admincatalogartistsget) | **GET** /api/v1/admin/catalog/artists | List artists
[**apiV1AdminCatalogArtistsIdDelete**](DefaultApi.md#apiv1admincatalogartistsiddelete) | **DELETE** /api/v1/admin/catalog/artists/{id} | Delete artist
[**apiV1AdminCatalogArtistsIdGet**](DefaultApi.md#apiv1admincatalogartistsidget) | **GET** /api/v1/admin/catalog/artists/{id} | Get artist
[**apiV1AdminCatalogArtistsIdPatch**](DefaultApi.md#apiv1admincatalogartistsidpatch) | **PATCH** /api/v1/admin/catalog/artists/{id} | Update artist metadata
[**apiV1AdminCatalogArtistsPost**](DefaultApi.md#apiv1admincatalogartistspost) | **POST** /api/v1/admin/catalog/artists | Create artist
[**apiV1AdminCatalogBatchImportPost**](DefaultApi.md#apiv1admincatalogbatchimportpost) | **POST** /api/v1/admin/catalog/batch-import | Batch import media objects as catalog tracks
[**apiV1AdminCatalogImportPost**](DefaultApi.md#apiv1admincatalogimportpost) | **POST** /api/v1/admin/catalog/import | Import media object as a catalog track
[**apiV1AdminCatalogTracksGet**](DefaultApi.md#apiv1admincatalogtracksget) | **GET** /api/v1/admin/catalog/tracks | List tracks
[**apiV1AdminCatalogTracksIdDelete**](DefaultApi.md#apiv1admincatalogtracksiddelete) | **DELETE** /api/v1/admin/catalog/tracks/{id} | Delete track
[**apiV1AdminCatalogTracksIdGet**](DefaultApi.md#apiv1admincatalogtracksidget) | **GET** /api/v1/admin/catalog/tracks/{id} | Get track
[**apiV1AdminCatalogTracksIdPatch**](DefaultApi.md#apiv1admincatalogtracksidpatch) | **PATCH** /api/v1/admin/catalog/tracks/{id} | Update track metadata
[**apiV1AdminCatalogTracksIdRelinkPost**](DefaultApi.md#apiv1admincatalogtracksidrelinkpost) | **POST** /api/v1/admin/catalog/tracks/{id}/relink | Relink track to a different media object
[**apiV1AdminCatalogTracksPost**](DefaultApi.md#apiv1admincatalogtrackspost) | **POST** /api/v1/admin/catalog/tracks | Create track
[**apiV1AdminMediaObjectsDuplicatesGet**](DefaultApi.md#apiv1adminmediaobjectsduplicatesget) | **GET** /api/v1/admin/media/objects/duplicates | Find duplicate media objects by content hash.
[**apiV1AdminMediaObjectsGet**](DefaultApi.md#apiv1adminmediaobjectsget) | **GET** /api/v1/admin/media/objects | List media objects by metadata filter with offset pagination.
[**apiV1AdminMediaObjectsIdGet**](DefaultApi.md#apiv1adminmediaobjectsidget) | **GET** /api/v1/admin/media/objects/{id} | Get a media object by ID.
[**apiV1AdminMediaObjectsIdLifecyclePost**](DefaultApi.md#apiv1adminmediaobjectsidlifecyclepost) | **POST** /api/v1/admin/media/objects/{id}/lifecycle | Update a media object lifecycle state.
[**apiV1AdminMediaObjectsIdTimelineGet**](DefaultApi.md#apiv1adminmediaobjectsidtimelineget) | **GET** /api/v1/admin/media/objects/{id}/timeline | Get the retained metadata timeline for a media object.
[**apiV1AdminMediaObjectsIdVerifyPost**](DefaultApi.md#apiv1adminmediaobjectsidverifypost) | **POST** /api/v1/admin/media/objects/{id}/verify | Verify a media object metadata reference against stored bytes.
[**apiV1AdminMediaObjectsLifecyclePost**](DefaultApi.md#apiv1adminmediaobjectslifecyclepost) | **POST** /api/v1/admin/media/objects/lifecycle | Bulk update media object lifecycle metadata.
[**apiV1AdminMediaObjectsPost**](DefaultApi.md#apiv1adminmediaobjectspost) | **POST** /api/v1/admin/media/objects | Register a media object metadata reference.
[**apiV1AdminMediaObjectsStatsGet**](DefaultApi.md#apiv1adminmediaobjectsstatsget) | **GET** /api/v1/admin/media/objects/stats | Get media object metadata statistics.
[**apiV1AdminMediaObjectsVerifyPost**](DefaultApi.md#apiv1adminmediaobjectsverifypost) | **POST** /api/v1/admin/media/objects/verify | Batch verify media object metadata references by backend ID or content hash.
[**apiV1AdminStorageBackendsGet**](DefaultApi.md#apiv1adminstoragebackendsget) | **GET** /api/v1/admin/storage/backends | List configured storage backends.
[**apiV1AdminStorageBackendsIdCapacityGet**](DefaultApi.md#apiv1adminstoragebackendsidcapacityget) | **GET** /api/v1/admin/storage/backends/{id}/capacity | Read and record backend capacity where supported.
[**apiV1AdminStorageBackendsIdDefaultPost**](DefaultApi.md#apiv1adminstoragebackendsiddefaultpost) | **POST** /api/v1/admin/storage/backends/{id}/default | Select a backend as default.
[**apiV1AdminStorageBackendsIdDisablePost**](DefaultApi.md#apiv1adminstoragebackendsiddisablepost) | **POST** /api/v1/admin/storage/backends/{id}/disable | Disable a non-default backend.
[**apiV1AdminStorageBackendsIdHealthGet**](DefaultApi.md#apiv1adminstoragebackendsidhealthget) | **GET** /api/v1/admin/storage/backends/{id}/health | Read latest recorded backend health.
[**apiV1AdminStorageBackendsIdProbePost**](DefaultApi.md#apiv1adminstoragebackendsidprobepost) | **POST** /api/v1/admin/storage/backends/{id}/probe | Run an on-demand backend probe.
[**apiV1AdminStorageBackendsPost**](DefaultApi.md#apiv1adminstoragebackendspost) | **POST** /api/v1/admin/storage/backends | Validate and register a storage backend.
[**apiV1AdminStorageBackendsRefreshPost**](DefaultApi.md#apiv1adminstoragebackendsrefreshpost) | **POST** /api/v1/admin/storage/backends/refresh | Refresh all enabled backend health and supported capacity state.
[**apiV1AdminStorageBackendsValidatePost**](DefaultApi.md#apiv1adminstoragebackendsvalidatepost) | **POST** /api/v1/admin/storage/backends/validate | Validate a backend candidate without persisting it.
[**apiV1AdminUsersGet**](DefaultApi.md#apiv1adminusersget) | **GET** /api/v1/admin/users | List users.
[**apiV1AdminUsersIdChangePasswordPost**](DefaultApi.md#apiv1adminusersidchangepasswordpost) | **POST** /api/v1/admin/users/{id}/change-password | Force-reset a user&#39;s password without requiring the current password.
[**apiV1AdminUsersIdDelete**](DefaultApi.md#apiv1adminusersiddelete) | **DELETE** /api/v1/admin/users/{id} | Delete a user.
[**apiV1AdminUsersIdDisablePost**](DefaultApi.md#apiv1adminusersiddisablepost) | **POST** /api/v1/admin/users/{id}/disable | Disable a user.
[**apiV1AdminUsersIdEnablePost**](DefaultApi.md#apiv1adminusersidenablepost) | **POST** /api/v1/admin/users/{id}/enable | Enable a previously disabled user.
[**apiV1AdminUsersIdGet**](DefaultApi.md#apiv1adminusersidget) | **GET** /api/v1/admin/users/{id} | Get a user by ID.
[**apiV1AdminUsersIdPatch**](DefaultApi.md#apiv1adminusersidpatch) | **PATCH** /api/v1/admin/users/{id} | Partially update a user&#39;s role or username.
[**apiV1AdminUsersPost**](DefaultApi.md#apiv1adminuserspost) | **POST** /api/v1/admin/users | Create a user.
[**apiV1AuthLoginPost**](DefaultApi.md#apiv1authloginpost) | **POST** /api/v1/auth/login | Create a session token from username and password.
[**apiV1AuthLogoutPost**](DefaultApi.md#apiv1authlogoutpost) | **POST** /api/v1/auth/logout | Revoke the current session token.
[**apiV1MeChangePasswordPost**](DefaultApi.md#apiv1mechangepasswordpost) | **POST** /api/v1/me/change-password | Change the authenticated viewer&#39;s password.
[**apiV1MeGet**](DefaultApi.md#apiv1meget) | **GET** /api/v1/me | Get the authenticated viewer&#39;s profile.
[**apiV1MeSessionsRevokeAllDevicesPost**](DefaultApi.md#apiv1mesessionsrevokealldevicespost) | **POST** /api/v1/me/sessions/revoke-all-devices | Revoke all sessions for the authenticated user, including the current session.
[**deleteStorageBackend**](DefaultApi.md#deletestoragebackend) | **DELETE** /api/v1/admin/storage/backends/{id} | Delete a storage backend
[**enableStorageBackend**](DefaultApi.md#enablestoragebackend) | **POST** /api/v1/admin/storage/backends/{id}/enable | Enable a storage backend
[**getStorageBackend**](DefaultApi.md#getstoragebackend) | **GET** /api/v1/admin/storage/backends/{id} | Get a storage backend by ID
[**healthzGet**](DefaultApi.md#healthzget) | **GET** /healthz | Return process health.
[**metricsGet**](DefaultApi.md#metricsget) | **GET** /metrics | Get public Prometheus-compatible process metrics.
[**patchMediaObject**](DefaultApi.md#patchmediaobject) | **PATCH** /api/v1/admin/media/objects/{id} | Update correctable metadata on a media object (assetKind, mimeType)
[**patchStorageBackend**](DefaultApi.md#patchstoragebackend) | **PATCH** /api/v1/admin/storage/backends/{id} | Update a storage backend&#39;s display name or priority
[**readyzGet**](DefaultApi.md#readyzget) | **GET** /readyz | Get public readiness status for this API process.
[**versionzGet**](DefaultApi.md#versionzget) | **GET** /versionz | Get public build metadata for this API process.


# **apiV1AdminCatalogAlbumsGet**
> ApiV1AdminCatalogAlbumsGet200Response apiV1AdminCatalogAlbumsGet(artistId, limit, offset, sortBy, sortOrder, releaseYearMin, releaseYearMax)

List albums

List albums. sortBy: title (default), sortTitle, releaseYear, createdAt, updatedAt.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String artistId = artistId_example; // String | 
final int limit = 56; // int | Maximum number of items to return (1–500, default 50).
final int offset = 56; // int | Zero-based index of the first item to return.
final String sortBy = sortBy_example; // String | Field to sort by. See endpoint description for valid values.
final String sortOrder = sortOrder_example; // String | Sort direction.
final int releaseYearMin = 56; // int | Include only albums with release year >= this value
final int releaseYearMax = 56; // int | Include only albums with release year <= this value

try {
    final response = api.apiV1AdminCatalogAlbumsGet(artistId, limit, offset, sortBy, sortOrder, releaseYearMin, releaseYearMax);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogAlbumsGet: $e\n');
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

# **apiV1AdminCatalogAlbumsIdDelete**
> apiV1AdminCatalogAlbumsIdDelete(id)

Delete album

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Catalog entity identifier

try {
    api.apiV1AdminCatalogAlbumsIdDelete(id);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogAlbumsIdDelete: $e\n');
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

# **apiV1AdminCatalogAlbumsIdGet**
> CatalogAlbum apiV1AdminCatalogAlbumsIdGet(id)

Get album

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.apiV1AdminCatalogAlbumsIdGet(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogAlbumsIdGet: $e\n');
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

# **apiV1AdminCatalogAlbumsIdPatch**
> CatalogAlbum apiV1AdminCatalogAlbumsIdPatch(id, catalogUpdateAlbumRequest)

Update album metadata

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Catalog entity identifier
final CatalogUpdateAlbumRequest catalogUpdateAlbumRequest = ; // CatalogUpdateAlbumRequest | 

try {
    final response = api.apiV1AdminCatalogAlbumsIdPatch(id, catalogUpdateAlbumRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogAlbumsIdPatch: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **catalogUpdateAlbumRequest** | [**CatalogUpdateAlbumRequest**](CatalogUpdateAlbumRequest.md)|  | 

### Return type

[**CatalogAlbum**](CatalogAlbum.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogAlbumsPost**
> CatalogAlbum apiV1AdminCatalogAlbumsPost()

Create album

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();

try {
    final response = api.apiV1AdminCatalogAlbumsPost();
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogAlbumsPost: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**CatalogAlbum**](CatalogAlbum.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogArtistsGet**
> ApiV1AdminCatalogArtistsGet200Response apiV1AdminCatalogArtistsGet(limit, offset, sortBy, sortOrder)

List artists

List artists. sortBy: name (default), sortName, createdAt, updatedAt.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final int limit = 56; // int | Maximum number of items to return (1–500, default 50).
final int offset = 56; // int | Zero-based index of the first item to return.
final String sortBy = sortBy_example; // String | Field to sort by. See endpoint description for valid values.
final String sortOrder = sortOrder_example; // String | Sort direction.

try {
    final response = api.apiV1AdminCatalogArtistsGet(limit, offset, sortBy, sortOrder);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogArtistsGet: $e\n');
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

# **apiV1AdminCatalogArtistsIdDelete**
> apiV1AdminCatalogArtistsIdDelete(id)

Delete artist

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Catalog entity identifier

try {
    api.apiV1AdminCatalogArtistsIdDelete(id);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogArtistsIdDelete: $e\n');
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

# **apiV1AdminCatalogArtistsIdGet**
> CatalogArtist apiV1AdminCatalogArtistsIdGet(id)

Get artist

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.apiV1AdminCatalogArtistsIdGet(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogArtistsIdGet: $e\n');
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

# **apiV1AdminCatalogArtistsIdPatch**
> CatalogArtist apiV1AdminCatalogArtistsIdPatch(id, catalogUpdateArtistRequest)

Update artist metadata

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Catalog entity identifier
final CatalogUpdateArtistRequest catalogUpdateArtistRequest = ; // CatalogUpdateArtistRequest | 

try {
    final response = api.apiV1AdminCatalogArtistsIdPatch(id, catalogUpdateArtistRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogArtistsIdPatch: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **catalogUpdateArtistRequest** | [**CatalogUpdateArtistRequest**](CatalogUpdateArtistRequest.md)|  | 

### Return type

[**CatalogArtist**](CatalogArtist.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogArtistsPost**
> CatalogArtist apiV1AdminCatalogArtistsPost()

Create artist

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();

try {
    final response = api.apiV1AdminCatalogArtistsPost();
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogArtistsPost: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**CatalogArtist**](CatalogArtist.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogBatchImportPost**
> CatalogBatchImportResult apiV1AdminCatalogBatchImportPost(catalogBatchImportRequest)

Batch import media objects as catalog tracks

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final CatalogBatchImportRequest catalogBatchImportRequest = ; // CatalogBatchImportRequest | 

try {
    final response = api.apiV1AdminCatalogBatchImportPost(catalogBatchImportRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogBatchImportPost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **catalogBatchImportRequest** | [**CatalogBatchImportRequest**](CatalogBatchImportRequest.md)|  | 

### Return type

[**CatalogBatchImportResult**](CatalogBatchImportResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogImportPost**
> CatalogTrack apiV1AdminCatalogImportPost(catalogImportRequest)

Import media object as a catalog track

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final CatalogImportRequest catalogImportRequest = ; // CatalogImportRequest | 

try {
    final response = api.apiV1AdminCatalogImportPost(catalogImportRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogImportPost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **catalogImportRequest** | [**CatalogImportRequest**](CatalogImportRequest.md)|  | 

### Return type

[**CatalogTrack**](CatalogTrack.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogTracksGet**
> ApiV1AdminCatalogTracksGet200Response apiV1AdminCatalogTracksGet(artistId, albumId, limit, offset, sortBy, sortOrder, genre)

List tracks

List tracks. sortBy: title (default), sortTitle, trackNumber, discNumber, durationMs, createdAt, updatedAt.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String artistId = artistId_example; // String | 
final String albumId = albumId_example; // String | 
final int limit = 56; // int | Maximum number of items to return (1–500, default 50).
final int offset = 56; // int | Zero-based index of the first item to return.
final String sortBy = sortBy_example; // String | Field to sort by. See endpoint description for valid values.
final String sortOrder = sortOrder_example; // String | Sort direction.
final String genre = genre_example; // String | Filter tracks by genre (case-insensitive exact match)

try {
    final response = api.apiV1AdminCatalogTracksGet(artistId, albumId, limit, offset, sortBy, sortOrder, genre);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogTracksGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **artistId** | **String**|  | [optional] 
 **albumId** | **String**|  | [optional] 
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

# **apiV1AdminCatalogTracksIdDelete**
> apiV1AdminCatalogTracksIdDelete(id)

Delete track

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Catalog entity identifier

try {
    api.apiV1AdminCatalogTracksIdDelete(id);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogTracksIdDelete: $e\n');
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

# **apiV1AdminCatalogTracksIdGet**
> CatalogTrack apiV1AdminCatalogTracksIdGet(id)

Get track

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Catalog entity identifier

try {
    final response = api.apiV1AdminCatalogTracksIdGet(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogTracksIdGet: $e\n');
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

# **apiV1AdminCatalogTracksIdPatch**
> CatalogTrack apiV1AdminCatalogTracksIdPatch(id, catalogUpdateTrackRequest)

Update track metadata

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Catalog entity identifier
final CatalogUpdateTrackRequest catalogUpdateTrackRequest = ; // CatalogUpdateTrackRequest | 

try {
    final response = api.apiV1AdminCatalogTracksIdPatch(id, catalogUpdateTrackRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogTracksIdPatch: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **catalogUpdateTrackRequest** | [**CatalogUpdateTrackRequest**](CatalogUpdateTrackRequest.md)|  | 

### Return type

[**CatalogTrack**](CatalogTrack.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogTracksIdRelinkPost**
> CatalogTrack apiV1AdminCatalogTracksIdRelinkPost(id, catalogRelinkTrackRequest)

Relink track to a different media object

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Catalog entity identifier
final CatalogRelinkTrackRequest catalogRelinkTrackRequest = ; // CatalogRelinkTrackRequest | 

try {
    final response = api.apiV1AdminCatalogTracksIdRelinkPost(id, catalogRelinkTrackRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogTracksIdRelinkPost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Catalog entity identifier | 
 **catalogRelinkTrackRequest** | [**CatalogRelinkTrackRequest**](CatalogRelinkTrackRequest.md)|  | 

### Return type

[**CatalogTrack**](CatalogTrack.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminCatalogTracksPost**
> CatalogTrack apiV1AdminCatalogTracksPost()

Create track

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();

try {
    final response = api.apiV1AdminCatalogTracksPost();
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminCatalogTracksPost: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**CatalogTrack**](CatalogTrack.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminMediaObjectsDuplicatesGet**
> MediaObjectDuplicateReport apiV1AdminMediaObjectsDuplicatesGet(backendId, minCopies)

Find duplicate media objects by content hash.

Returns metadata-only groups of media objects that share the same content hash, optionally limited to one backend. Does not read media bytes.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String backendId = backendId_example; // String | Optional backend ID used to limit duplicate detection to one storage backend.
final int minCopies = 56; // int | Minimum number of objects required for a duplicate group. Defaults to 2.

try {
    final response = api.apiV1AdminMediaObjectsDuplicatesGet(backendId, minCopies);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminMediaObjectsDuplicatesGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **backendId** | **String**| Optional backend ID used to limit duplicate detection to one storage backend. | [optional] 
 **minCopies** | **int**| Minimum number of objects required for a duplicate group. Defaults to 2. | [optional] [default to 2]

### Return type

[**MediaObjectDuplicateReport**](MediaObjectDuplicateReport.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminMediaObjectsGet**
> ApiV1AdminMediaObjectsGet200Response apiV1AdminMediaObjectsGet(backendId, contentHash, verificationStatus, lifecycleState, assetKind, sortBy, sortOrder, limit, offset)

List media objects by metadata filter with offset pagination.

Returns a bounded media-object metadata page filtered by exactly one of backendId, contentHash, verificationStatus, lifecycleState, or assetKind.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String backendId = backendId_example; // String | List media objects stored on this backend. Mutually exclusive with contentHash, verificationStatus, lifecycleState, and assetKind.
final String contentHash = contentHash_example; // String | List media objects with this content hash. Mutually exclusive with backendId, verificationStatus, lifecycleState, and assetKind.
final String verificationStatus = verificationStatus_example; // String | List media objects by their latest verification state; unknown means no verification result has been recorded yet. Mutually exclusive with backendId, contentHash, lifecycleState, and assetKind.
final String lifecycleState = lifecycleState_example; // String | List media objects by metadata lifecycle state. Mutually exclusive with backendId, contentHash, verificationStatus, and assetKind.
final String assetKind = assetKind_example; // String | List media objects by registered asset kind. Mutually exclusive with backendId, contentHash, verificationStatus, and lifecycleState.
final String sortBy = sortBy_example; // String | Sort media object list results before pagination. Defaults to backend_object_key to preserve stable repository ordering.
final String sortOrder = sortOrder_example; // String | Sort direction applied to sortBy before limit and offset are evaluated.
final int limit = 56; // int | Maximum number of media objects to return. Defaults to 100 and cannot exceed 500.
final int offset = 56; // int | Zero-based number of filtered media objects to skip before returning this page.

try {
    final response = api.apiV1AdminMediaObjectsGet(backendId, contentHash, verificationStatus, lifecycleState, assetKind, sortBy, sortOrder, limit, offset);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminMediaObjectsGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **backendId** | **String**| List media objects stored on this backend. Mutually exclusive with contentHash, verificationStatus, lifecycleState, and assetKind. | [optional] 
 **contentHash** | **String**| List media objects with this content hash. Mutually exclusive with backendId, verificationStatus, lifecycleState, and assetKind. | [optional] 
 **verificationStatus** | **String**| List media objects by their latest verification state; unknown means no verification result has been recorded yet. Mutually exclusive with backendId, contentHash, lifecycleState, and assetKind. | [optional] 
 **lifecycleState** | **String**| List media objects by metadata lifecycle state. Mutually exclusive with backendId, contentHash, verificationStatus, and assetKind. | [optional] 
 **assetKind** | **String**| List media objects by registered asset kind. Mutually exclusive with backendId, contentHash, verificationStatus, and lifecycleState. | [optional] 
 **sortBy** | **String**| Sort media object list results before pagination. Defaults to backend_object_key to preserve stable repository ordering. | [optional] [default to 'backend_object_key']
 **sortOrder** | **String**| Sort direction applied to sortBy before limit and offset are evaluated. | [optional] [default to 'asc']
 **limit** | **int**| Maximum number of media objects to return. Defaults to 100 and cannot exceed 500. | [optional] [default to 100]
 **offset** | **int**| Zero-based number of filtered media objects to skip before returning this page. | [optional] [default to 0]

### Return type

[**ApiV1AdminMediaObjectsGet200Response**](ApiV1AdminMediaObjectsGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminMediaObjectsIdGet**
> MediaObject apiV1AdminMediaObjectsIdGet(id)

Get a media object by ID.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Media object identifier assigned by the media object registry.

try {
    final response = api.apiV1AdminMediaObjectsIdGet(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminMediaObjectsIdGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Media object identifier assigned by the media object registry. | 

### Return type

[**MediaObject**](MediaObject.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminMediaObjectsIdLifecyclePost**
> MediaObject apiV1AdminMediaObjectsIdLifecyclePost(id, mediaObjectLifecycleRequest)

Update a media object lifecycle state.

Updates only metadata lifecycle state and updatedAt. It does not delete media bytes or mutate storage backends.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Media object identifier assigned by the media object registry.
final MediaObjectLifecycleRequest mediaObjectLifecycleRequest = ; // MediaObjectLifecycleRequest | 

try {
    final response = api.apiV1AdminMediaObjectsIdLifecyclePost(id, mediaObjectLifecycleRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminMediaObjectsIdLifecyclePost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Media object identifier assigned by the media object registry. | 
 **mediaObjectLifecycleRequest** | [**MediaObjectLifecycleRequest**](MediaObjectLifecycleRequest.md)|  | 

### Return type

[**MediaObject**](MediaObject.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminMediaObjectsIdTimelineGet**
> MediaObjectTimeline apiV1AdminMediaObjectsIdTimelineGet(id)

Get the retained metadata timeline for a media object.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Media object identifier assigned by the media object registry.

try {
    final response = api.apiV1AdminMediaObjectsIdTimelineGet(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminMediaObjectsIdTimelineGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Media object identifier assigned by the media object registry. | 

### Return type

[**MediaObjectTimeline**](MediaObjectTimeline.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminMediaObjectsIdVerifyPost**
> MediaObjectVerificationResult apiV1AdminMediaObjectsIdVerifyPost(id)

Verify a media object metadata reference against stored bytes.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Media object identifier assigned by the media object registry.

try {
    final response = api.apiV1AdminMediaObjectsIdVerifyPost(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminMediaObjectsIdVerifyPost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Media object identifier assigned by the media object registry. | 

### Return type

[**MediaObjectVerificationResult**](MediaObjectVerificationResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminMediaObjectsLifecyclePost**
> MediaObjectLifecycleUpdateReport apiV1AdminMediaObjectsLifecyclePost(mediaObjectBulkLifecycleRequest)

Bulk update media object lifecycle metadata.

Updates only media object lifecycle metadata selected by exactly one metadata filter. Set dryRun to true to preview matched objects and outcomes without persisting changes. Does not delete media bytes or mutate storage backends.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final MediaObjectBulkLifecycleRequest mediaObjectBulkLifecycleRequest = ; // MediaObjectBulkLifecycleRequest | 

try {
    final response = api.apiV1AdminMediaObjectsLifecyclePost(mediaObjectBulkLifecycleRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminMediaObjectsLifecyclePost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **mediaObjectBulkLifecycleRequest** | [**MediaObjectBulkLifecycleRequest**](MediaObjectBulkLifecycleRequest.md)|  | 

### Return type

[**MediaObjectLifecycleUpdateReport**](MediaObjectLifecycleUpdateReport.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminMediaObjectsPost**
> MediaObject apiV1AdminMediaObjectsPost(mediaObjectRequest)

Register a media object metadata reference.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final MediaObjectRequest mediaObjectRequest = ; // MediaObjectRequest | 

try {
    final response = api.apiV1AdminMediaObjectsPost(mediaObjectRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminMediaObjectsPost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **mediaObjectRequest** | [**MediaObjectRequest**](MediaObjectRequest.md)|  | 

### Return type

[**MediaObject**](MediaObject.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminMediaObjectsStatsGet**
> MediaObjectStats apiV1AdminMediaObjectsStatsGet(backendId)

Get media object metadata statistics.

Returns metadata-only media object counts for all backends or one backend without reading media bytes.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String backendId = backendId_example; // String | Optional backend ID used to limit statistics to one storage backend.

try {
    final response = api.apiV1AdminMediaObjectsStatsGet(backendId);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminMediaObjectsStatsGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **backendId** | **String**| Optional backend ID used to limit statistics to one storage backend. | [optional] 

### Return type

[**MediaObjectStats**](MediaObjectStats.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminMediaObjectsVerifyPost**
> MediaObjectVerificationReport apiV1AdminMediaObjectsVerifyPost(backendId, contentHash)

Batch verify media object metadata references by backend ID or content hash.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String backendId = backendId_example; // String | Verify media objects stored on this backend. Mutually exclusive with contentHash.
final String contentHash = contentHash_example; // String | Verify media objects with this content hash. Mutually exclusive with backendId.

try {
    final response = api.apiV1AdminMediaObjectsVerifyPost(backendId, contentHash);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminMediaObjectsVerifyPost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **backendId** | **String**| Verify media objects stored on this backend. Mutually exclusive with contentHash. | [optional] 
 **contentHash** | **String**| Verify media objects with this content hash. Mutually exclusive with backendId. | [optional] 

### Return type

[**MediaObjectVerificationReport**](MediaObjectVerificationReport.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminStorageBackendsGet**
> ApiV1AdminStorageBackendsGet200Response apiV1AdminStorageBackendsGet()

List configured storage backends.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();

try {
    final response = api.apiV1AdminStorageBackendsGet();
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminStorageBackendsGet: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**ApiV1AdminStorageBackendsGet200Response**](ApiV1AdminStorageBackendsGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminStorageBackendsIdCapacityGet**
> CapacityReport apiV1AdminStorageBackendsIdCapacityGet(id)

Read and record backend capacity where supported.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Storage backend identifier assigned by the server-managed backend registry.

try {
    final response = api.apiV1AdminStorageBackendsIdCapacityGet(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminStorageBackendsIdCapacityGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Storage backend identifier assigned by the server-managed backend registry. | 

### Return type

[**CapacityReport**](CapacityReport.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminStorageBackendsIdDefaultPost**
> StorageBackend apiV1AdminStorageBackendsIdDefaultPost(id)

Select a backend as default.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Storage backend identifier assigned by the server-managed backend registry.

try {
    final response = api.apiV1AdminStorageBackendsIdDefaultPost(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminStorageBackendsIdDefaultPost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Storage backend identifier assigned by the server-managed backend registry. | 

### Return type

[**StorageBackend**](StorageBackend.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminStorageBackendsIdDisablePost**
> StorageBackend apiV1AdminStorageBackendsIdDisablePost(id)

Disable a non-default backend.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Storage backend identifier assigned by the server-managed backend registry.

try {
    final response = api.apiV1AdminStorageBackendsIdDisablePost(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminStorageBackendsIdDisablePost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Storage backend identifier assigned by the server-managed backend registry. | 

### Return type

[**StorageBackend**](StorageBackend.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminStorageBackendsIdHealthGet**
> ProbeResult apiV1AdminStorageBackendsIdHealthGet(id)

Read latest recorded backend health.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Storage backend identifier assigned by the server-managed backend registry.

try {
    final response = api.apiV1AdminStorageBackendsIdHealthGet(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminStorageBackendsIdHealthGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Storage backend identifier assigned by the server-managed backend registry. | 

### Return type

[**ProbeResult**](ProbeResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminStorageBackendsIdProbePost**
> ProbeResult apiV1AdminStorageBackendsIdProbePost(id)

Run an on-demand backend probe.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Storage backend identifier assigned by the server-managed backend registry.

try {
    final response = api.apiV1AdminStorageBackendsIdProbePost(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminStorageBackendsIdProbePost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Storage backend identifier assigned by the server-managed backend registry. | 

### Return type

[**ProbeResult**](ProbeResult.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminStorageBackendsPost**
> StorageBackend apiV1AdminStorageBackendsPost(storageBackendRequest)

Validate and register a storage backend.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final StorageBackendRequest storageBackendRequest = ; // StorageBackendRequest | 

try {
    final response = api.apiV1AdminStorageBackendsPost(storageBackendRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminStorageBackendsPost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **storageBackendRequest** | [**StorageBackendRequest**](StorageBackendRequest.md)|  | 

### Return type

[**StorageBackend**](StorageBackend.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminStorageBackendsRefreshPost**
> RefreshReport apiV1AdminStorageBackendsRefreshPost()

Refresh all enabled backend health and supported capacity state.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();

try {
    final response = api.apiV1AdminStorageBackendsRefreshPost();
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminStorageBackendsRefreshPost: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**RefreshReport**](RefreshReport.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminStorageBackendsValidatePost**
> StorageBackend apiV1AdminStorageBackendsValidatePost(storageBackendRequest)

Validate a backend candidate without persisting it.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final StorageBackendRequest storageBackendRequest = ; // StorageBackendRequest | 

try {
    final response = api.apiV1AdminStorageBackendsValidatePost(storageBackendRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminStorageBackendsValidatePost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **storageBackendRequest** | [**StorageBackendRequest**](StorageBackendRequest.md)|  | 

### Return type

[**StorageBackend**](StorageBackend.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminUsersGet**
> ApiV1AdminUsersGet200Response apiV1AdminUsersGet(limit, offset, sortBy, sortOrder, username, role, enabled)

List users.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final int limit = 56; // int | Maximum number of items to return (0 = all).
final int offset = 56; // int | Zero-based index of the first item to return.
final String sortBy = sortBy_example; // String | Field to sort by.
final String sortOrder = sortOrder_example; // String | Sort direction.
final String username = username_example; // String | Filter by exact username match.
final String role = role_example; // String | Filter by user role.
final bool enabled = true; // bool | Filter by enabled status.

try {
    final response = api.apiV1AdminUsersGet(limit, offset, sortBy, sortOrder, username, role, enabled);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminUsersGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **limit** | **int**| Maximum number of items to return (0 = all). | [optional] [default to 0]
 **offset** | **int**| Zero-based index of the first item to return. | [optional] [default to 0]
 **sortBy** | **String**| Field to sort by. | [optional] [default to 'username']
 **sortOrder** | **String**| Sort direction. | [optional] [default to 'asc']
 **username** | **String**| Filter by exact username match. | [optional] 
 **role** | **String**| Filter by user role. | [optional] 
 **enabled** | **bool**| Filter by enabled status. | [optional] 

### Return type

[**ApiV1AdminUsersGet200Response**](ApiV1AdminUsersGet200Response.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminUsersIdChangePasswordPost**
> apiV1AdminUsersIdChangePasswordPost(id, forceChangePasswordRequest)

Force-reset a user's password without requiring the current password.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | User identifier
final ForceChangePasswordRequest forceChangePasswordRequest = ; // ForceChangePasswordRequest | 

try {
    api.apiV1AdminUsersIdChangePasswordPost(id, forceChangePasswordRequest);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminUsersIdChangePasswordPost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| User identifier | 
 **forceChangePasswordRequest** | [**ForceChangePasswordRequest**](ForceChangePasswordRequest.md)|  | 

### Return type

void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminUsersIdDelete**
> apiV1AdminUsersIdDelete(id, id2)

Delete a user.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | 
final String id2 = id_example; // String | User identifier

try {
    api.apiV1AdminUsersIdDelete(id, id2);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminUsersIdDelete: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**|  | 
 **id2** | **String**| User identifier | 

### Return type

void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminUsersIdDisablePost**
> UserView apiV1AdminUsersIdDisablePost(id, id2)

Disable a user.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | 
final String id2 = id_example; // String | User identifier

try {
    final response = api.apiV1AdminUsersIdDisablePost(id, id2);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminUsersIdDisablePost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**|  | 
 **id2** | **String**| User identifier | 

### Return type

[**UserView**](UserView.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminUsersIdEnablePost**
> UserView apiV1AdminUsersIdEnablePost(id, id2)

Enable a previously disabled user.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | 
final String id2 = id_example; // String | User identifier

try {
    final response = api.apiV1AdminUsersIdEnablePost(id, id2);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminUsersIdEnablePost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**|  | 
 **id2** | **String**| User identifier | 

### Return type

[**UserView**](UserView.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminUsersIdGet**
> UserView apiV1AdminUsersIdGet(id)

Get a user by ID.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | User identifier

try {
    final response = api.apiV1AdminUsersIdGet(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminUsersIdGet: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| User identifier | 

### Return type

[**UserView**](UserView.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminUsersIdPatch**
> UserView apiV1AdminUsersIdPatch(id, patchUserRequest)

Partially update a user's role or username.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | User identifier
final PatchUserRequest patchUserRequest = ; // PatchUserRequest | 

try {
    final response = api.apiV1AdminUsersIdPatch(id, patchUserRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminUsersIdPatch: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| User identifier | 
 **patchUserRequest** | [**PatchUserRequest**](PatchUserRequest.md)|  | 

### Return type

[**UserView**](UserView.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AdminUsersPost**
> UserView apiV1AdminUsersPost(createUserRequest)

Create a user.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final CreateUserRequest createUserRequest = ; // CreateUserRequest | 

try {
    final response = api.apiV1AdminUsersPost(createUserRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AdminUsersPost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **createUserRequest** | [**CreateUserRequest**](CreateUserRequest.md)|  | 

### Return type

[**UserView**](UserView.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AuthLoginPost**
> LoginResponse apiV1AuthLoginPost(loginRequest)

Create a session token from username and password.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final LoginRequest loginRequest = ; // LoginRequest | 

try {
    final response = api.apiV1AuthLoginPost(loginRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AuthLoginPost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **loginRequest** | [**LoginRequest**](LoginRequest.md)|  | 

### Return type

[**LoginResponse**](LoginResponse.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1AuthLogoutPost**
> apiV1AuthLogoutPost()

Revoke the current session token.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();

try {
    api.apiV1AuthLogoutPost();
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1AuthLogoutPost: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

void (empty response body)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1MeChangePasswordPost**
> apiV1MeChangePasswordPost(changePasswordRequest)

Change the authenticated viewer's password.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final ChangePasswordRequest changePasswordRequest = ; // ChangePasswordRequest | 

try {
    api.apiV1MeChangePasswordPost(changePasswordRequest);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1MeChangePasswordPost: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **changePasswordRequest** | [**ChangePasswordRequest**](ChangePasswordRequest.md)|  | 

### Return type

void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1MeGet**
> UserView apiV1MeGet()

Get the authenticated viewer's profile.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();

try {
    final response = api.apiV1MeGet();
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1MeGet: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**UserView**](UserView.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **apiV1MeSessionsRevokeAllDevicesPost**
> DeleteAdminUserSessions200Response apiV1MeSessionsRevokeAllDevicesPost()

Revoke all sessions for the authenticated user, including the current session.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();

try {
    final response = api.apiV1MeSessionsRevokeAllDevicesPost();
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->apiV1MeSessionsRevokeAllDevicesPost: $e\n');
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

# **deleteStorageBackend**
> deleteStorageBackend(id)

Delete a storage backend

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Storage backend identifier assigned by the server-managed backend registry.

try {
    api.deleteStorageBackend(id);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->deleteStorageBackend: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Storage backend identifier assigned by the server-managed backend registry. | 

### Return type

void (empty response body)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **enableStorageBackend**
> StorageBackend enableStorageBackend(id)

Enable a storage backend

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Storage backend identifier assigned by the server-managed backend registry.

try {
    final response = api.enableStorageBackend(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->enableStorageBackend: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Storage backend identifier assigned by the server-managed backend registry. | 

### Return type

[**StorageBackend**](StorageBackend.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **getStorageBackend**
> StorageBackend getStorageBackend(id)

Get a storage backend by ID

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Storage backend identifier assigned by the server-managed backend registry.

try {
    final response = api.getStorageBackend(id);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->getStorageBackend: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Storage backend identifier assigned by the server-managed backend registry. | 

### Return type

[**StorageBackend**](StorageBackend.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **healthzGet**
> HealthzGet200Response healthzGet()

Return process health.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();

try {
    final response = api.healthzGet();
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->healthzGet: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**HealthzGet200Response**](HealthzGet200Response.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **metricsGet**
> String metricsGet()

Get public Prometheus-compatible process metrics.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();

try {
    final response = api.metricsGet();
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->metricsGet: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

**String**

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: text/plain; version=0.0.4

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **patchMediaObject**
> MediaObject patchMediaObject(id, patchMediaObjectRequest)

Update correctable metadata on a media object (assetKind, mimeType)

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Media object identifier assigned by the media object registry.
final PatchMediaObjectRequest patchMediaObjectRequest = ; // PatchMediaObjectRequest | 

try {
    final response = api.patchMediaObject(id, patchMediaObjectRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->patchMediaObject: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Media object identifier assigned by the media object registry. | 
 **patchMediaObjectRequest** | [**PatchMediaObjectRequest**](PatchMediaObjectRequest.md)|  | 

### Return type

[**MediaObject**](MediaObject.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **patchStorageBackend**
> StorageBackend patchStorageBackend(id, patchStorageBackendRequest)

Update a storage backend's display name or priority

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();
final String id = id_example; // String | Storage backend identifier assigned by the server-managed backend registry.
final PatchStorageBackendRequest patchStorageBackendRequest = ; // PatchStorageBackendRequest | 

try {
    final response = api.patchStorageBackend(id, patchStorageBackendRequest);
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->patchStorageBackend: $e\n');
}
```

### Parameters

Name | Type | Description  | Notes
------------- | ------------- | ------------- | -------------
 **id** | **String**| Storage backend identifier assigned by the server-managed backend registry. | 
 **patchStorageBackendRequest** | [**PatchStorageBackendRequest**](PatchStorageBackendRequest.md)|  | 

### Return type

[**StorageBackend**](StorageBackend.md)

### Authorization

[bearerAuth](../README.md#bearerAuth)

### HTTP request headers

 - **Content-Type**: application/json
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **readyzGet**
> ReadinessReport readyzGet()

Get public readiness status for this API process.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();

try {
    final response = api.readyzGet();
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->readyzGet: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**ReadinessReport**](ReadinessReport.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

# **versionzGet**
> ServiceInfo versionzGet()

Get public build metadata for this API process.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getDefaultApi();

try {
    final response = api.versionzGet();
    print(response);
} on DioException catch (e) {
    print('Exception when calling DefaultApi->versionzGet: $e\n');
}
```

### Parameters
This endpoint does not need any parameter.

### Return type

[**ServiceInfo**](ServiceInfo.md)

### Authorization

No authorization required

### HTTP request headers

 - **Content-Type**: Not defined
 - **Accept**: application/json

[[Back to top]](#) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to Model list]](../README.md#documentation-for-models) [[Back to README]](../README.md)

