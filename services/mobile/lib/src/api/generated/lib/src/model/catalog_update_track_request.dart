//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_update_track_request.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogUpdateTrackRequest {
  /// Returns a new [CatalogUpdateTrackRequest] instance.
  CatalogUpdateTrackRequest({

     this.title,

     this.sortTitle,

     this.artistId,

     this.albumId,

     this.trackNumber,

     this.discNumber,

     this.durationMs,

     this.genre,

     this.replayGainDb,
  });

      /// Track title. Must not be empty if provided.
  @JsonKey(
    
    name: r'title',
    required: false,
    includeIfNull: false,
  )


  final String? title;



      /// Sort key for the track title. May be empty to clear.
  @JsonKey(
    
    name: r'sortTitle',
    required: false,
    includeIfNull: false,
  )


  final String? sortTitle;



      /// ID of the performing artist. Must reference an existing artist.
  @JsonKey(
    
    name: r'artistId',
    required: false,
    includeIfNull: false,
  )


  final String? artistId;



      /// ID of the parent album. Empty string removes album association.
  @JsonKey(
    
    name: r'albumId',
    required: false,
    includeIfNull: false,
  )


  final String? albumId;



          // minimum: 0
  @JsonKey(
    
    name: r'trackNumber',
    required: false,
    includeIfNull: false,
  )


  final int? trackNumber;



          // minimum: 0
  @JsonKey(
    
    name: r'discNumber',
    required: false,
    includeIfNull: false,
  )


  final int? discNumber;



      /// Track duration in milliseconds.
          // minimum: 0
  @JsonKey(
    
    name: r'durationMs',
    required: false,
    includeIfNull: false,
  )


  final int? durationMs;



  @JsonKey(
    
    name: r'genre',
    required: false,
    includeIfNull: false,
  )


  final String? genre;



  @JsonKey(
    
    name: r'replayGainDb',
    required: false,
    includeIfNull: false,
  )


  final double? replayGainDb;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogUpdateTrackRequest &&
      other.title == title &&
      other.sortTitle == sortTitle &&
      other.artistId == artistId &&
      other.albumId == albumId &&
      other.trackNumber == trackNumber &&
      other.discNumber == discNumber &&
      other.durationMs == durationMs &&
      other.genre == genre &&
      other.replayGainDb == replayGainDb;

    @override
    int get hashCode =>
        title.hashCode +
        sortTitle.hashCode +
        artistId.hashCode +
        albumId.hashCode +
        trackNumber.hashCode +
        discNumber.hashCode +
        durationMs.hashCode +
        genre.hashCode +
        replayGainDb.hashCode;

  factory CatalogUpdateTrackRequest.fromJson(Map<String, dynamic> json) => _$CatalogUpdateTrackRequestFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogUpdateTrackRequestToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

