package sei

import "strconv"

// PayloadType is the SEI message payload type as defined in Annex D of
// ITU-T Rec. H.264 (and registered extensions).
type PayloadType uint32

// Payload types defined by ITU-T Rec. H.264 Annex D and the AVC amendments.
// The list is intentionally long so that callers receive a meaningful
// String() representation even for messages this package does not parse.
const (
	PayloadTypeBufferingPeriod                       PayloadType = 0
	PayloadTypePicTiming                             PayloadType = 1
	PayloadTypePanScanRect                           PayloadType = 2
	PayloadTypeFillerPayload                         PayloadType = 3
	PayloadTypeUserDataRegisteredITUTT35             PayloadType = 4
	PayloadTypeUserDataUnregistered                  PayloadType = 5
	PayloadTypeRecoveryPoint                         PayloadType = 6
	PayloadTypeDecRefPicMarkingRepetition            PayloadType = 7
	PayloadTypeSparePic                              PayloadType = 8
	PayloadTypeSceneInfo                             PayloadType = 9
	PayloadTypeSubSeqInfo                            PayloadType = 10
	PayloadTypeSubSeqLayerCharacteristics            PayloadType = 11
	PayloadTypeSubSeqCharacteristics                 PayloadType = 12
	PayloadTypeFullFrameFreeze                       PayloadType = 13
	PayloadTypeFullFrameFreezeRelease                PayloadType = 14
	PayloadTypeFullFrameSnapshot                     PayloadType = 15
	PayloadTypeProgressiveRefinementSegmentStart     PayloadType = 16
	PayloadTypeProgressiveRefinementSegmentEnd       PayloadType = 17
	PayloadTypeMotionConstrainedSliceGroupSet        PayloadType = 18
	PayloadTypeFilmGrainCharacteristics              PayloadType = 19
	PayloadTypeDeblockingFilterDisplayPreference     PayloadType = 20
	PayloadTypeStereoVideoInfo                       PayloadType = 21
	PayloadTypePostFilterHint                        PayloadType = 22
	PayloadTypeToneMappingInfo                       PayloadType = 23
	PayloadTypeScalabilityInfo                       PayloadType = 24
	PayloadTypeSubPicScalableLayer                   PayloadType = 25
	PayloadTypeNonRequiredLayerRep                   PayloadType = 26
	PayloadTypePriorityLayerInfo                     PayloadType = 27
	PayloadTypeLayersNotPresent                      PayloadType = 28
	PayloadTypeLayerDependencyChange                 PayloadType = 29
	PayloadTypeScalableNesting                       PayloadType = 30
	PayloadTypeBaseLayerTemporalHRD                  PayloadType = 31
	PayloadTypeQualityLayerIntegrityCheck            PayloadType = 32
	PayloadTypeRedundantPicProperty                  PayloadType = 33
	PayloadTypeTL0DepRepIndex                        PayloadType = 34
	PayloadTypeTLSwitchingPoint                      PayloadType = 35
	PayloadTypeParallelDecodingInfo                  PayloadType = 36
	PayloadTypeMVCScalableNesting                    PayloadType = 37
	PayloadTypeViewScalabilityInfo                   PayloadType = 38
	PayloadTypeMultiviewSceneInfo                    PayloadType = 39
	PayloadTypeMultiviewAcquisitionInfo              PayloadType = 40
	PayloadTypeNonRequiredViewComponent              PayloadType = 41
	PayloadTypeViewDependencyChange                  PayloadType = 42
	PayloadTypeOperationPointsNotPresent             PayloadType = 43
	PayloadTypeBaseViewTemporalHRD                   PayloadType = 44
	PayloadTypeFramePackingArrangement               PayloadType = 45
	PayloadTypeMultiviewViewPosition                 PayloadType = 46
	PayloadTypeDisplayOrientation                    PayloadType = 47
	PayloadTypeMVCDViewScalabilityInfo               PayloadType = 48
	PayloadTypeDepthRepresentationInfo               PayloadType = 49
	PayloadTypeThreeDimensionalReferenceDisplaysInfo PayloadType = 50
	PayloadTypeDepthTimingInfo                       PayloadType = 51
	PayloadTypeDepthSamplingInfo                     PayloadType = 52
	PayloadTypeConstrainedDepthParameterSetIdentifier PayloadType = 53
	PayloadTypeGreenMetadata                         PayloadType = 56
	PayloadTypeMasteringDisplayColourVolume          PayloadType = 137
	PayloadTypeColourRemappingInfo                   PayloadType = 142
	PayloadTypeContentLightLevelInfo                 PayloadType = 144
	PayloadTypeAlternativeTransferCharacteristics    PayloadType = 147
	PayloadTypeAmbientViewingEnvironment             PayloadType = 148
	PayloadTypeContentColourVolume                   PayloadType = 149
)

// String returns the canonical name of the payload type from the H.264 spec,
// or a numeric fallback for unknown types.
func (p PayloadType) String() string {
	if name, ok := payloadTypeNames[p]; ok {
		return name
	}
	return "PayloadType(" + strconv.FormatUint(uint64(p), 10) + ")"
}

