# inori_api.model.MediaObjectStats

## Load the model package
```dart
import 'package:inori_api/api.dart';
```

## Properties
Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**backendId** | **String** | Backend filter used to calculate this stats response, omitted for all backends. | [optional] 
**byAssetKind** | **Map&lt;String, int&gt;** |  | 
**byBackendId** | **Map&lt;String, int&gt;** |  | 
**byLifecycleState** | **Map&lt;String, int&gt;** |  | 
**byVerificationStatus** | **Map&lt;String, int&gt;** | Counts by latest verification state. Includes verified, failed, and unknown buckets. | 
**totalObjects** | **int** |  | 
**totalSizeBytes** | **int** |  | 

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


