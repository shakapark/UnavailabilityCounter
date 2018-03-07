package maintenance

import(
	"time"
)

type Maintenance struct {
	Active     bool
	LastUpdate time.Time
}

func New() *Maintenance {
	t := time.Now()
	return &Maintenance{
		Active: false,
		LastUpdate: t,
	}
}

func (m *Maintenance) GetStatus() bool {
	return m.Active
}

func (m *Maintenance) GetLastUpdate() time.Time {
	return m.LastUpdate
}

func (m *Maintenance) String() string {
	var s string
	if m.IsEnable() { s = "On" }
	else { s = "Off" }
	msg := "Maintenance is " + s + "since " + m.GetLastUpdate().String()
	return msg
}

func (m *Maintenance) IsEnable() bool {
	if m.GetStatus() { return true }
	else { return false }
}

func (m *Maintenance) Enable() {
	if !m.IsEnable() {
		t := time.Now()
		m.LastUpdate = t
		m.Active = true
	}
}

func (m *Maintenance) Disable() {
	if m.IsEnable() {
		t := time.Now()
		m.LastUpdate = t
		m.Active = false
	}
}
