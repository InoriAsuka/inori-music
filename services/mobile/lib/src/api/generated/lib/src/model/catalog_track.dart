//
// AUTO-GENERATED FILE, DO NOT MODIFY!
//

// ignore_for_file: unused_element
import 'package:copy_with_extension/copy_with_extension.dart';
import 'package:json_annotation/json_annotation.dart';

part 'catalog_track.g.dart';


@CopyWith()
@JsonSerializable(
  checked: true,
  createToJson: true,
  disallowUnrecognizedKeys: false,
  explicitToJson: true,
)
class CatalogTrack {
  /// Returns a new [CatalogTrack] instance.
  CatalogTrack({

     this.albumId,

    required  this.artistId,

    required  this.createdAt,

     this.discNumber,

     this.durationMs,

    required  this.id,

    required  this.mediaObjectId,

     this.sortTitle,

    required  this.title,

     this.trackNumber,

    required  this.updatedAt,

     this.genre,

     this.isFavorite = false,

     this.lyricsMediaObjectId,

     this.replayGainDb,
  });

  @JsonKey(
    
    name: r'albumId',
    required: false,
    includeIfNull: false,
  )


  final String? albumId;



  @JsonKey(
    
    name: r'artistId',
    required: true,
    includeIfNull: false,
  )


  final String artistId;



  @JsonKey(
    
    name: r'createdAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime createdAt;



          // minimum: 0
  @JsonKey(
    
    name: r'discNumber',
    required: false,
    includeIfNull: false,
  )


  final int? discNumber;



          // minimum: 0
  @JsonKey(
    
    name: r'durationMs',
    required: false,
    includeIfNull: false,
  )


  final int? durationMs;



  @JsonKey(
    
    name: r'id',
    required: true,
    includeIfNull: false,
  )


  final String id;



  @JsonKey(
    
    name: r'mediaObjectId',
    required: true,
    includeIfNull: false,
  )


  final String mediaObjectId;



  @JsonKey(
    
    name: r'sortTitle',
    required: false,
    includeIfNull: false,
  )


  final String? sortTitle;



  @JsonKey(
    
    name: r'title',
    required: true,
    includeIfNull: false,
  )


  final String title;



          // minimum: 0
  @JsonKey(
    
    name: r'trackNumber',
    required: false,
    includeIfNull: false,
  )


  final int? trackNumber;



  @JsonKey(
    
    name: r'updatedAt',
    required: true,
    includeIfNull: false,
  )


  final DateTime updatedAt;



      /// Optional genre tag for the track (e.g. Rock, Jazz, Classical)
  @JsonKey(
    
    name: r'genre',
    required: false,
    includeIfNull: false,
  )


  final String? genre;



      /// Whether the authenticated viewer has favorited this track. Always false on admin endpoints.
  @JsonKey(
    defaultValue: false,
    name: r'isFavorite',
    required: false,
    includeIfNull: false,
  )


  final bool? isFavorite;



      /// ID of the media object used as lyrics (optional)
  @JsonKey(
    
    name: r'lyricsMediaObjectId',
    required: false,
    includeIfNull: false,
  )


  final String? lyricsMediaObjectId;



      /// ReplayGain normalization value in dB (null = not analyzed)
  @JsonKey(
    
    name: r'replayGainDb',
    required: false,
    includeIfNull: false,
  )


  final double? replayGainDb;





    @override
    bool operator ==(Object other) => identical(this, other) || other is CatalogTrack &&
      other.albumId == albumId &&
      other.artistId == artistId &&
      other.createdAt == createdAt &&
      other.discNumber == discNumber &&
      other.durationMs == durationMs &&
      other.id == id &&
      other.mediaObjectId == mediaObjectId &&
      other.sortTitle == sortTitle &&
      other.title == title &&
      other.trackNumber == trackNumber &&
      other.updatedAt == updatedAt &&
      other.genre == genre &&
      other.isFavorite == isFavorite &&
      other.lyricsMediaObjectId == lyricsMediaObjectId &&
      other.replayGainDb == replayGainDb;

    @override
    int get hashCode =>
        albumId.hashCode +
        artistId.hashCode +
        createdAt.hashCode +
        discNumber.hashCode +
        durationMs.hashCode +
        id.hashCode +
        mediaObjectId.hashCode +
        sortTitle.hashCode +
        title.hashCode +
        trackNumber.hashCode +
        updatedAt.hashCode +
        genre.hashCode +
        isFavorite.hashCode +
        lyricsMediaObjectId.hashCode +
        replayGainDb.hashCode;

  factory CatalogTrack.fromJson(Map<String, dynamic> json) => _$CatalogTrackFromJson(json);

  Map<String, dynamic> toJson() => _$CatalogTrackToJson(this);

  @override
  String toString() {
    return toJson().toString();
  }

}

