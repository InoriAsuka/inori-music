# inori_api.api.CatalogAdminApi

## Load the API package
```dart
import 'package:inori_api/api.dart';
```

All URIs are relative to *http://127.0.0.1:8080*

Method | HTTP request | Description
------------- | ------------- | -------------
[**getAlbumStatsBreakdown**](CatalogAdminApi.md#getalbumstatsbreakdown) | **GET** /api/v1/admin/catalog/stats/albums | Per-album catalog stats breakdown
[**getArtistStatsBreakdown**](CatalogAdminApi.md#getartiststatsbreakdown) | **GET** /api/v1/admin/catalog/stats/artists | Per-artist catalog stats breakdown
[**getCatalogStats**](CatalogAdminApi.md#getcatalogstats) | **GET** /api/v1/admin/catalog/stats | Catalog entity count statistics
[**getPlaylistStatsBreakdown**](CatalogAdminApi.md#getplayliststatsbreakdown) | **GET** /api/v1/admin/catalog/stats/playlists | Per-playlist catalog stats breakdown
[**getRecentlyAddedCatalogItems**](CatalogAdminApi.md#getrecentlyaddedcatalogitems) | **GET** /api/v1/admin/catalog/recently-added | List recently added catalog items
[**getRecentlyUpdatedCatalogItems**](CatalogAdminApi.md#getrecentlyupdatedcatalogitems) | **GET** /api/v1/admin/catalog/recently-updated | List recently updated catalog items


# **getAlbumStatsBreakdown**
> CatalogAlbumStatsBreakdown getAlbumStatsBreakdown()

Per-album catalog stats breakdown

Returns per-album track counts derived from catalog metadata.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogAdminApi();

try {
    final response = api.getAlbumStatsBreakdown();
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogAdminApi->getAlbumStatsBreakdown: $e\n');
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

# **getArtistStatsBreakdown**
> CatalogArtistStatsBreakdown getArtistStatsBreakdown()

Per-artist catalog stats breakdown

Returns per-artist album and track counts derived from catalog metadata.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogAdminApi();

try {
    final response = api.getArtistStatsBreakdown();
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogAdminApi->getArtistStatsBreakdown: $e\n');
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

# **getCatalogStats**
> CatalogStats getCatalogStats()

Catalog entity count statistics

Returns metadata-only aggregate counts for artists, albums, tracks, and playlists.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogAdminApi();

try {
    final response = api.getCatalogStats();
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogAdminApi->getCatalogStats: $e\n');
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

# **getPlaylistStatsBreakdown**
> CatalogPlaylistStatsBreakdown getPlaylistStatsBreakdown()

Per-playlist catalog stats breakdown

Returns per-playlist track counts derived from catalog metadata.

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogAdminApi();

try {
    final response = api.getPlaylistStatsBreakdown();
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogAdminApi->getPlaylistStatsBreakdown: $e\n');
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

# **getRecentlyAddedCatalogItems**
> RecentCatalogResult getRecentlyAddedCatalogItems(kind, limit)

List recently added catalog items

Returns a newest-first unified timeline of recently created artists, albums, tracks, and playlists. Use kind to filter to one entity type and limit to cap the response size (default 20, max 100).

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogAdminApi();
final RecentItemKind kind = ; // RecentItemKind | Optional entity kind filter.
final int limit = 56; // int | Maximum number of items to return. Defaults to 20; values above 100 are clamped.

try {
    final response = api.getRecentlyAddedCatalogItems(kind, limit);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogAdminApi->getRecentlyAddedCatalogItems: $e\n');
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

# **getRecentlyUpdatedCatalogItems**
> UpdatedCatalogResult getRecentlyUpdatedCatalogItems(kind, limit)

List recently updated catalog items

Returns a newest-first unified timeline of recently updated artists, albums, tracks, and playlists. Use kind to filter to one entity type and limit to cap the response size (default 20, max 100).

### Example
```dart
import 'package:inori_api/api.dart';

final api = InoriApi().getCatalogAdminApi();
final RecentItemKind kind = ; // RecentItemKind | Optional entity kind filter.
final int limit = 56; // int | Maximum number of items to return. Defaults to 20; values above 100 are clamped.

try {
    final response = api.getRecentlyUpdatedCatalogItems(kind, limit);
    print(response);
} on DioException catch (e) {
    print('Exception when calling CatalogAdminApi->getRecentlyUpdatedCatalogItems: $e\n');
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

