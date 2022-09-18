package object

type Environment struct {
	store map[string]Object
	outer *Environment
}

func NewEnvironment(env *Environment) *Environment {
	s := make(map[string]Object)
	return &Environment{store: s, outer: env}
}

func (e *Environment) Set(identifier string, object Object) Object {
	e.store[identifier] = object
	return object
}

func (e *Environment) Get(identifier string) (Object, bool) {
	o, ok := e.store[identifier]
	if ok {
		return o, true
	}
	if e.outer != nil {
		return e.outer.Get(identifier)
	}
	return nil, false
}
