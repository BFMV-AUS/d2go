package memory

import (
	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/item"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/utils"
	"sort"
)

func (gd *GameReader) Items(playerPosition data.Position) data.Items {
	hoveredUnitID, hoveredType, isHovered := gd.hoveredData()

	baseAddr := gd.Process.moduleBaseAddressPtr + gd.offset.UnitTable + (4 * 1024)
	unitTableBuffer := gd.Process.ReadBytesFromMemory(baseAddr, 128*8)

	items := data.Items{}
	for i := 0; i < 128; i++ {
		itemOffset := 8 * i
		itemUnitPtr := uintptr(ReadUIntFromBuffer(unitTableBuffer, uint(itemOffset), Uint64))
		for itemUnitPtr > 0 {
			itemDataBuffer := gd.Process.ReadBytesFromMemory(itemUnitPtr, 144)
			itemType := ReadUIntFromBuffer(itemDataBuffer, 0x00, Uint32)
			txtFileNo := ReadUIntFromBuffer(itemDataBuffer, 0x04, Uint32)
			unitID := ReadUIntFromBuffer(itemDataBuffer, 0x08, Uint32)
			// itemLoc = 0 in inventory, 1 equipped, 2 in belt, 3 on ground, 4 cursor, 5 dropping, 6 socketed
			itemLoc := ReadUIntFromBuffer(itemDataBuffer, 0x0C, Uint32)

			if itemType != 4 {
				itemUnitPtr = uintptr(gd.Process.ReadUInt(itemUnitPtr+0x150, Uint64))
				continue
			}

			unitDataPtr := uintptr(ReadUIntFromBuffer(itemDataBuffer, 0x10, Uint64))
			unitDataBuffer := gd.Process.ReadBytesFromMemory(unitDataPtr, 144)
			flags := ReadUIntFromBuffer(unitDataBuffer, 0x18, Uint32)
			invPage := ReadUIntFromBuffer(unitDataBuffer, 0x55, Uint8)
			itemQuality := ReadUIntFromBuffer(unitDataBuffer, 0x00, Uint32)
			//itemOwnerNPC := ReadUIntFromBuffer(unitDataBuffer, 0x0C, Uint32)

			// Item coordinates (X, Y)
			pathPtr := uintptr(ReadUIntFromBuffer(itemDataBuffer, 0x38, Uint64))
			pathBuffer := gd.Process.ReadBytesFromMemory(pathPtr, 144)
			itemX := ReadUIntFromBuffer(pathBuffer, 0x10, Uint16)
			itemY := ReadUIntFromBuffer(pathBuffer, 0x14, Uint16)

			// Item Stats
			statsListExPtr := uintptr(ReadUIntFromBuffer(itemDataBuffer, 0x88, Uint64))
			statsListExBuffer := gd.Process.ReadBytesFromMemory(statsListExPtr, 180)
			statPtr := uintptr(ReadUIntFromBuffer(statsListExBuffer, 0x30, Uint64))
			statCount := ReadUIntFromBuffer(statsListExBuffer, 0x38, Uint32)
			statExPtr := uintptr(ReadUIntFromBuffer(statsListExBuffer, 0x88, Uint64))
			statExCount := ReadUIntFromBuffer(statsListExBuffer, 0x90, Uint32)

			stats := gd.getItemStats(statCount, statPtr, statExCount, statExPtr)

			name := item.GetNameByEnum(txtFileNo)
			itemHovered := false
			if isHovered && hoveredType == 4 && hoveredUnitID == unitID {
				itemHovered = true
			}

			itm := data.Item{
				UnitID:  data.UnitID(unitID),
				Name:    name,
				Quality: item.Quality(itemQuality),
				Position: data.Position{
					X: int(itemX),
					Y: int(itemY),
				},
				IsHovered: itemHovered,
				Stats:     stats,
			}
			setProperties(&itm, uint32(flags))

			switch itemLoc {
			case 0:
				if itm.IsVendor {
					items.Shop = append(items.Shop, itm)
				} else if invPage == 0 {
					items.Inventory = append(items.Inventory, itm)
				}
			case 1:
				items.Equipped = append(items.Equipped, itm)
			case 2:
				items.Belt = append(items.Belt, itm)
			case 3, 5:
				items.Ground = append(items.Ground, itm)
			}

			itemUnitPtr = uintptr(gd.Process.ReadUInt(itemUnitPtr+0x150, Uint64))
		}
	}

	sort.SliceStable(items.Ground, func(i, j int) bool {
		distanceI := utils.DistanceFromPoint(playerPosition, items.Ground[i].Position)
		distanceJ := utils.DistanceFromPoint(playerPosition, items.Ground[j].Position)

		return distanceI < distanceJ
	})

	return items
}

func (gd *GameReader) getItemStats(statCount uint, statPtr uintptr, statExCount uint, statExPtr uintptr) map[stat.Stat]int {
	stats := map[stat.Stat]int{}

	if statCount < 20 && statCount > 0 {
		statBuffer := gd.Process.ReadBytesFromMemory(statPtr, statCount*10)
		for i := 0; i < int(statCount); i++ {
			offset := uint(i * 8)
			statEnum := ReadUIntFromBuffer(statBuffer, offset+0x2, Uint16)
			statValue := ReadUIntFromBuffer(statBuffer, offset+0x4, Uint32)
			st, value := getStatData(statEnum, statValue)
			stats[st] = value
		}
	}

	if statExCount < 20 && statCount > 0 {
		statBuffer := gd.Process.ReadBytesFromMemory(statExPtr, statExCount*10)
		for i := 0; i < int(statExCount); i++ {
			offset := uint(i * 8)
			statEnum := ReadUIntFromBuffer(statBuffer, offset+0x2, Uint16)
			statValue := ReadUIntFromBuffer(statBuffer, offset+0x4, Uint32)
			st, value := getStatData(statEnum, statValue)
			stats[st] = value
		}
	}

	return stats
}