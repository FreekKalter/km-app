package main

import "time"

type Kilometers struct {
	Id                            int64
	Date                          time.Time
	Begin, Eerste, Laatste, Terug int
	Comment                       string
}

func NewKilometers() *Kilometers {
	k := new(Kilometers)
	k.Date = time.Now().UTC()
	return k
}

func (k *Kilometers) getMax() int {
	if k.Terug > 0 {
		return k.Terug
	}
	if k.Laatste > 0 {
		return k.Laatste
	}
	if k.Eerste > 0 {
		return k.Eerste
	}
	if k.Begin > 0 {
		return k.Begin
	}
	return 0
}

func (k *Kilometers) addPost(pv PostValue) {
	switch pv.Name {
	case "begin":
		k.Begin = pv.Value
	case "eerste":
		k.Eerste = pv.Value
	case "laatste":
		k.Laatste = pv.Value
	case "terug":
		k.Terug = pv.Value
	}
}
