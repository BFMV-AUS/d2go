package data

import (
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"strings"
)

type Items struct {
	Belt      Belt
	Inventory Inventory
	Shop      []Item
	Ground    []Item
	Equipped  []Item
}

type Inventory []Item
type UnitID int

type Item struct {
	UnitID
	Name       item.Name
	Quality    item.Quality
	Position   Position
	Ethereal   bool
	IsHovered  bool
	Stats      map[stat.Stat]int
	Identified bool
	IsVendor   bool
}

func (i Item) IsPotion() bool {
	return i.IsHealingPotion() || i.IsManaPotion() || i.IsRejuvPotion()
}

func (i Item) IsHealingPotion() bool {
	return strings.Contains(string(i.Name), string(HealingPotion))
}

func (i Item) IsManaPotion() bool {
	return strings.Contains(string(i.Name), string(ManaPotion))
}
func (i Item) IsRejuvPotion() bool {
	return strings.Contains(string(i.Name), string(RejuvenationPotion))
}
