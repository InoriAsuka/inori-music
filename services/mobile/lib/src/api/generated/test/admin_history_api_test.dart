import 'package:test/test.dart';
import 'package:inori_api/inori_api.dart';


/// tests for AdminHistoryApi
void main() {
  final instance = InoriApi().getAdminHistoryApi();

  group(AdminHistoryApi, () {
    // Get history aggregate stats
    //
    // Returns system-wide playback aggregate counts (admin only).
    //
    //Future<HistoryStats> getAdminHistoryStats({ DateTime since, DateTime until }) async
    test('test getAdminHistoryStats', () async {
      // TODO
    });

    // Get play history grouped by time bucket (admin)
    //
    // Returns play event counts grouped by day, week, or month within the specified time window. Both since and until are required. Admin only.
    //
    //Future<TimelineResult> getAdminHistoryTimeline(DateTime since, DateTime until, { String granularity, String userId, String trackId }) async
    test('test getAdminHistoryTimeline', () async {
      // TODO
    });

    // Get most-played tracks
    //
    // Returns tracks with the highest total play count across all users (admin only).
    //
    //Future<TopTracksResult> getAdminTopTracks({ int limit, DateTime since, DateTime until }) async
    test('test getAdminTopTracks', () async {
      // TODO
    });

    // Get most-active users
    //
    // Returns users with the highest total play event count (admin only).
    //
    //Future<TopUsersResult> getAdminTopUsers({ int limit, DateTime since, DateTime until }) async
    test('test getAdminTopUsers', () async {
      // TODO
    });

    // Get play history stats for a specific track (admin)
    //
    // Returns aggregate play counts (total events, unique listeners) for any track. Admin only.
    //
    //Future<TrackHistoryStats> getAdminTrackStats(String trackId, { DateTime since, DateTime until }) async
    test('test getAdminTrackStats', () async {
      // TODO
    });

    // Get top listeners for a specific track (admin)
    //
    // Returns the users who have played a track the most. Admin only.
    //
    //Future<TopUsersResult> getAdminTrackTopListeners(String trackId, { int limit, DateTime since, DateTime until }) async
    test('test getAdminTrackTopListeners', () async {
      // TODO
    });

    // Get play history stats for a specific user (admin)
    //
    // Returns aggregate play counts (total events, unique tracks) for any user. Admin only.
    //
    //Future<UserHistoryStats> getAdminUserStats(String userId, { DateTime since, DateTime until }) async
    test('test getAdminUserStats', () async {
      // TODO
    });

    // Get top tracks for a specific user (admin)
    //
    // Returns the tracks most frequently played by any user. Admin only.
    //
    //Future<TopTracksResult> getAdminUserTopTracks(String userId, { int limit, DateTime since, DateTime until }) async
    test('test getAdminUserTopTracks', () async {
      // TODO
    });

  });
}
