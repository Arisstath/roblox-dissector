package main
import "fmt"
import "bytes"

func (x pbool) Show() string {
	return fmt.Sprintf("%v", x)
}
func (x psint) Show() string {
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
func (x RebindObject) Show() string {
	return fmt.Sprintf("%d / %d", x.Referent1, x.Referent2)
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

func (x Region3) Show() string {
	return fmt.Sprintf("%s, %s", x.Start.Show(), x.End.Show())
}
func (x Region3uint16) Show() string {
	return fmt.Sprintf("%s, %s", x.Start.Show(), x.End.Show())
}

func (x Instance) Show() string {
	return Object(x).Show()
}

func (x Tuple) Show() string {
	var ret bytes.Buffer
	ret.WriteString("[")

	for _, y := range x {
		ret.WriteString(fmt.Sprintf("(%s) %s, ", y.Type, y.Value))
	}

	ret.WriteString("]")
	return ret.String()
}

func (x Array) Show() string {
	return Tuple(x).Show()
}

func (x Dictionary) Show() string {
	var ret bytes.Buffer
	ret.WriteString("{")

	for k, v := range x {
		ret.WriteString(fmt.Sprintf("%s: (%s) %s, ", k, v.Type, v.Value))
	}

	ret.WriteString("}")
	return ret.String()
}

func (x Map) Show() string {
	return Dictionary(x).Show()
}
