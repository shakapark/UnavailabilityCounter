package instance

import(
	"github.com/shakapark/UnavailabilityCounter/indispo"
)

type Instance struct {
	name		string
	indispos	*indispo.Indispos
}

type Instances struct {
	list	[]*Instance
}

func New(name string) *Instance {
	return &Instance{
		name: name,
		indispos: indispo.News(),
	}
}

func (i *Instance) GetName() string {
	return i.name
}

func (i *Instance) GetIndispos() *indispo.Indispos {
	return i.indispos
}

func (i *Instance) AddIndispo(name string) {
	i.GetIndispos().Add(name)
}

func (i *Instance) GetIndispo(name string) *indispo.Indispo {
	return i.GetIndispos().GetIndispo(name)
}

func (is *Instances) GetList() []*Instance {
	return is.list
}

func (is *Instances) GetInstance(name string) *Instance {
	for _, i := range is.GetList() {
		if i.GetName() == name {
			return i
		}
	}
	return nil
}

func (is *Instances) Add(name string) {
	is.list = append(is.list, New(name))
}
