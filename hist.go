package main

import (
	"math"
)

type Hist struct {
	SampleRate int
	BufferLen  int
	Bars       []*Bar
	FilterA    float64
	FilterN    float64
	filterAd   float64
	step       float64
}

type Bar struct {
	Color       string
	Set         []int
	expandedSet []int

	level    float64
	avgLevel float64
}

func (h *Hist) Init() {
	h.filterAd = 1 - h.FilterA
	h.step = float64(h.SampleRate) / float64(h.BufferLen)
	for _, bar := range h.Bars {
		bar.expandSet()
		bar.avgLevel = 1
	}
}

func (h *Hist) Update(levels []float64) {
	for _, bar := range h.Bars {
		bar.update(levels)
		bar.normalize(h)
	}
}

func (h *Hist) Draw(barDrawer func(*Bar)) {
	for _, bar := range h.Bars {
		barDrawer(bar)
	}
}

func (ch *Bar) expandSet() {
	set := ch.Set
	expanded := make([]int, 0, len(set))
	last := 0
	for _, cur := range set {
		if cur < 0 {
			for last++; last <= -cur; last++ {
				expanded = append(expanded, last)
			}
		} else {
			last = cur
			expanded = append(expanded, last)
		}
	}
	ch.expandedSet = expanded
}

func (bar *Bar) Levels() (level, avg float64) {
	return bar.level, bar.avgLevel
}

func (bar *Bar) update(levels []float64) {
	// вычисление среднего уровня для набора частот
	bar.level = 0
	for _, i := range bar.expandedSet {
		bar.level += levels[i]
	}
	bar.level /= float64(len(bar.expandedSet))
}

// Приведение уровня сигнала к диапазону [0; 1], небольшая постфильтрация
func (ch *Bar) normalize(cnf *Hist) {
	// Максимальный уровень одного отсчета math.MaxInt16
	// Анализируется cnf.BufferLen отсчетов
	// Зеркальная половина отбрасывается
	ch.level /= float64(math.MaxInt16 / 2 * cnf.BufferLen)

	// Коррекция среднего по времени уровня сигнала ab фильтром
	ch.avgLevel = cnf.filterAd*ch.avgLevel + cnf.FilterA*ch.level

	// Коррекция уровня сигнала, исходя из того, что средний сигнал
	// принимается за половину? выходного
	ch.level /= ch.avgLevel * cnf.FilterN

	// Обрезка уровня сигнала
	if ch.level > 1 {
		ch.level = 1
	}
}
