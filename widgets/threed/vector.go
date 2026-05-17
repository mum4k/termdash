// threed/vector.go

package threed

import "math"

// Rotate rotates a 3D vector around the X, Y, and Z axes.
func (v Vector3D) Rotate(rot Vector3D) Vector3D {
	// Rotation around X-axis
	radX := rot.X
	cosX := math.Cos(radX)
	sinX := math.Sin(radX)
	y1 := v.Y*cosX - v.Z*sinX
	z1 := v.Y*sinX + v.Z*cosX

	// Rotation around Y-axis
	radY := rot.Y
	cosY := math.Cos(radY)
	sinY := math.Sin(radY)
	x2 := v.X*cosY + z1*sinY
	z2 := -v.X*sinY + z1*cosY

	// Rotation around Z-axis
	radZ := rot.Z
	cosZ := math.Cos(radZ)
	sinZ := math.Sin(radZ)
	x3 := x2*cosZ - y1*sinZ
	y3 := x2*sinZ + y1*cosZ

	return Vector3D{
		X: x3,
		Y: y3,
		Z: z2,
	}
}

// Dot computes the dot product of two vectors.
func (v Vector3D) Dot(other Vector3D) float64 {
	return v.X*other.X + v.Y*other.Y + v.Z*other.Z
}

// Subtract subtracts another vector from this vector.
func (v Vector3D) Subtract(other Vector3D) Vector3D {
	return Vector3D{
		X: v.X - other.X,
		Y: v.Y - other.Y,
		Z: v.Z - other.Z,
	}
}

// Cross computes the cross product of two vectors.
func (v Vector3D) Cross(other Vector3D) Vector3D {
	return Vector3D{
		X: v.Y*other.Z - v.Z*other.Y,
		Y: v.Z*other.X - v.X*other.Z,
		Z: v.X*other.Y - v.Y*other.X,
	}
}

// Normalize returns a unit vector in the same direction.
func (v Vector3D) Normalize() Vector3D {
	lenSq := v.X*v.X + v.Y*v.Y + v.Z*v.Z
	if lenSq == 0 {
		return Vector3D{}
	}
	length := math.Sqrt(lenSq)
	return Vector3D{X: v.X / length, Y: v.Y / length, Z: v.Z / length}
}

// Multiply multiplies the vector by a scalar.
func (v Vector3D) Multiply(scalar float64) Vector3D {
	return Vector3D{
		X: v.X * scalar,
		Y: v.Y * scalar,
		Z: v.Z * scalar,
	}
}

// Add adds another vector to this vector.
func (v Vector3D) Add(other Vector3D) Vector3D {
	return Vector3D{
		X: v.X + other.X,
		Y: v.Y + other.Y,
		Z: v.Z + other.Z,
	}
}

// RotationMatrix is a pre-computed 3×3 rotation matrix that applies the same
// X→Y→Z Euler rotation as Vector3D.Rotate but at the cost of only 6 trig
// calls per frame instead of 6 per vertex.
type RotationMatrix [3][3]float64

// BuildRotationMatrix constructs a combined X→Y→Z rotation matrix from the
// given Euler angles (in radians). The resulting matrix matches the sequential
// application performed by Vector3D.Rotate.
//
// Derivation:
//
//	cx,sx = cos/sin(rot.X),  cy,sy = cos/sin(rot.Y),  cz,sz = cos/sin(rot.Z)
//	R = [[ cy*cz,  sx*sy*cz - cx*sz,  cx*sy*cz + sx*sz ],
//	     [ cy*sz,  sx*sy*sz + cx*cz,  cx*sy*sz - sx*cz ],
//	     [ -sy,    sx*cy,             cx*cy             ]]
func BuildRotationMatrix(rot Vector3D) RotationMatrix {
	cx, sx := math.Cos(rot.X), math.Sin(rot.X)
	cy, sy := math.Cos(rot.Y), math.Sin(rot.Y)
	cz, sz := math.Cos(rot.Z), math.Sin(rot.Z)
	return RotationMatrix{
		{cy * cz, sx*sy*cz - cx*sz, cx*sy*cz + sx*sz},
		{cy * sz, sx*sy*sz + cx*cz, cx*sy*sz - sx*cz},
		{-sy, sx * cy, cx * cy},
	}
}

// Apply rotates vector v by the pre-computed rotation matrix using 9
// multiplications and no trigonometry.
func (m RotationMatrix) Apply(v Vector3D) Vector3D {
	return Vector3D{
		X: m[0][0]*v.X + m[0][1]*v.Y + m[0][2]*v.Z,
		Y: m[1][0]*v.X + m[1][1]*v.Y + m[1][2]*v.Z,
		Z: m[2][0]*v.X + m[2][1]*v.Y + m[2][2]*v.Z,
	}
}
