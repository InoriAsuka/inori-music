//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'lyrics_response.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class LyricsResponse {
  /// Returns a new [LyricsResponse] instance.
  LyricsResponse({

    required  this.format,

    required  this.content,

     this.mediaObjectId,

     this.translation,

     this.source_,

     this.translationMediaObjectId,
  });

      /// Lyrics file format
  @JsonKey(
    
    name: r'format',
    required: true,
    includeIfNull: false,
  )


  final LyricsResponseFormatEnum format;



      /// Raw lyrics file content (UTF-8)
  @JsonKey(
    
    name: r'content',
    required: true,
    includeIfNull: false,
  )


  final String content;



  @JsonKey(
    
    name: r'mediaObjectId',
    required: false,
    includeIfNull: false,
  )


  final String? mediaObjectId;



      /// Raw translation file content (UTF-8), present when a translation has been uploaded
  @JsonKey(
    
    name: r'translation',
    required: false,
    includeIfNull: false,
  )


  final String? translation;



      /// Where the lyrics content originated
  @JsonKey(
    
    name: r'source',
    required: false,
    includeIfNull: false,
  )


  final LyricsResponseSource_Enum? source_;



  @JsonKey(
    
    name: r'translationMediaObjectId',
    required: false,
    includeIfNull: false,
  )


  final String? translationMediaObjectId;





    @override
    bool operator ==(Object other) => identical(this, other) || other is LyricsResponse &&
      other.format == format &&
      other.content == content &&
      other.mediaObjectId == mediaObjectId &&
      other.translation == translation &&
      other.source_ == source_ &&
      other.translationMediaObjectId == translationMediaObjectId;

    @override
    int get hashCode =>
        format.hashCode +
        content.hashCode +
        mediaObjectId.hashCode +
        translation.hashCode +
        source_.hashCode +
        translationMediaObjectId.hashCode;

  factory LyricsResponse.fromJson(Map<String, dynamic> json) => _$LyricsResponseFromJson(json);

  Map<String, dynamic> toJson() => _$LyricsResponseToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

/// Lyrics file format
enum LyricsResponseFormatEnum {
    /// Lyrics file format
@JsonValue(r'lrc')
lrc(r'lrc'),
    /// Lyrics file format
@JsonValue(r'srt')
srt(r'srt');

const LyricsResponseFormatEnum(this.value);

final String value;

@override
String toString() => value;
}


/// Where the lyrics content originated
enum LyricsResponseSource_Enum {
    /// Where the lyrics content originated
@JsonValue(r'embedded')
embedded(r'embedded'),
    /// Where the lyrics content originated
@JsonValue(r'manual')
manual(r'manual'),
    /// Where the lyrics content originated
@JsonValue(r'lrclib')
lrclib(r'lrclib');

const LyricsResponseSource_Enum(this.value);

final String value;

@override
String toString() => value;
}


