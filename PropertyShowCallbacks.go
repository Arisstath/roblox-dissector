package main
import "fmt"

func (x pbool) Show() string {
	return fmt.Sprintf("%v", x)
}
func (x pint) Show() string {
	return fmt.Sprintf("%d", x)
}
func (x pfloat) Show() string {
	return fmt.Sprintf("%G", x)
}
func (x pdouble) Show() string {
	return fmt.Sprintf("%G", x)
}

func (x Axes) Show() string {
	return fmt.Sprintf("%d", x)
}
func (x Faces) Show() string {
	return fmt.Sprintf("%d", x)
}

func (x BrickColor) Show() string {
	return fmt.Sprintf("%d", x)
}

func (x Object) Show() string {
	return fmt.Sprintf("%s / %d", x.Referent, x.ReferentInt)
}

func (x EnumValue) Show() string {
	return fmt.Sprintf("%d", x)
}

func (x pstring) Show() string {
	return string(x)
}

func (x BinaryString) Show() string {
	return fmt.Sprintf("... (len %d)", len([]byte(x)))
}

func (x ProtectedString) Show() string {
	return fmt.Sprintf("... (len %d)", len([]byte(x)))
}

func (x UDim) Show() string {
	return fmt.Sprintf("%G, %d", x.Scale, x.Offset)
}
func (x UDim2) Show() string {
	return fmt.Sprintf("%s, %s", x.X.Show(), x.Y.Show())
}

func (x Vector2) Show() string {
	return fmt.Sprintf("%G, %G", x.X, x.Y)
}
func (x Vector3) Show() string {
	return fmt.Sprintf("%G, %G, %G", x.X, x.Y, x.Z)
}
func (x Vector2uint16) Show() string {
	return fmt.Sprintf("%d, %d", x.X, x.Y)
}
func (x Vector3uint16) Show() string {
	return fmt.Sprintf("%d, %d, %d", x.X, x.Y, x.Z)
}

func (x Ray) Show() string {
	return fmt.Sprintf("%s, %s", x.Origin.Show(), x.Direction.Show())
}

func (x Color3) Show() string {
	return fmt.Sprintf("%G, %G, %G", x.R, x.G, x.B)
}
func (x Color3uint8) Show() string {
	return fmt.Sprintf("%d, %d, %d", x.R, x.G, x.B)
}

func (x CFrame) Show() string {
	return fmt.Sprintf("%s, matrix %G, %G, %G, %G, special %d", x.Position.Show(), x.Matrix[0], x.Matrix[1], x.Matrix[2], x.Matrix[3], x.SpecialRotMatrix)
}

func (x Content) Show() string {
	return string(x)
}

func (x SystemAddress) Show() string {
	return x.String()
}
