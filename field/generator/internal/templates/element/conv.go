package element

const Conv = `

// rSquare where r is the Montgommery constant
// see section 2.3.2 of Tolga Acar's thesis
// https://www.microsoft.com/en-us/research/wp-content/uploads/1998/06/97Acar.pdf
var rSquare = {{.ElementName}}{
	{{- range $i := .RSquare}}
	{{$i}},{{end}}
}

// toMont converts z to Montgomery form
// sets and returns z = z * r²
func (z *{{.ElementName}}) toMont() *{{.ElementName}} {
	return z.Mul(z, &rSquare)
}

// String returns the decimal representation of z as generated by
// z.Text(10).
func (z *{{.ElementName}}) String() string {
	return z.Text(10)
}

// toBigInt returns z as a big.Int in Montgomery form
func (z *{{.ElementName}}) toBigInt(res *big.Int) *big.Int {
       var b [Bytes]byte
       {{- range $i := reverse .NbWordsIndexesFull}}
               {{- $j := mul $i 8}}
               {{- $k := sub $.NbWords 1}}
               {{- $k := sub $k $i}}
               {{- $jj := add $j 8}}
               binary.BigEndian.PutUint64(b[{{$j}}:{{$jj}}], z[{{$k}}])
       {{- end}}

       return res.SetBytes(b[:])
}

// Text returns the string representation of z in the given base.
// Base must be between 2 and 36, inclusive. The result uses the
// lower-case letters 'a' to 'z' for digit values 10 to 35.
// No prefix (such as "0x") is added to the string. If z is a nil
// pointer it returns "<nil>".
// If base == 10 and -z fits in a uint16 prefix "-" is added to the string.
func (z *{{.ElementName}}) Text(base int) string {
	if base < 2 || base > 36 {
		panic("invalid base")
	}
	if z == nil {
		return "<nil>"
	}

	const maxUint16 = 65535
	{{- if eq $.NbWords 1}}
		if base == 10 {
			var zzNeg {{.ElementName}}
			zzNeg.Neg(z)
			zzNeg.fromMont()
			if zzNeg[0] <= maxUint16 && zzNeg[0] != 0 {
				return "-" + strconv.FormatUint(zzNeg[0], base)
			}
		}
		zz := z.Bits()
		return strconv.FormatUint(zz[0], base)
	{{- else }}
		if base == 10 {
			var zzNeg {{.ElementName}}
			zzNeg.Neg(z)
			zzNeg.fromMont()
			if zzNeg.FitsOnOneWord() && zzNeg[0] <= maxUint16 && zzNeg[0] != 0  {
				return "-" + strconv.FormatUint(zzNeg[0], base)
			}
		}
		zz := *z
		zz.fromMont()
		if zz.FitsOnOneWord() {
			return strconv.FormatUint(zz[0], base)
		} 
		vv := pool.BigInt.Get()
		r := zz.toBigInt(vv).Text(base)
		pool.BigInt.Put(vv)
		return r
	{{- end}}
}

// BigInt sets and return z as a *big.Int
func (z *{{.ElementName}}) BigInt(res *big.Int) *big.Int {
	_z := *z
	_z.fromMont()
	return _z.toBigInt(res)
}

// ToBigIntRegular returns z as a big.Int in regular form
// 
// Deprecated: use BigInt(*big.Int) instead
func (z {{.ElementName}}) ToBigIntRegular(res *big.Int) *big.Int {
	z.fromMont()
	return z.toBigInt(res)
}

// Bits provides access to z by returning its value as a little-endian [{{.NbWords}}]uint64 array. 
// Bits is intended to support implementation of missing low-level {{.ElementName}}
// functionality outside this package; it should be avoided otherwise.
func (z *{{.ElementName}}) Bits() [{{.NbWords}}]uint64 {
	_z := *z
	fromMont(&_z)
	return _z
}

// Bytes returns the value of z as a big-endian byte array
func (z *{{.ElementName}}) Bytes() (res [Bytes]byte) {
	BigEndian.PutElement(&res, *z)
	return
}

// Marshal returns the value of z as a big-endian byte slice
func (z *{{.ElementName}}) Marshal() []byte {
	b := z.Bytes()
	return b[:]
}

// SetBytes interprets e as the bytes of a big-endian unsigned integer,
// sets z to that value, and returns z.
func (z *{{.ElementName}}) SetBytes(e []byte) *{{.ElementName}} {
	if len(e) == Bytes {
		// fast path
		v, err := BigEndian.Element((*[Bytes]byte)(e))
		if err == nil {
			*z = v
			return z 
		}
	}

	// slow path.
	// get a big int from our pool
	vv := pool.BigInt.Get()
	vv.SetBytes(e)

	// set big int
	z.SetBigInt(vv)

	// put temporary object back in pool
	pool.BigInt.Put(vv)

	return z 
}

// SetBytesCanonical interprets e as the bytes of a big-endian {{.NbBytes}}-byte integer.
// If e is not a {{.NbBytes}}-byte slice or encodes a value higher than q, 
// SetBytesCanonical returns an error.
func (z *{{.ElementName}}) SetBytesCanonical(e []byte) error {
	if len(e) != Bytes {
		return errors.New("invalid {{.PackageName}}.{{.ElementName}} encoding")
	}
	v, err := BigEndian.Element((*[Bytes]byte)(e))
	if err != nil {
		return err
	}
	*z = v
	return nil
}


// SetBigInt sets z to v and returns z
func (z *{{.ElementName}}) SetBigInt(v *big.Int) *{{.ElementName}} {
	z.SetZero()

	var zero big.Int

	// fast path
	c := v.Cmp(&_modulus)
	if c == 0 {
		// v == 0
		return z
	} else if c != 1 && v.Cmp(&zero) != -1 {
		// 0 < v < q
		return z.setBigInt(v)
	}

	// get temporary big int from the pool
	vv := pool.BigInt.Get()

	// copy input + modular reduction
	vv.Mod(v, &_modulus)

	// set big int byte value
	z.setBigInt(vv)

	// release object into pool
	pool.BigInt.Put(vv)
	return z
}

// setBigInt assumes 0 ⩽ v < q
func (z *{{.ElementName}}) setBigInt(v *big.Int) *{{.ElementName}} {
	vBits := v.Bits()

	if bits.UintSize == 64 {
		for i := 0; i < len(vBits); i++ {
			z[i] = uint64(vBits[i])
		}
	} else {
		for i := 0; i < len(vBits); i++ {
			if i%2 == 0 {
				z[i/2] = uint64(vBits[i])
			} else {
				z[i/2] |= uint64(vBits[i]) << 32
			}
		}
	}

	return z.toMont()
}

// SetString creates a big.Int with number and calls SetBigInt on z
//
// The number prefix determines the actual base: A prefix of
// ''0b'' or ''0B'' selects base 2, ''0'', ''0o'' or ''0O'' selects base 8,
// and ''0x'' or ''0X'' selects base 16. Otherwise, the selected base is 10
// and no prefix is accepted.
//
// For base 16, lower and upper case letters are considered the same:
// The letters 'a' to 'f' and 'A' to 'F' represent digit values 10 to 15.
//
// An underscore character ''_'' may appear between a base
// prefix and an adjacent digit, and between successive digits; such
// underscores do not change the value of the number.
// Incorrect placement of underscores is reported as a panic if there
// are no other errors.
//
// If the number is invalid this method leaves z unchanged and returns nil, error.
func (z *{{.ElementName}}) SetString(number string) (*{{.ElementName}}, error) {
	// get temporary big int from the pool
	vv := pool.BigInt.Get()

	if _, ok := vv.SetString(number, 0); !ok {
		return nil, errors.New("{{.ElementName}}.SetString failed -> can't parse number into a big.Int " + number)
	}

	z.SetBigInt(vv)

	// release object into pool
	pool.BigInt.Put(vv)

	return z, nil
}


// MarshalJSON returns json encoding of z (z.Text(10))
// If z == nil, returns null
func (z *{{.ElementName}}) MarshalJSON() ([]byte, error) {
	if z == nil {
		return []byte("null"), nil
	}
	const maxSafeBound = 15 // we encode it as number if it's small
	s := z.Text(10)
	if len(s) <= maxSafeBound {
		return []byte(s), nil
	}
	var sbb strings.Builder
	sbb.WriteByte('"')
	sbb.WriteString(s)
	sbb.WriteByte('"')
	return []byte(sbb.String()), nil
}

// UnmarshalJSON accepts numbers and strings as input
// See {{.ElementName}}.SetString for valid prefixes (0x, 0b, ...)
func (z *{{.ElementName}}) UnmarshalJSON(data []byte) error {
	s := string(data)
	if len(s) > Bits*3 {
		return errors.New("value too large (max = {{.ElementName}}.Bits * 3)")
	}

	// we accept numbers and strings, remove leading and trailing quotes if any
	if len(s) > 0 && s[0] == '"' {
		s = s[1:]
	}
	if len(s) > 0 && s[len(s)-1] == '"' {
		s = s[:len(s)-1]
	}

	// get temporary big int from the pool
	vv := pool.BigInt.Get()

	if _, ok := vv.SetString(s, 0); !ok {
		return errors.New("can't parse into a big.Int: " + s)
	}

	z.SetBigInt(vv)

	// release object into pool
	pool.BigInt.Put(vv)
	return nil
}


// A ByteOrder specifies how to convert byte slices into a {{.ElementName}}
type ByteOrder interface {
	Element(*[Bytes]byte) ({{.ElementName}}, error)
	PutElement(*[Bytes]byte, {{.ElementName}})
	String() string
}


// BigEndian is the big-endian implementation of ByteOrder and AppendByteOrder.
var BigEndian bigEndian

type bigEndian struct{}

// Element interpret b is a big-endian {{.NbBytes}}-byte slice.
// If b encodes a value higher than q, Element returns error.
func (bigEndian) Element(b *[Bytes]byte) ({{.ElementName}}, error) {
	var z {{.ElementName}}
	{{- range $i := reverse .NbWordsIndexesFull}}
		{{- $j := mul $i 8}}
		{{- $k := sub $.NbWords 1}}
		{{- $k := sub $k $i}}
		{{- $jj := add $j 8}}
		z[{{$k}}] = binary.BigEndian.Uint64((*b)[{{$j}}:{{$jj}}])
	{{- end}}

	if !z.smallerThanModulus() {
		return {{.ElementName}}{}, errors.New("invalid {{.PackageName}}.{{.ElementName}} encoding")
	}

	z.toMont()
	return z, nil
}

func (bigEndian) PutElement(b *[Bytes]byte, e {{.ElementName}})  {
	e.fromMont()

	{{- range $i := reverse .NbWordsIndexesFull}}
		{{- $j := mul $i 8}}
		{{- $k := sub $.NbWords 1}}
		{{- $k := sub $k $i}}
		{{- $jj := add $j 8}}
		binary.BigEndian.PutUint64((*b)[{{$j}}:{{$jj}}], e[{{$k}}])
	{{- end}}
}

func (bigEndian) String() string { return "BigEndian" }



// LittleEndian is the little-endian implementation of ByteOrder and AppendByteOrder.
var LittleEndian littleEndian

type littleEndian struct{}

func (littleEndian) Element(b *[Bytes]byte) ({{.ElementName}}, error) {
	var z {{.ElementName}}
	{{- range $i := .NbWordsIndexesFull}}
		{{- $j := mul $i 8}}
		{{- $jj := add $j 8}}
		z[{{$i}}] = binary.LittleEndian.Uint64((*b)[{{$j}}:{{$jj}}])
	{{- end}}

	if !z.smallerThanModulus() {
		return {{.ElementName}}{}, errors.New("invalid {{.PackageName}}.{{.ElementName}} encoding")
	}

	z.toMont()
	return z, nil
}

func (littleEndian) PutElement(b *[Bytes]byte, e {{.ElementName}})  {
	e.fromMont()

	{{- range $i := .NbWordsIndexesFull}}
		{{- $j := mul $i 8}}
		{{- $jj := add $j 8}}
		binary.LittleEndian.PutUint64((*b)[{{$j}}:{{$jj}}], e[{{$i}}])
	{{- end}}
}

func (littleEndian) String() string { return "LittleEndian" }




`
