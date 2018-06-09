package maintenance

import (
	"time"
)

//Maintenance Represant an indispo's maintenance
type Maintenance struct {
	Active     bool
	LastUpdate time.Time
}

//New Return pointer of maintenance disabled
func New() *Maintenance {
	t := time.Now()
	return &Maintenance{
		Active:     false,
		LastUpdate: t,
	}
}

//GetStatus Return true if maintenance is enabled
func (m *Maintenance) GetStatus() bool {
	return m.Active
}

//GetLastUpdate Return date of the last update
func (m *Maintenance) GetLastUpdate() time.Time {
	return m.LastUpdate
}

func (m *Maintenance) String() string {
	var s string
	if m.GetStatus() {
		s = "On"
	} else {
		s = "Off"
	}
	msg := "Maintenance is " + s + "since " + m.GetLastUpdate().String()
	return msg
}

//Enable Enable Maintenance
func (m *Maintenance) Enable() {
	if !m.GetStatus() {
		t := time.Now()
		m.LastUpdate = t
		m.Active = true
	}
}

//Disable Disable Maintenance
func (m *Maintenance) Disable() {
	if m.GetStatus() {
		t := time.Now()
		m.LastUpdate = t
		m.Active = false
	}
}
