package uuid

type formatDetails struct {
	peakCount   uint
	usedCount   uint
	padCount    uint
	needPrePad  bool
	needPostPad bool
	hasSecond   bool
	first       bits
	second      bits
}

func formatStudy(verb rune, hasPlus, hasSharp, hasMinus, hasWidth bool, width uint, x bits) formatDetails {
	var d formatDetails

	y := x.just(bitValid | bitTextIsDense | bitTextIsModeY | bitTextIsModeX)
	switch verb {
	case 'd':
		d.first = bitValid | bitTextIsDense

	case 's', 'q':
		if hasPlus {
			d.first = bitValid | bitTextIsDense
		} else if hasSharp {
			d.first = bitValid | textModeCanonical
		} else {
			d.first = y
		}
		if verb == 'q' {
			d.first |= bitReserved
		}

	case 'v':
		if hasPlus {
			d.first = bitValid | bitTextIsDense | bitReserved
			d.second = bitValid | textModeBracketed
			d.hasSecond = true
		} else if hasSharp {
			d.first = bitValid | textModeCanonical
		} else {
			d.first = y
		}

	default:
		d.first = y
	}

	need := func(n uint) {
		d.peakCount = d.usedCount + n
	}

	use := func(n uint) {
		d.usedCount += n
		if d.peakCount < d.usedCount {
			d.peakCount = d.usedCount
		}
	}

	add := func(x bits) {
		used, peak := x.textLength()
		need(peak)
		use(used)
	}

	if d.first.has(bitReserved) {
		use(1)
	}
	add(d.first)
	if d.first.has(bitReserved) {
		use(1)
	}
	if d.hasSecond {
		use(1)
		add(d.second)
	}
	if hasWidth && width > d.usedCount {
		d.padCount = width - d.usedCount
		if hasMinus {
			d.needPostPad = true
			use(d.padCount)
		} else {
			d.needPrePad = true
			d.peakCount += d.padCount
			d.usedCount += d.padCount
		}
	}
	return d
}

func formatApply(w *sliceWriter, in []byte, d formatDetails) {
	if d.needPrePad {
		w.Fill(' ', d.padCount)
	}
	if d.first.has(bitReserved) {
		w.WriteByte('"')
	}
	marshalText(w, in, d.first)
	if d.first.has(bitReserved) {
		w.WriteByte('"')
	}
	if d.hasSecond {
		w.WriteByte(' ')
		marshalText(w, in, d.second)
	}
	if d.needPostPad {
		w.Fill(' ', d.padCount)
	}
}
