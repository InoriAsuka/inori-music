import 'package:test/test.dart';
import 'package:inori_api/inori_api.dart';


/// tests for UserPlaylistsApi
void main() {
  final instance = InoriApi().getUserPlaylistsApi();

  group(UserPlaylistsApi, () {
    // Add track to user playlist
    //
    //Future<UserPlaylist> addUserPlaylistTrack(String id, AddUserPlaylistTrackRequest addUserPlaylistTrackRequest) async
    test('test addUserPlaylistTrack', () async {
      // TODO
    });

    // Create user playlist
    //
    //Future<UserPlaylist> createUserPlaylist(CreateUserPlaylistRequest createUserPlaylistRequest) async
    test('test createUserPlaylist', () async {
      // TODO
    });

    // Delete user playlist
    //
    //Future deleteUserPlaylist(String id) async
    test('test deleteUserPlaylist', () async {
      // TODO
    });

    // Get user playlist
    //
    //Future<UserPlaylist> getUserPlaylist(String id) async
    test('test getUserPlaylist', () async {
      // TODO
    });

    // Get user playlist tracks
    //
    //Future<GetUserPlaylistTracks200Response> getUserPlaylistTracks(String id) async
    test('test getUserPlaylistTracks', () async {
      // TODO
    });

    // List user playlists
    //
    //Future<ListUserPlaylists200Response> listUserPlaylists() async
    test('test listUserPlaylists', () async {
      // TODO
    });

    // Remove track from user playlist
    //
    //Future<UserPlaylist> removeUserPlaylistTrack(String trackId, String id) async
    test('test removeUserPlaylistTrack', () async {
      // TODO
    });

    // Replace all tracks in user playlist
    //
    //Future<UserPlaylist> setUserPlaylistTracks(String id, SetUserPlaylistTracksRequest setUserPlaylistTracksRequest) async
    test('test setUserPlaylistTracks', () async {
      // TODO
    });

    // Update user playlist metadata
    //
    //Future<UserPlaylist> updateUserPlaylist(String id, UpdateUserPlaylistRequest updateUserPlaylistRequest) async
    test('test updateUserPlaylist', () async {
      // TODO
    });

  });
}
