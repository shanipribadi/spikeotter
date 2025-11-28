package spikeotter

import (
	"math/rand"
	"sync"
)

type Rand struct {
	l sync.Mutex
	z *rand.Zipf
}

func NewRand(imax uint64) *Rand {
	r := rand.New(rand.NewSource(rand.Int63()))
	z := rand.NewZipf(r, 1.01, 1.01, imax)
	return &Rand{
		z: z,
	}
}

func (r *Rand) Int() int {
	r.l.Lock()
	defer r.l.Unlock()
	return int(r.z.Uint64())
}
