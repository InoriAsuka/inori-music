import 'package:inori_api/src/model/add_playlist_track_request.dart';
import 'package:inori_api/src/model/add_user_playlist_track_request.dart';
import 'package:inori_api/src/model/admin_list_user_favorites200_response.dart';
import 'package:inori_api/src/model/admin_list_user_favorites200_response_pagination.dart';
import 'package:inori_api/src/model/album_artwork_response.dart';
import 'package:inori_api/src/model/api_v1_admin_catalog_albums_get200_response.dart';
import 'package:inori_api/src/model/api_v1_admin_catalog_artists_get200_response.dart';
import 'package:inori_api/src/model/api_v1_admin_catalog_playlists_get200_response.dart';
import 'package:inori_api/src/model/api_v1_admin_catalog_playlists_id_tracks_get200_response.dart';
import 'package:inori_api/src/model/api_v1_admin_catalog_tracks_get200_response.dart';
import 'package:inori_api/src/model/api_v1_admin_media_objects_get200_response.dart';
import 'package:inori_api/src/model/api_v1_admin_storage_backends_get200_response.dart';
import 'package:inori_api/src/model/api_v1_admin_users_get200_response.dart';
import 'package:inori_api/src/model/api_v1_admin_users_get200_response_pagination.dart';
import 'package:inori_api/src/model/api_v1_me_history_top_tracks_get200_response.dart';
import 'package:inori_api/src/model/backend_config.dart';
import 'package:inori_api/src/model/batch_delete_request.dart';
import 'package:inori_api/src/model/batch_delete_result.dart';
import 'package:inori_api/src/model/capability_set.dart';
import 'package:inori_api/src/model/capacity_report.dart';
import 'package:inori_api/src/model/catalog_album.dart';
import 'package:inori_api/src/model/catalog_album_stat_item.dart';
import 'package:inori_api/src/model/catalog_album_stats_breakdown.dart';
import 'package:inori_api/src/model/catalog_artist.dart';
import 'package:inori_api/src/model/catalog_artist_stat_item.dart';
import 'package:inori_api/src/model/catalog_artist_stats_breakdown.dart';
import 'package:inori_api/src/model/catalog_batch_import_request.dart';
import 'package:inori_api/src/model/catalog_batch_import_result.dart';
import 'package:inori_api/src/model/catalog_batch_import_result_item.dart';
import 'package:inori_api/src/model/catalog_import_request.dart';
import 'package:inori_api/src/model/catalog_pagination_meta.dart';
import 'package:inori_api/src/model/catalog_playlist_stat_item.dart';
import 'package:inori_api/src/model/catalog_playlist_stats_breakdown.dart';
import 'package:inori_api/src/model/catalog_relink_track_request.dart';
import 'package:inori_api/src/model/catalog_search_result.dart';
import 'package:inori_api/src/model/catalog_stats.dart';
import 'package:inori_api/src/model/catalog_track.dart';
import 'package:inori_api/src/model/catalog_update_album_request.dart';
import 'package:inori_api/src/model/catalog_update_artist_request.dart';
import 'package:inori_api/src/model/catalog_update_track_request.dart';
import 'package:inori_api/src/model/change_password_request.dart';
import 'package:inori_api/src/model/create_playlist_request.dart';
import 'package:inori_api/src/model/create_user_playlist_request.dart';
import 'package:inori_api/src/model/create_user_request.dart';
import 'package:inori_api/src/model/delete_admin_user_sessions200_response.dart';
import 'package:inori_api/src/model/distributed_config.dart';
import 'package:inori_api/src/model/error_envelope.dart';
import 'package:inori_api/src/model/error_envelope_error.dart';
import 'package:inori_api/src/model/favorites_page.dart';
import 'package:inori_api/src/model/force_change_password_request.dart';
import 'package:inori_api/src/model/get_admin_user_sessions200_response.dart';
import 'package:inori_api/src/model/get_my_history_summary200_response.dart';
import 'package:inori_api/src/model/get_user_playlist_tracks200_response.dart';
import 'package:inori_api/src/model/global_history_summary.dart';
import 'package:inori_api/src/model/healthz_get200_response.dart';
import 'package:inori_api/src/model/history_stats.dart';
import 'package:inori_api/src/model/list_my_track_history200_response.dart';
import 'package:inori_api/src/model/list_user_playlists200_response.dart';
import 'package:inori_api/src/model/local_config.dart';
import 'package:inori_api/src/model/login_request.dart';
import 'package:inori_api/src/model/login_response.dart';
import 'package:inori_api/src/model/lyrics_response.dart';
import 'package:inori_api/src/model/media_object.dart';
import 'package:inori_api/src/model/media_object_bulk_lifecycle_request.dart';
import 'package:inori_api/src/model/media_object_duplicate_group.dart';
import 'package:inori_api/src/model/media_object_duplicate_report.dart';
import 'package:inori_api/src/model/media_object_lifecycle_change.dart';
import 'package:inori_api/src/model/media_object_lifecycle_request.dart';
import 'package:inori_api/src/model/media_object_lifecycle_update_report.dart';
import 'package:inori_api/src/model/media_object_lifecycle_update_result.dart';
import 'package:inori_api/src/model/media_object_request.dart';
import 'package:inori_api/src/model/media_object_selection_filter.dart';
import 'package:inori_api/src/model/media_object_stats.dart';
import 'package:inori_api/src/model/media_object_timeline.dart';
import 'package:inori_api/src/model/media_object_timeline_event.dart';
import 'package:inori_api/src/model/media_object_verification_report.dart';
import 'package:inori_api/src/model/media_object_verification_result.dart';
import 'package:inori_api/src/model/my_track_summary.dart';
import 'package:inori_api/src/model/nfs_config.dart';
import 'package:inori_api/src/model/pagination_metadata.dart';
import 'package:inori_api/src/model/patch_media_object_request.dart';
import 'package:inori_api/src/model/patch_storage_backend_request.dart';
import 'package:inori_api/src/model/patch_user_request.dart';
import 'package:inori_api/src/model/play_event.dart';
import 'package:inori_api/src/model/play_event_list.dart';
import 'package:inori_api/src/model/playlist.dart';
import 'package:inori_api/src/model/playlist_tracks_result.dart';
import 'package:inori_api/src/model/probe_result.dart';
import 'package:inori_api/src/model/readiness_check.dart';
import 'package:inori_api/src/model/readiness_report.dart';
import 'package:inori_api/src/model/recent_catalog_item.dart';
import 'package:inori_api/src/model/recent_catalog_result.dart';
import 'package:inori_api/src/model/record_play_event_request.dart';
import 'package:inori_api/src/model/refresh_report.dart';
import 'package:inori_api/src/model/refresh_result.dart';
import 'package:inori_api/src/model/s3_config.dart';
import 'package:inori_api/src/model/smb_config.dart';
import 'package:inori_api/src/model/search_result_item.dart';
import 'package:inori_api/src/model/service_info.dart';
import 'package:inori_api/src/model/session_view.dart';
import 'package:inori_api/src/model/set_playlist_tracks_request.dart';
import 'package:inori_api/src/model/set_user_playlist_tracks_request.dart';
import 'package:inori_api/src/model/storage_backend.dart';
import 'package:inori_api/src/model/storage_backend_request.dart';
import 'package:inori_api/src/model/timeline_bucket.dart';
import 'package:inori_api/src/model/timeline_result.dart';
import 'package:inori_api/src/model/top_tracks_result.dart';
import 'package:inori_api/src/model/top_users_result.dart';
import 'package:inori_api/src/model/track_history_stats.dart';
import 'package:inori_api/src/model/track_history_summary.dart';
import 'package:inori_api/src/model/track_play_count.dart';
import 'package:inori_api/src/model/track_playback_descriptor.dart';
import 'package:inori_api/src/model/update_play_event_request.dart';
import 'package:inori_api/src/model/update_playlist_request.dart';
import 'package:inori_api/src/model/update_user_playlist_request.dart';
import 'package:inori_api/src/model/updated_catalog_item.dart';
import 'package:inori_api/src/model/updated_catalog_result.dart';
import 'package:inori_api/src/model/upload_track_lyrics201_response.dart';
import 'package:inori_api/src/model/user_history_stats.dart';
import 'package:inori_api/src/model/user_history_summary.dart';
import 'package:inori_api/src/model/user_play_count.dart';
import 'package:inori_api/src/model/user_playlist.dart';
import 'package:inori_api/src/model/user_track_stats.dart';
import 'package:inori_api/src/model/user_view.dart';

