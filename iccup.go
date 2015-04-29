package main

import (
	"log"
	"math"
)

type Ladderer interface {
	CalculateNewPoints(winnerPoints, loserPoints int64, winnerDivision, loserDivision *Division) (winnerNew, loserNew int64)
	InitDivisions()
	GetDifference(div1, div2 *Division) int64
	// GetDivision(float64) (division *Division, position int64)
}

type Iccup struct {
}

func (iccup *Iccup) InitDivisions() {

	log.Println("Initializing division")

	div, err := dbMap.Select(Division{}, "SELECT * FROM divisions WHERE system='iccup' ORDER BY promotion_threshold")
	if err != nil {
		panic(err)
	}

	if len(div) > 0 {
		for x := range div {
			divisions = append(divisions, div[x].(*Division))
			log.Println(div[x].(*Division))
		}
	} else {
		i := int64(0)
		for {
			if i > divisionCount {
				break
			}
			var rating float64
			if i == 0 {
				rating = 0
			} else {
				rating = divisionFirstRating + (float64(i-1) * divisionIncrements)
			}

			divisions = append(divisions, &Division{
				PromotionThreshold: rating,
				DemotionThreshold:  rating - 1,
				Name:               divisionNames[i],
				Id:                 0,
				LadderGroup:        i,
			})

			i++
		}

		for _, x := range divisions {
			err = dbMap.Insert(x)
			if err != nil {
				panic(err)
			}
		}
	}

	return
}

func (iccup *Iccup) CalculateNewPoints(winnerPoints, loserPoints int64, winnerDivision, loserDivision *Division) (winnerNew, loserNew int64) {
	// Right now win is calculated from diff = 0 and lose from diff = 4,
	// consider calculating all from diff = 4 or diff = 0

	base := int64(100) // From diff = 0
	win_step := int64(25)
	win_min := float64(0)
	// lose_base := int64(0) // From diff = 4
	// lose_step := int64(5) // From E
	lose_max := float64(0)

	difference := iccup.GetDifference(winnerDivision, loserDivision) // from E=0 to A+=12

	winnerNew = winnerPoints + int64(math.Max(win_min, float64(base+difference*win_step)))
	loserNew = loserPoints + int64(math.Min(lose_max, float64(-(10+loserDivision.LadderGroup)+(difference-3)*(10+loserDivision.LadderGroup))))

	return
}

func (iccup *Iccup) GetDifference(div1, div2 *Division) int64 {
	if div1 == nil || div2 == nil {
		return 0
	}
	var (
		p1, p2 int64
		i      = int64(len(divisions))
	)

	for {
		i--

		if divisions[i] == div1 {
			p1 = i
		} else if divisions[i] == div2 {
			p2 = i
		}

		if i == 0 {
			break
		}
	}

	return p2 - p1
}

// func (d Divisions) GetDivision(points float64) (division *Division, position int64) {
// 	i := int64(len(d))
// 	for {
// 		i--

// 		if points >= d[i].PromotionThreshold {
// 			return d[i], i
// 		}

// 		if i == 0 {
// 			break
// 		}
// 	}

// 	return nil, 0
// }
