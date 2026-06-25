import 'package:test/test.dart';
import 'package:inori_api/inori_api.dart';


/// tests for MeApi
void main() {
  final instance = InoriApi().getMeApi();

  group(MeApi, () {
    // List active sessions for the authenticated user
    //
    //Future<GetAdminUserSessions200Response> getMyActiveSessions() async
    test('test getMyActiveSessions', () async {
      // TODO
    });

    // Revoke all sessions except the current one
    //
    //Future<DeleteAdminUserSessions200Response> revokeMyOtherSessions() async
    test('test revokeMyOtherSessions', () async {
      // TODO
    });

  });
}
