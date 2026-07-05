// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appTitle => 'Inori Music';

  @override
  String get login => 'Sign In';

  @override
  String get logout => 'Sign Out';

  @override
  String get username => 'Username';

  @override
  String get password => 'Password';

  @override
  String get loginError => 'Invalid username or password';

  @override
  String get home => 'Home';

  @override
  String get artists => 'Artists';

  @override
  String get albums => 'Albums';

  @override
  String get tracks => 'Tracks';

  @override
  String get playlists => 'Playlists';

  @override
  String get search => 'Search';

  @override
  String get library => 'Library';

  @override
  String get favorites => 'Favorites';

  @override
  String get history => 'History';

  @override
  String get settings => 'Settings';

  @override
  String get nowPlaying => 'Now Playing';

  @override
  String get queue => 'Queue';

  @override
  String get shuffle => 'Shuffle';

  @override
  String get repeat => 'Repeat';

  @override
  String get play => 'Play';

  @override
  String get pause => 'Pause';

  @override
  String get next => 'Next';

  @override
  String get previous => 'Previous';

  @override
  String get addToQueue => 'Add to Queue';

  @override
  String get playNext => 'Play Next';

  @override
  String get noResults => 'No results found';

  @override
  String get loading => 'Loading…';

  @override
  String get error => 'Something went wrong';

  @override
  String get retry => 'Retry';

  @override
  String get changePassword => 'Change Password';

  @override
  String get currentPassword => 'Current Password';

  @override
  String get newPassword => 'New Password';

  @override
  String get confirmPassword => 'Confirm Password';

  @override
  String get save => 'Save';

  @override
  String get cancel => 'Cancel';

  @override
  String get delete => 'Delete';

  @override
  String get deleteAll => 'Delete All';

  @override
  String get sessions => 'Sessions';

  @override
  String get revokeAll => 'Revoke All Sessions';

  @override
  String get revokeAllConfirmBody =>
      'This will sign you out from all devices. Continue?';

  @override
  String get fieldRequired => 'Required';

  @override
  String get passwordMinLength => 'Minimum 8 characters';

  @override
  String get changePasswordSuccess => 'Password changed successfully';

  @override
  String get signOutConfirmBody => 'Are you sure you want to sign out?';

  @override
  String get language => 'Language';

  @override
  String get theme => 'Theme';

  @override
  String get darkMode => 'Dark Mode';

  @override
  String get nothingPlaying => 'Nothing playing';

  @override
  String get noFavoritesYet => 'No favorites yet';

  @override
  String get historyStats => 'History Stats';

  @override
  String get topTracks => 'Top Tracks';

  @override
  String get activityChart => '30-day Activity';

  @override
  String get totalPlays => 'Total Plays';

  @override
  String get uniqueTracks => 'Unique Tracks';

  @override
  String get today => 'Today';

  @override
  String get yesterday => 'Yesterday';

  @override
  String daysAgo(int days) {
    return '$days days ago';
  }

  @override
  String get searchHint => 'Search artists, albums, tracks…';

  @override
  String get searchPrompt => 'Start typing to search…';

  @override
  String get serverUrl => 'Server URL';

  @override
  String get deleteHistory => 'Delete History';

  @override
  String get noHistory => 'No history yet';

  @override
  String get noData => 'No data';

  @override
  String get recentSearches => 'Recent Searches';

  @override
  String get clearSearchHistory => 'Clear search history';
}