var payloadTypeNames = map[PayloadType]string{
	PayloadTypeBufferingPeriod:                        "buffering_period",
	PayloadTypePicTiming:                              "pic_timing",
	PayloadTypePanScanRect:                            "pan_scan_rect",
	PayloadTypeFillerPayload:                          "filler_payload",
	PayloadTypeUserDataRegisteredITUTT35:              "user_data_registered_itu_t_t35",
	PayloadTypeUserDataUnregistered:                   "user_data_unregistered",
	PayloadTypeRecoveryPoint:                          "recovery_point",
	PayloadTypeDecRefPicMarkingRepetition:             "dec_ref_pic_marking_repetition",
	PayloadTypeSparePic:                               "spare_pic",
	PayloadTypeSceneInfo:                              "scene_info",
	PayloadTypeSubSeqInfo:                             "sub_seq_info",
	PayloadTypeSubSeqLayerCharacteristics:             "sub_seq_layer_characteristics",
	PayloadTypeSubSeqCharacteristics:                  "sub_seq_characteristics",
	PayloadTypeFullFrameFreeze:                        "full_frame_freeze",
	PayloadTypeFullFrameFreezeRelease:                 "full_frame_freeze_release",
	PayloadTypeFullFrameSnapshot:                      "full_frame_snapshot",
	PayloadTypeProgressiveRefinementSegmentStart:      "progressive_refinement_segment_start",
	PayloadTypeProgressiveRefinementSegmentEnd:        "progressive_refinement_segment_end",
	PayloadTypeMotionConstrainedSliceGroupSet:         "motion_constrained_slice_group_set",
	PayloadTypeFilmGrainCharacteristics:               "film_grain_characteristics",
	PayloadTypeDeblockingFilterDisplayPreference:      "deblocking_filter_display_preference",
	PayloadTypeStereoVideoInfo:                        "stereo_video_info",
	PayloadTypePostFilterHint:                         "post_filter_hint",
	PayloadTypeToneMappingInfo:                        "tone_mapping_info",
	PayloadTypeScalabilityInfo:                        "scalability_info",
	PayloadTypeSubPicScalableLayer:                    "sub_pic_scalable_layer",
	PayloadTypeNonRequiredLayerRep:                    "non_required_layer_rep",
	PayloadTypePriorityLayerInfo:                      "priority_layer_info",
	PayloadTypeLayersNotPresent:                       "layers_not_present",
	PayloadTypeLayerDependencyChange:                  "layer_dependency_change",
	PayloadTypeScalableNesting:                        "scalable_nesting",
	PayloadTypeBaseLayerTemporalHRD:                   "base_layer_temporal_hrd",
	PayloadTypeQualityLayerIntegrityCheck:             "quality_layer_integrity_check",
	PayloadTypeRedundantPicProperty:                   "redundant_pic_property",
	PayloadTypeTL0DepRepIndex:                         "tl0_dep_rep_index",
	PayloadTypeTLSwitchingPoint:                       "tl_switching_point",
	PayloadTypeParallelDecodingInfo:                   "parallel_decoding_info",
	PayloadTypeMVCScalableNesting:                     "mvc_scalable_nesting",
	PayloadTypeViewScalabilityInfo:                    "view_scalability_info",
	PayloadTypeMultiviewSceneInfo:                     "multiview_scene_info",
	PayloadTypeMultiviewAcquisitionInfo:               "multiview_acquisition_info",
	PayloadTypeNonRequiredViewComponent:               "non_required_view_component",
	PayloadTypeViewDependencyChange:                   "view_dependency_change",
	PayloadTypeOperationPointsNotPresent:              "operation_points_not_present",
	PayloadTypeBaseViewTemporalHRD:                    "base_view_temporal_hrd",
	PayloadTypeFramePackingArrangement:                "frame_packing_arrangement",
	PayloadTypeMultiviewViewPosition:                  "multiview_view_position",
	PayloadTypeDisplayOrientation:                     "display_orientation",
	PayloadTypeMVCDViewScalabilityInfo:                "mvcd_view_scalability_info",
	PayloadTypeDepthRepresentationInfo:                "depth_representation_info",
	PayloadTypeThreeDimensionalReferenceDisplaysInfo:  "three_dimensional_reference_displays_info",
	PayloadTypeDepthTimingInfo:                        "depth_timing_info",
	PayloadTypeDepthSamplingInfo:                      "depth_sampling_info",
	PayloadTypeConstrainedDepthParameterSetIdentifier: "constrained_depth_parameter_set_identifier",
	PayloadTypeGreenMetadata:                          "green_metadata",
	PayloadTypeMasteringDisplayColourVolume:           "mastering_display_colour_volume",
	PayloadTypeColourRemappingInfo:                    "colour_remapping_info",
	PayloadTypeContentLightLevelInfo:                  "content_light_level_info",
	PayloadTypeAlternativeTransferCharacteristics:     "alternative_transfer_characteristics",
	PayloadTypeAmbientViewingEnvironment:              "ambient_viewing_environment",
	PayloadTypeContentColourVolume:                    "content_colour_volume",
}
