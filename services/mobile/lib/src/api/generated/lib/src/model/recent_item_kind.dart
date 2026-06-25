//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:json_annotation/json_annotation.dart';

/// Catalog entity kind for a recent timeline item.
enum RecentItemKind {
          /// Catalog entity kind for a recent timeline item.
      @JsonValue(r'artist')
      artist(r'artist'),
          /// Catalog entity kind for a recent timeline item.
      @JsonValue(r'album')
      album(r'album'),
          /// Catalog entity kind for a recent timeline item.
      @JsonValue(r'track')
      track(r'track'),
          /// Catalog entity kind for a recent timeline item.
      @JsonValue(r'playlist')
      playlist(r'playlist');

  const RecentItemKind(this.value);

  final String value;

  @override
  String toString() => value;
}
