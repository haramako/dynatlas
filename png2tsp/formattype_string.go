// Code generated by "stringer -type=FormatType"; DO NOT EDIT

package main

import "fmt"

const _FormatType_name = "BothPVRTCETC1PVRTC_SPLIT_ALPHAETC1_SPLIT_ALPHA"

var _FormatType_index = [...]uint8{0, 4, 9, 13, 30, 46}

func (i FormatType) String() string {
	if i < 0 || i >= FormatType(len(_FormatType_index)-1) {
		return fmt.Sprintf("FormatType(%d)", i)
	}
	return _FormatType_name[_FormatType_index[i]:_FormatType_index[i+1]]
}
