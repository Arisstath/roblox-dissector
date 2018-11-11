package bitstreams

type Reference struct {
    Scope string
    Id uint32
    IsNull bool
}

func NewReference(scope string, id uint32) *Reference {
    return &Reference{IsNull: id == 0, Scope: scope, Id: id}
}
