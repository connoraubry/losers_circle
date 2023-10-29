package set

type Set struct {
	set map[string]bool
}

func New() *Set {
	s := &Set{}
	s.set = make(map[string]bool)
	return s
}

func (s *Set) Add(name string) {
	s.set[name] = true
}
func (s *Set) Remove(name string) {
	if s.IsIn(name) {
		delete(s.set, name)
	}
}
func (s *Set) IsIn(name string) bool {
	_, ok := s.set[name]
	return ok
}
func (s *Set) Len() int {
	return len(s.set)
}
