package main

import (
	"log"
	"math"
	// "fmt"
)


type Ladderer interface {
	CalculateNewPoints(winnerPoints, loserPoints int64, winnerDivision, loserDivision *Division) (winnerNew, loserNew int64)
	InitDivisions()
	GetDifference(div1, div2 *Division) int64
	getInitialPlacementMatches() int64
	// GetDivision(float64) (division *Division, position int64)
}

type Iccup struct {
}

var (
	// If difference > maxDiff, use loseX[0] or loseX[8]
	maxDiff = int64(4);
	losePoints = [5][9]int64{
		[9]int64{0, 13, 25, 37, 50, 63, 75, 100, 150}, // E
		[9]int64{0, 13, 25, 37, 50, 63, 75, 100, 150}, // D
		[9]int64{10, 19, 37, 56, 75, 93, 112, 131, 150}, // C
		[9]int64{10, 25, 50, 75, 100, 125, 150, 175, 200}, // B
		[9]int64{20, 50, 80, 110, 140, 170, 200, 230, 260}, // A
	}

	base = int64(100) // From diff = 0
	win_step = int64(25)
	win_min = float64(10)

	initialPlacementMatches = int64(0);
)

func (iccup *Iccup) InitDivisions() {

	log.Println("Initializing divisions")
	divisionCount,err := dbMap.SelectInt("SELECT COUNT(*) FROM divisions WHERE system='iccup'");
	divisions = make(Divisions, 0, divisionCount)

	divs, err := dbMap.Select(Division{}, "SELECT * FROM divisions WHERE system='iccup' ORDER BY promotion_threshold")
	if err != nil {
		panic(err)
	}

	if len(divs) > 0 {
		for x := range divs {
			div := divs[x].(*Division);
			div.Id = int64(x);  // Assign useful Id
			divisions = append(divisions, div)
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

	// We should make sure that we're using the right divisions and not the ones provided, which might not be updated correctly
	winnerDivision,_ = divisions.GetDivision(winnerPoints);
	loserDivision,_ = divisions.GetDivision(loserPoints);
	
	difference := iccup.GetDifference(winnerDivision, loserDivision) // from E=0 to A+=12
	// group_diff := winnerDivision.LadderGroup - loserDivision.LadderGroup // For new ICCUP system
	
	if(difference > maxDiff){
		difference = maxDiff; // Last element
	}else if(difference < -maxDiff){
		difference = -maxDiff; // First element
	}
	winnerNew = winnerPoints + int64(math.Max(win_min, float64(base+difference*win_step)))
	if(winnerNew > int64(divisions[len(divisions)-1].PromotionThreshold)){
		winnerNew = int64(divisions[len(divisions)-1].PromotionThreshold)
	}
	// fmt.Println("winnerpoints", math.Max(win_min, float64(base+difference*win_step)))
	// fmt.Println("difference", difference, "maxDiff", maxDiff)

	// Losing 
	// losingDiffIndex := difference
	// if(losingDiffIndex > maxDiff){
	// 	losingDiffIndex = maxDiff; // Last element
	// }else if(losingDiffIndex < -maxDiff){
	// 	losingDiffIndex = -maxDiff; // First element
	// }
	// losingDiffIndex := difference + maxDiff;

	// fmt.Println("LadderGroup", loserDivision.LadderGroup)
	// fmt.Println("losingDiffIndex",losingDiffIndex)
	// fmt.Println("points", losePoints[loserDivision.LadderGroup][losingDiffIndex])
	loserNew = loserPoints - losePoints[loserDivision.LadderGroup][difference + maxDiff]
	if(loserNew < 0){
		loserNew = 0;
	}

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
		}
		if divisions[i] == div2 {
			p2 = i
		}

		if i == 0 {
			break
		}
	}

	return p2 - p1
}

func (d *Iccup) getInitialPlacementMatches() int64 {
	return initialPlacementMatches;
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
