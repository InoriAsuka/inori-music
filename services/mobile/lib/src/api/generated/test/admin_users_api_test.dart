import 'package:test/test.dart';
import 'package:inori_api/inori_api.dart';


/// tests for AdminUsersApi
void main() {
  final instance = InoriApi().getAdminUsersApi();

  group(AdminUsersApi, () {
    // Revoke all active sessions for a user
    //
    //Future<DeleteAdminUserSessions200Response> deleteAdminUserSessions(String id) async
    test('test deleteAdminUserSessions', () async {
      // TODO
    });

    // List active sessions for a user
    //
    //Future<GetAdminUserSessions200Response> getAdminUserSessions(String id) async
    test('test getAdminUserSessions', () async {
      // TODO
    });

  });
}
