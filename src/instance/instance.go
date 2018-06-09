package instance

import (
	"UnavailabilityCounter/src/indispo"
)

//Instance Represent an instance with a name and an indispos ([]indispo)
type Instance struct {
	name     string
	indispos *indispo.Indispos
}

//Instances Represent a list of instance
type Instances struct {
	list []*Instance
}

//New Return a pointer of an instance with the name in parameter
func New(name string) *Instance {
	return &Instance{
		name:     name,
		indispos: indispo.News(),
	}
}

//GetName Return instance's name
func (i *Instance) GetName() string {
	return i.name
}

//GetIndispos Return instance's indispos
func (i *Instance) GetIndispos() *indispo.Indispos {
	return i.indispos
}

//AddIndispo Add an indispo to indispos with the name in parameter
func (i *Instance) AddIndispo(name string) {
	i.GetIndispos().Add(name)
}

//GetIndispo Return an indispo with the name in parameter or nil if name don't exist
func (i *Instance) GetIndispo(name string) *indispo.Indispo {
	return i.GetIndispos().GetIndispo(name)
}

//News Return a pointer of empty Instances
func News() *Instances {
	return &Instances{
		list: []*Instance{},
	}
}

//GetList Return Instance array
func (is *Instances) GetList() []*Instance {
	return is.list
}

//GetInstance Return Instance with the name in prameter,
//nil if the name isn't present
func (is *Instances) GetInstance(name string) *Instance {
	for _, i := range is.GetList() {
		if i.GetName() == name {
			return i
		}
	}
	return nil
}

//Add Add an Instance in the array
func (is *Instances) Add(name string) {
	is.list = append(is.list, New(name))
}
