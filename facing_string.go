// Code generated by "stringer -type=Facing"; DO NOT EDIT.

package molecular

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[TOP-0]
	_ = x[BOTTOM-1]
	_ = x[LEFT-2]
	_ = x[RIGHT-3]
	_ = x[FRONT-4]
	_ = x[BACK-5]
}

const _Facing_name = "TOPBOTTOMLEFTRIGHTFRONTBACK"

var _Facing_index = [...]uint8{0, 3, 9, 13, 18, 23, 27}

func (i Facing) String() string {
	if i >= Facing(len(_Facing_index)-1) {
		return "Facing(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _Facing_name[_Facing_index[i]:_Facing_index[i+1]]
}