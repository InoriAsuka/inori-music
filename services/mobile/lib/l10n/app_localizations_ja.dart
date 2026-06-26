// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Japanese (`ja`).
class AppLocalizationsJa extends AppLocalizations {
  AppLocalizationsJa([String locale = 'ja']) : super(locale);

  @override
  String get appTitle => 'Inori Music';

  @override
  String get login => 'ログイン';

  @override
  String get logout => 'ログアウト';

  @override
  String get username => 'ユーザー名';

  @override
  String get password => 'パスワード';

  @override
  String get loginError => 'ユーザー名またはパスワードが間違っています';

  @override
  String get home => 'ホーム';

  @override
  String get artists => 'アーティスト';

  @override
  String get albums => 'アルバム';

  @override
  String get tracks => 'トラック';

  @override
  String get playlists => 'プレイリスト';

  @override
  String get search => '検索';

  @override
  String get library => 'ライブラリ';

  @override
  String get favorites => 'お気に入り';

  @override
  String get history => '再生履歴';

  @override
  String get settings => '設定';

  @override
  String get nowPlaying => '再生中';

  @override
  String get queue => 'キュー';

  @override
  String get shuffle => 'シャッフル';

  @override
  String get repeat => 'リピート';

  @override
  String get play => '再生';

  @override
  String get pause => '一時停止';

  @override
  String get next => '次へ';

  @override
  String get previous => '前へ';

  @override
  String get addToQueue => 'キューに追加';

  @override
  String get playNext => '次に再生';

  @override
  String get noResults => '結果が見つかりませんでした';

  @override
  String get loading => '読み込み中…';

  @override
  String get error => 'エラーが発生しました';

  @override
  String get retry => '再試行';

  @override
  String get changePassword => 'パスワード変更';

  @override
  String get currentPassword => '現在のパスワード';

  @override
  String get newPassword => '新しいパスワード';

  @override
  String get confirmPassword => 'パスワードを確認';

  @override
  String get save => '保存';

  @override
  String get cancel => 'キャンセル';

  @override
  String get delete => '削除';

  @override
  String get deleteAll => 'すべて削除';

  @override
  String get sessions => 'セッション管理';

  @override
  String get revokeAll => 'すべてのセッションを失効';

  @override
  String get revokeAllConfirmBody => 'すべてのデバイスからサインアウトします。続行しますか？';

  @override
  String get fieldRequired => '必須です';

  @override
  String get passwordMinLength => '8文字以上';

  @override
  String get changePasswordSuccess => 'パスワードが変更されました';

  @override
  String get signOutConfirmBody => 'サインアウトしてもよろしいですか？';

  @override
  String get language => '言語';

  @override
  String get theme => 'テーマ';

  @override
  String get darkMode => 'ダークモード';

  @override
  String get nothingPlaying => '再生していません';

  @override
  String get noFavoritesYet => 'お気に入りがありません';

  @override
  String get historyStats => '再生統計';

  @override
  String get topTracks => 'よく聴くトラック';

  @override
  String get activityChart => '30日間の再生記録';

  @override
  String get totalPlays => '合計再生回数';

  @override
  String get uniqueTracks => 'ユニークトラック数';

  @override
  String get today => '今日';

  @override
  String get yesterday => '昨日';

  @override
  String daysAgo(int days) {
    return '$days日前';
  }

  @override
  String get searchHint => 'アーティスト・アルバム・トラックを検索…';

  @override
  String get searchPrompt => 'キーワードを入力して検索…';

  @override
  String get deleteHistory => '履歴を削除';

  @override
  String get noHistory => '再生履歴がありません';

  @override
  String get noData => 'データがありません';
}
