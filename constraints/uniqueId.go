package constraints

var uniqueId int

func newId() (id int) {
    id = uniqueId
    uniqueId++
    return
}
