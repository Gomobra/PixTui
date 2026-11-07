package native

// RuneWidth returns the display width of a rune (simplified wcwidth)
func RuneWidth(r rune) int {
	if r < 32 || r == 127 {
		return 0
	}
	if r >= 0x1100 && r <= 0x115f {
		return 2
	}
	if r >= 0x11a3 && r <= 0x11a7 {
		return 2
	}
	if r >= 0x11fa && r <= 0x11ff {
		return 2
	}
	if r >= 0x231a && r <= 0x231b {
		return 2
	}
	if r >= 0x2329 && r <= 0x232a {
		return 2
	}
	if r >= 0x23cf && r <= 0x23cf {
		return 2
	}
	if r >= 0x23e9 && r <= 0x23f3 {
		return 2
	}
	if r >= 0x23f8 && r <= 0x23fa {
		return 2
	}
	if r >= 0x2e80 && r <= 0x2e99 {
		return 2
	}
	if r >= 0x2e9b && r <= 0x2ef3 {
		return 2
	}
	if r >= 0x2f00 && r <= 0x2fd5 {
		return 2
	}
	if r >= 0x2ff0 && r <= 0x2ffb {
		return 2
	}
	if r >= 0x3000 && r <= 0x303e {
		return 2
	}
	if r >= 0x3041 && r <= 0x3096 {
		return 2
	}
	if r >= 0x3099 && r <= 0x309a {
		return 0
	}
	if r >= 0x309d && r <= 0x309e {
		return 2
	}
	if r >= 0x30a0 && r <= 0x30fa {
		return 2
	}
	if r >= 0x30fc && r <= 0x30ff {
		return 2
	}
	if r >= 0x3105 && r <= 0x312d {
		return 2
	}
	if r >= 0x3131 && r <= 0x318e {
		return 2
	}
	if r >= 0x3190 && r <= 0x31ba {
		return 2
	}
	if r >= 0x31c0 && r <= 0x31e3 {
		return 2
	}
	if r >= 0x31f0 && r <= 0x321e {
		return 2
	}
	if r >= 0x3220 && r <= 0x3247 {
		return 2
	}
	if r >= 0x3250 && r <= 0x32fe {
		return 2
	}
	if r >= 0x3300 && r <= 0x4dbf {
		return 2
	}
	if r >= 0x4e00 && r <= 0xa48c {
		return 2
	}
	if r >= 0xa490 && r <= 0xa4c6 {
		return 2
	}
	if r >= 0xa960 && r <= 0xa97c {
		return 2
	}
	if r >= 0xac00 && r <= 0xd7a3 {
		return 2
	}
	if r >= 0xd7b0 && r <= 0xd7c6 {
		return 2
	}
	if r >= 0xd7cb && r <= 0xd7fb {
		return 2
	}
	if r >= 0xf900 && r <= 0xfaff {
		return 2
	}
	if r >= 0xfe10 && r <= 0xfe19 {
		return 2
	}
	if r >= 0xfe30 && r <= 0xfe6b {
		return 2
	}
	if r >= 0xff01 && r <= 0xff60 {
		return 2
	}
	if r >= 0xff8c && r <= 0xffbf {
		return 2
	}
	if r >= 0xffe0 && r <= 0xffe6 {
		return 2
	}
	if r >= 0xffe8 && r <= 0xffe9 {
		return 2
	}
	if r >= 0xffed && r <= 0xffed {
		return 2
	}
	if r >= 0x1b000 && r <= 0x1b001 {
		return 2
	}
	if r >= 0x1f200 && r <= 0x1f202 {
		return 2
	}
	if r >= 0x1f210 && r <= 0x1f23a {
		return 2
	}
	if r >= 0x1f240 && r <= 0x1f248 {
		return 2
	}
	if r >= 0x1f250 && r <= 0x1f251 {
		return 2
	}
	if r >= 0x1f300 && r <= 0x1f320 {
		return 2
	}
	if r >= 0x1f32d && r <= 0x1f335 {
		return 2
	}
	if r >= 0x1f337 && r <= 0x1f37c {
		return 2
	}
	if r >= 0x1f37e && r <= 0x1f393 {
		return 2
	}
	if r >= 0x1f3a0 && r <= 0x1f3ca {
		return 2
	}
	if r >= 0x1f3cd && r <= 0x1f3cd {
		return 2
	}
	if r >= 0x1f3ce && r <= 0x1f3ce {
		return 2
	}
	if r >= 0x1f3cf && r <= 0x1f3cf {
		return 2
	}
	if r >= 0x1f3d0 && r <= 0x1f3e3 {
		return 2
	}
	if r >= 0x1f3e4 && r <= 0x1f3e4 {
		return 2
	}
	if r >= 0x1f3e5 && r <= 0x1f3e5 {
		return 2
	}
	if r >= 0x1f3e6 && r <= 0x1f3e6 {
		return 2
	}
	if r >= 0x1f3e7 && r <= 0x1f3e7 {
		return 2
	}
	if r >= 0x1f3e8 && r <= 0x1f3e8 {
		return 2
	}
	if r >= 0x1f3e9 && r <= 0x1f3e9 {
		return 2
	}
	if r >= 0x1f3ea && r <= 0x1f3ea {
		return 2
	}
	if r >= 0x1f3eb && r <= 0x1f3eb {
		return 2
	}
	if r >= 0x1f3ec && r <= 0x1f3ec {
		return 2
	}
	if r >= 0x1f3ed && r <= 0x1f3ed {
		return 2
	}
	if r >= 0x1f3ee && r <= 0x1f3ee {
		return 2
	}
	if r >= 0x1f3f0 && r <= 0x1f3f0 {
		return 2
	}
	if r >= 0x1f3f3 && r <= 0x1f3f3 {
		return 2
	}
	if r >= 0x1f3f4 && r <= 0x1f3f4 {
		return 2
	}
	if r >= 0x1f3f5 && r <= 0x1f3f5 {
		return 2
	}
	if r >= 0x1f3f6 && r <= 0x1f3f6 {
		return 2
	}
	if r >= 0x1f3f7 && r <= 0x1f3f7 {
		return 2
	}
	if r >= 0x1f3f8 && r <= 0x1f3f8 {
		return 2
	}
	if r >= 0x1f3f9 && r <= 0x1f3f9 {
		return 2
	}
	if r >= 0x1f3fa && r <= 0x1f3fa {
		return 2
	}
	if r >= 0x1f3fb && r <= 0x1f3fb {
		return 2
	}
	if r >= 0x1f3fc && r <= 0x1f3fc {
		return 2
	}
	if r >= 0x1f3fd && r <= 0x1f3fd {
		return 2
	}
	if r >= 0x1f3fe && r <= 0x1f3fe {
		return 2
	}
	if r >= 0x1f3ff && r <= 0x1f3ff {
		return 2
	}
	if r >= 0x1f400 && r <= 0x1f43e {
		return 2
	}
	if r >= 0x1f440 && r <= 0x1f440 {
		return 2
	}
	if r >= 0x1f442 && r <= 0x1f4fc {
		return 2
	}
	if r >= 0x1f4fd && r <= 0x1f4ff {
		return 2
	}
	if r >= 0x1f500 && r <= 0x1f53d {
		return 2
	}
	if r >= 0x1f53e && r <= 0x1f544 {
		return 2
	}
	if r >= 0x1f545 && r <= 0x1f545 {
		return 2
	}
	if r >= 0x1f546 && r <= 0x1f546 {
		return 2
	}
	if r >= 0x1f547 && r <= 0x1f547 {
		return 2
	}
	if r >= 0x1f548 && r <= 0x1f548 {
		return 2
	}
	if r >= 0x1f549 && r <= 0x1f549 {
		return 2
	}
	if r >= 0x1f54a && r <= 0x1f54a {
		return 2
	}
	if r >= 0x1f54b && r <= 0x1f54b {
		return 2
	}
	if r >= 0x1f54c && r <= 0x1f54c {
		return 2
	}
	if r >= 0x1f54d && r <= 0x1f54d {
		return 2
	}
	if r >= 0x1f54e && r <= 0x1f54e {
		return 2
	}
	if r >= 0x1f54f && r <= 0x1f54f {
		return 2
	}
	if r >= 0x1f550 && r <= 0x1f567 {
		return 2
	}
	if r >= 0x1f568 && r <= 0x1f568 {
		return 2
	}
	if r >= 0x1f569 && r <= 0x1f569 {
		return 2
	}
	if r >= 0x1f56a && r <= 0x1f56a {
		return 2
	}
	if r >= 0x1f56b && r <= 0x1f56b {
		return 2
	}
	if r >= 0x1f56c && r <= 0x1f56c {
		return 2
	}
	if r >= 0x1f56d && r <= 0x1f56d {
		return 2
	}
	if r >= 0x1f56e && r <= 0x1f56e {
		return 2
	}
	if r >= 0x1f56f && r <= 0x1f56f {
		return 2
	}
	if r >= 0x1f570 && r <= 0x1f570 {
		return 2
	}
	if r >= 0x1f571 && r <= 0x1f571 {
		return 2
	}
	if r >= 0x1f572 && r <= 0x1f572 {
		return 2
	}
	if r >= 0x1f573 && r <= 0x1f573 {
		return 2
	}
	if r >= 0x1f574 && r <= 0x1f574 {
		return 2
	}
	if r >= 0x1f575 && r <= 0x1f575 {
		return 2
	}
	if r >= 0x1f576 && r <= 0x1f576 {
		return 2
	}
	if r >= 0x1f577 && r <= 0x1f577 {
		return 2
	}
	if r >= 0x1f578 && r <= 0x1f578 {
		return 2
	}
	if r >= 0x1f579 && r <= 0x1f579 {
		return 2
	}
	if r >= 0x1f57a && r <= 0x1f57a {
		return 2
	}
	if r >= 0x1f57b && r <= 0x1f57b {
		return 2
	}
	if r >= 0x1f57c && r <= 0x1f57c {
		return 2
	}
	if r >= 0x1f57d && r <= 0x1f57d {
		return 2
	}
	if r >= 0x1f57e && r <= 0x1f57e {
		return 2
	}
	if r >= 0x1f57f && r <= 0x1f57f {
		return 2
	}
	if r >= 0x1f580 && r <= 0x1f580 {
		return 2
	}
	if r >= 0x1f581 && r <= 0x1f581 {
		return 2
	}
	if r >= 0x1f582 && r <= 0x1f582 {
		return 2
	}
	if r >= 0x1f583 && r <= 0x1f583 {
		return 2
	}
	if r >= 0x1f584 && r <= 0x1f584 {
		return 2
	}
	if r >= 0x1f585 && r <= 0x1f585 {
		return 2
	}
	if r >= 0x1f586 && r <= 0x1f586 {
		return 2
	}
	if r >= 0x1f587 && r <= 0x1f587 {
		return 2
	}
	if r >= 0x1f588 && r <= 0x1f588 {
		return 2
	}
	if r >= 0x1f589 && r <= 0x1f589 {
		return 2
	}
	if r >= 0x1f58a && r <= 0x1f58a {
		return 2
	}
	if r >= 0x1f58b && r <= 0x1f58b {
		return 2
	}
	if r >= 0x1f58c && r <= 0x1f58c {
		return 2
	}
	if r >= 0x1f58d && r <= 0x1f58d {
		return 2
	}
	if r >= 0x1f58e && r <= 0x1f58e {
		return 2
	}
	if r >= 0x1f58f && r <= 0x1f58f {
		return 2
	}
	if r >= 0x1f590 && r <= 0x1f590 {
		return 2
	}
	if r >= 0x1f591 && r <= 0x1f591 {
		return 2
	}
	if r >= 0x1f592 && r <= 0x1f592 {
		return 2
	}
	if r >= 0x1f593 && r <= 0x1f593 {
		return 2
	}
	if r >= 0x1f594 && r <= 0x1f594 {
		return 2
	}
	if r >= 0x1f595 && r <= 0x1f595 {
		return 2
	}
	if r >= 0x1f596 && r <= 0x1f596 {
		return 2
	}
	if r >= 0x1f597 && r <= 0x1f597 {
		return 2
	}
	if r >= 0x1f598 && r <= 0x1f598 {
		return 2
	}
	if r >= 0x1f599 && r <= 0x1f599 {
		return 2
	}
	if r >= 0x1f59a && r <= 0x1f59a {
		return 2
	}
	if r >= 0x1f59b && r <= 0x1f59b {
		return 2
	}
	if r >= 0x1f59c && r <= 0x1f59c {
		return 2
	}
	if r >= 0x1f59d && r <= 0x1f59d {
		return 2
	}
	if r >= 0x1f59e && r <= 0x1f59e {
		return 2
	}
	if r >= 0x1f59f && r <= 0x1f59f {
		return 2
	}
	if r >= 0x1f5a0 && r <= 0x1f5a0 {
		return 2
	}
	if r >= 0x1f5a1 && r <= 0x1f5a1 {
		return 2
	}
	if r >= 0x1f5a2 && r <= 0x1f5a2 {
		return 2
	}
	if r >= 0x1f5a3 && r <= 0x1f5a3 {
		return 2
	}
	if r >= 0x1f5a4 && r <= 0x1f5a4 {
		return 2
	}
	if r >= 0x1f5a5 && r <= 0x1f5a5 {
		return 2
	}
	if r >= 0x1f5a6 && r <= 0x1f5a6 {
		return 2
	}
	if r >= 0x1f5a7 && r <= 0x1f5a7 {
		return 2
	}
	if r >= 0x1f5a8 && r <= 0x1f5a8 {
		return 2
	}
	if r >= 0x1f5a9 && r <= 0x1f5a9 {
		return 2
	}
	if r >= 0x1f5aa && r <= 0x1f5aa {
		return 2
	}
	if r >= 0x1f5ab && r <= 0x1f5ab {
		return 2
	}
	if r >= 0x1f5ac && r <= 0x1f5ac {
		return 2
	}
	if r >= 0x1f5ad && r <= 0x1f5ad {
		return 2
	}
	if r >= 0x1f5ae && r <= 0x1f5ae {
		return 2
	}
	if r >= 0x1f5af && r <= 0x1f5af {
		return 2
	}
	if r >= 0x1f5b0 && r <= 0x1f5b0 {
		return 2
	}
	if r >= 0x1f5b1 && r <= 0x1f5b1 {
		return 2
	}
	if r >= 0x1f5b2 && r <= 0x1f5b2 {
		return 2
	}
	if r >= 0x1f5b3 && r <= 0x1f5b3 {
		return 2
	}
	if r >= 0x1f5b4 && r <= 0x1f5b4 {
		return 2
	}
	if r >= 0x1f5b5 && r <= 0x1f5b5 {
		return 2
	}
	if r >= 0x1f5b6 && r <= 0x1f5b6 {
		return 2
	}
	if r >= 0x1f5b7 && r <= 0x1f5b7 {
		return 2
	}
	if r >= 0x1f5b8 && r <= 0x1f5b8 {
		return 2
	}
	if r >= 0x1f5b9 && r <= 0x1f5b9 {
		return 2
	}
	if r >= 0x1f5ba && r <= 0x1f5ba {
		return 2
	}
	if r >= 0x1f5bb && r <= 0x1f5bb {
		return 2
	}
	if r >= 0x1f5bc && r <= 0x1f5bc {
		return 2
	}
	if r >= 0x1f5bd && r <= 0x1f5bd {
		return 2
	}
	if r >= 0x1f5be && r <= 0x1f5be {
		return 2
	}
	if r >= 0x1f5bf && r <= 0x1f5bf {
		return 2
	}
	if r >= 0x1f5c0 && r <= 0x1f5c0 {
		return 2
	}
	if r >= 0x1f5c1 && r <= 0x1f5c1 {
		return 2
	}
	if r >= 0x1f5c2 && r <= 0x1f5c2 {
		return 2
	}
	if r >= 0x1f5c3 && r <= 0x1f5c3 {
		return 2
	}
	if r >= 0x1f5c4 && r <= 0x1f5c4 {
		return 2
	}
	return 1
}
