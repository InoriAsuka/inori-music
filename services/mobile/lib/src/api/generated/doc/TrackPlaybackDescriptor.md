# inori_api.model.TrackPlaybackDescriptor

## Load the model package
```dart
import 'package:inori_api/api.dart';
```

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**trackId** | **String** | Catalog track identifier. | 
**mediaObjectId** | **String** | Media object identifier for the linked audio file. | 
**mimeType** | **String** | MIME type of the audio file, e.g. audio/flac. | 
**durationMs** | **int** | Track duration in milliseconds. | 
**backendId** | **String** | Storage backend identifier where the object resides. | 
**backendType** | **String** | Storage backend type (local, nfs, smb, s3, distributed). Omitted when the backend record cannot be resolved. | [optional] 
**objectKey** | **String** | Object key within the storage backend. | 
**presignedUrl** | **String** | AWS Signature Version 4 presigned GET URL, valid for 15 minutes. Present only when the backend supports presigned URLs and credentials are configured; omitted otherwise. | [optional] 
**streamUrl** | **String** | Server-proxied streaming URL for filesystem backends (local, NFS, SMB). Append ?token=<bearer> to authenticate. Present only when presignedUrl is absent and the backend supports server-side streaming. | [optional] 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