final _regList = RegExp(r'^List<(.*)>$');
final _regSet = RegExp(r'^Set<(.*)>$');
final _regMap = RegExp(r'^Map<String,(.*)>$');

  ReturnType deserialize<ReturnType, BaseType>(dynamic value, String targetType, {bool growable= true}) {
      switch (targetType) {
        case 'String':
          return '$value' as ReturnType;
        case 'int':
          return (value is int ? value : int.parse('$value')) as ReturnType;
        case 'bool':
          if (value is bool) {
            return value as ReturnType;
          }
          final valueString = '$value'.toLowerCase();
          return (valueString == 'true' || valueString == '1') as ReturnType;
        case 'double':
          return (value is double ? value : double.parse('$value')) as ReturnType;
        case 'AddPlaylistTrackRequest':
          return AddPlaylistTrackRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'AddUserPlaylistTrackRequest':
          return AddUserPlaylistTrackRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'AdminListUserFavorites200Response':
          return AdminListUserFavorites200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'AdminListUserFavorites200ResponsePagination':
          return AdminListUserFavorites200ResponsePagination.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'AlbumArtworkResponse':
          return AlbumArtworkResponse.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ApiV1AdminCatalogAlbumsGet200Response':
          return ApiV1AdminCatalogAlbumsGet200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ApiV1AdminCatalogArtistsGet200Response':
          return ApiV1AdminCatalogArtistsGet200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ApiV1AdminCatalogPlaylistsGet200Response':
          return ApiV1AdminCatalogPlaylistsGet200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ApiV1AdminCatalogPlaylistsIdTracksGet200Response':
          return ApiV1AdminCatalogPlaylistsIdTracksGet200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ApiV1AdminCatalogTracksGet200Response':
          return ApiV1AdminCatalogTracksGet200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ApiV1AdminMediaObjectsGet200Response':
          return ApiV1AdminMediaObjectsGet200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ApiV1AdminStorageBackendsGet200Response':
          return ApiV1AdminStorageBackendsGet200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ApiV1AdminUsersGet200Response':
          return ApiV1AdminUsersGet200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ApiV1AdminUsersGet200ResponsePagination':
          return ApiV1AdminUsersGet200ResponsePagination.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ApiV1MeHistoryTopTracksGet200Response':
          return ApiV1MeHistoryTopTracksGet200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'BackendConfig':
          return BackendConfig.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'BackendType':
          
          
        case 'BatchDeleteRequest':
          return BatchDeleteRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'BatchDeleteResult':
          return BatchDeleteResult.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CapabilitySet':
          return CapabilitySet.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CapacityReport':
          return CapacityReport.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogAlbum':
          return CatalogAlbum.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogAlbumStatItem':
          return CatalogAlbumStatItem.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogAlbumStatsBreakdown':
          return CatalogAlbumStatsBreakdown.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogArtist':
          return CatalogArtist.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogArtistStatItem':
          return CatalogArtistStatItem.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogArtistStatsBreakdown':
          return CatalogArtistStatsBreakdown.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogBatchImportRequest':
          return CatalogBatchImportRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogBatchImportResult':
          return CatalogBatchImportResult.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogBatchImportResultItem':
          return CatalogBatchImportResultItem.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogImportRequest':
          return CatalogImportRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogPaginationMeta':
          return CatalogPaginationMeta.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogPlaylistStatItem':
          return CatalogPlaylistStatItem.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogPlaylistStatsBreakdown':
          return CatalogPlaylistStatsBreakdown.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogRelinkTrackRequest':
          return CatalogRelinkTrackRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogSearchResult':
          return CatalogSearchResult.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogStats':
          return CatalogStats.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogTrack':
          return CatalogTrack.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogUpdateAlbumRequest':
          return CatalogUpdateAlbumRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogUpdateArtistRequest':
          return CatalogUpdateArtistRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CatalogUpdateTrackRequest':
          return CatalogUpdateTrackRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ChangePasswordRequest':
          return ChangePasswordRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CreatePlaylistRequest':
          return CreatePlaylistRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CreateUserPlaylistRequest':
          return CreateUserPlaylistRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'CreateUserRequest':
          return CreateUserRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'DeleteAdminUserSessions200Response':
          return DeleteAdminUserSessions200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'DistributedConfig':
          return DistributedConfig.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ErrorEnvelope':
          return ErrorEnvelope.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ErrorEnvelopeError':
          return ErrorEnvelopeError.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'FavoritesPage':
          return FavoritesPage.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ForceChangePasswordRequest':
          return ForceChangePasswordRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'GetAdminUserSessions200Response':
          return GetAdminUserSessions200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'GetMyHistorySummary200Response':
          return GetMyHistorySummary200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'GetUserPlaylistTracks200Response':
          return GetUserPlaylistTracks200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'GlobalHistorySummary':
          return GlobalHistorySummary.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'HealthStatus':
          
          
        case 'HealthzGet200Response':
          return HealthzGet200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'HistoryStats':
          return HistoryStats.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ListMyTrackHistory200Response':
          return ListMyTrackHistory200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ListUserPlaylists200Response':
          return ListUserPlaylists200Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'LocalConfig':
          return LocalConfig.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'LoginRequest':
          return LoginRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'LoginResponse':
          return LoginResponse.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'LyricsResponse':
          return LyricsResponse.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'MediaObject':
          return MediaObject.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'MediaObjectBulkLifecycleRequest':
          return MediaObjectBulkLifecycleRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'MediaObjectDuplicateGroup':
          return MediaObjectDuplicateGroup.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'MediaObjectDuplicateReport':
          return MediaObjectDuplicateReport.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'MediaObjectLifecycleChange':
          return MediaObjectLifecycleChange.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'MediaObjectLifecycleRequest':
          return MediaObjectLifecycleRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'MediaObjectLifecycleUpdateReport':
          return MediaObjectLifecycleUpdateReport.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'MediaObjectLifecycleUpdateResult':
          return MediaObjectLifecycleUpdateResult.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'MediaObjectRequest':
          return MediaObjectRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'MediaObjectSelectionFilter':
          return MediaObjectSelectionFilter.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'MediaObjectStats':
          return MediaObjectStats.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'MediaObjectTimeline':
          return MediaObjectTimeline.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'MediaObjectTimelineEvent':
          return MediaObjectTimelineEvent.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'MediaObjectVerificationReport':
          return MediaObjectVerificationReport.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'MediaObjectVerificationResult':
          return MediaObjectVerificationResult.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'MyTrackSummary':
          return MyTrackSummary.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'NFSConfig':
          return NFSConfig.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'PaginationMetadata':
          return PaginationMetadata.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'PatchMediaObjectRequest':
          return PatchMediaObjectRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'PatchStorageBackendRequest':
          return PatchStorageBackendRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'PatchUserRequest':
          return PatchUserRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'PlayEvent':
          return PlayEvent.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'PlayEventList':
          return PlayEventList.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'Playlist':
          return Playlist.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'PlaylistTracksResult':
          return PlaylistTracksResult.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ProbeResult':
          return ProbeResult.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ReadinessCheck':
          return ReadinessCheck.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'ReadinessReport':
          return ReadinessReport.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'RecentCatalogItem':
          return RecentCatalogItem.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'RecentCatalogResult':
          return RecentCatalogResult.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'RecentItemKind':
          
          
        case 'RecordPlayEventRequest':
          return RecordPlayEventRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'RefreshReport':
          return RefreshReport.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'RefreshResult':
          return RefreshResult.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'S3Config':
          return S3Config.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'SMBConfig':
          return SMBConfig.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'SearchResultItem':
          return SearchResultItem.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'SearchResultKind':
          
          
        case 'ServiceInfo':
          return ServiceInfo.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'SessionView':
          return SessionView.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'SetPlaylistTracksRequest':
          return SetPlaylistTracksRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'SetUserPlaylistTracksRequest':
          return SetUserPlaylistTracksRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'StorageBackend':
          return StorageBackend.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'StorageBackendRequest':
          return StorageBackendRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'TimelineBucket':
          return TimelineBucket.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'TimelineResult':
          return TimelineResult.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'TopTracksResult':
          return TopTracksResult.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'TopUsersResult':
          return TopUsersResult.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'TrackHistoryStats':
          return TrackHistoryStats.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'TrackHistorySummary':
          return TrackHistorySummary.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'TrackPlayCount':
          return TrackPlayCount.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'TrackPlaybackDescriptor':
          return TrackPlaybackDescriptor.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'UpdatePlayEventRequest':
          return UpdatePlayEventRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'UpdatePlaylistRequest':
          return UpdatePlaylistRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'UpdateUserPlaylistRequest':
          return UpdateUserPlaylistRequest.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'UpdatedCatalogItem':
          return UpdatedCatalogItem.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'UpdatedCatalogResult':
          return UpdatedCatalogResult.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'UploadTrackLyrics201Response':
          return UploadTrackLyrics201Response.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'UserHistoryStats':
          return UserHistoryStats.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'UserHistorySummary':
          return UserHistorySummary.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'UserPlayCount':
          return UserPlayCount.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'UserPlaylist':
          return UserPlaylist.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'UserRole':
          
          
        case 'UserTrackStats':
          return UserTrackStats.fromJson(value as Map<String, dynamic>) as ReturnType;
        case 'UserView':
          return UserView.fromJson(value as Map<String, dynamic>) as ReturnType;
        default:
          RegExpMatch? match;

          if (value is List && (match = _regList.firstMatch(targetType)) != null) {
            targetType = match![1]!; // ignore: parameter_assignments
            return value
              .map<BaseType>((dynamic v) => deserialize<BaseType, BaseType>(v, targetType, growable: growable))
              .toList(growable: growable) as ReturnType;
          }
          if (value is Set && (match = _regSet.firstMatch(targetType)) != null) {
            targetType = match![1]!; // ignore: parameter_assignments
            return value
              .map<BaseType>((dynamic v) => deserialize<BaseType, BaseType>(v, targetType, growable: growable))
              .toSet() as ReturnType;
          }
          if (value is Map && (match = _regMap.firstMatch(targetType)) != null) {
            targetType = match![1]!.trim(); // ignore: parameter_assignments
            return Map<String, BaseType>.fromIterables(
              value.keys as Iterable<String>,
              value.values.map((dynamic v) => deserialize<BaseType, BaseType>(v, targetType, growable: growable)),
            ) as ReturnType;
          }
          break;
    }
    throw Exception('Cannot deserialize');
  }