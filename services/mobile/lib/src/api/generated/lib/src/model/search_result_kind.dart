//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:json_annotation/json_annotation.dart';

/// The entity kind of a catalog search result item.
enum SearchResultKind {
          /// The entity kind of a catalog search result item.
      @JsonValue(r'artist')
      artist(r'artist'),
          /// The entity kind of a catalog search result item.
      @JsonValue(r'album')
      album(r'album'),
          /// The entity kind of a catalog search result item.
      @JsonValue(r'track')
      track(r'track');

  const SearchResultKind(this.value);

  final String value;

  @override
  String toString() => value;
}
