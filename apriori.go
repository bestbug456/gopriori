package gopriori

import "sync"

// Apriori is the struct responsible of contain the support map
// for each combination which is grater than the imposed threshold
type Apriori struct {
	Supportmap       map[string]int
	original         []string
	TotalTransaction int
}

// Support return the support of a given ID, if the id is not found
// then -1 is returned
func (a *Apriori) Support(ids []string) float64 {
	id := a.generateID(ids)
	_, exist := a.Supportmap[id]
	if !exist {
		return -1
	}
	return float64(a.Supportmap[id]) / float64(a.TotalTransaction)
}

// Confidence return the confidence of X result to Y, if the id is not found
// then -1 is returned
func (a *Apriori) Confidence(X, Y []string) float64 {
	idx := a.generateID(X)
	idy := a.generateID(Y)
	combined := idx + " " + idy
	support1, exist := a.Supportmap[idx]
	if !exist {
		return -1
	}
	support2, exist := a.Supportmap[combined]
	if !exist {
		return -1
	}
	return float64(support2) / float64(support1)
}

// Lift return the lift of X to Y (how likely X mean Y), if the id is not found
// then -1 is returned
func (a *Apriori) Lift(X, Y []string) float64 {
	idx := a.generateID(X)
	idy := a.generateID(Y)
	combined := idx + " " + idy
	support1, exist := a.Supportmap[idx]
	if !exist {
		return -1
	}
	support2, exist := a.Supportmap[idy]
	if !exist {
		return -1
	}
	support3, exist := a.Supportmap[combined]
	if !exist {
		return -1
	}
	return float64(support3) / (float64(support1) * float64(support2))
}

// Delete remove a rule from the dataset, if exist
func (a *Apriori) Delete(ids []string) {
	ID := a.generateID(ids)
	delete(a.Supportmap, ID)
}

func (a *Apriori) generateID(ids []string) string {
	var id string
	for i := 0; i < len(a.original); i++ {
		for y := 0; y < len(ids); y++ {
			if a.original[i] == ids[y] {
				if id == "" {
					id += ids[y]
				} else {
					id += " " + ids[y]
				}
			}
		}
	}
	return id
}

// Threshold contain al the information which can be used for filtering the result
type Threshold struct {
	SupportThreshold int
}

// Train permit to pass a list of transaction. Each transaction should
// have one or more string which permit to identify an item, notice this
// implementation of apriori census each string with a unique integer identifier
// this with a small initial fee this permit to reduce the train time
func Train(transaction [][]string, configuration Threshold) *Apriori {

	starter, matrix := Compression(transaction)
	idmap := make(map[string][]int)
	original := make([]string, len(starter))
	var i int
	for id, value := range starter {
		// copy by value, not by reference
		idmap[id] = []int{value[0]}
		original[i] = id
		i++
	}
	supportmap := make(map[string]int)
	for {
		idmap, supportmap = calculateSupportMap(idmap, supportmap, matrix, configuration)
		idmap = combination(starter, idmap)
		if len(idmap) == 0 {
			break
		}
	}
	return &Apriori{
		Supportmap:       supportmap,
		TotalTransaction: len(transaction),
		original:         original,
	}
}

func calculateSupportMap(idmap map[string][]int, supportmap map[string]int, matrix [][]int, configuration Threshold) (map[string][]int, map[string]int) {

	survivedID := make(map[string][]int)
	supportcounterchan := make(chan supportcounter, len(idmap))
	var wg sync.WaitGroup
	for ID, valueID := range idmap {
		wg.Add(1)
		go calculateSingleSupport(ID, valueID, matrix, configuration, supportcounterchan, &wg)
	}
	wg.Wait()
	for {
		select {
		case msg := <-supportcounterchan:
			supportmap[msg.ID] = msg.counter
			survivedID[msg.ID] = msg.valueID
		default:
			return survivedID, supportmap
		}
	}
}

// supportcounter is used for send message from the main routine when the analysis is finished
type supportcounter struct {
	counter int
	ID      string
	valueID []int
}

func calculateSingleSupport(ID string, valueID []int, matrix [][]int, configuration Threshold, results chan supportcounter, wg *sync.WaitGroup) {
	defer wg.Done()
	var totalsupport int
	for i := 0; i < len(matrix); i++ {
		var cnt int
		for y := 0; y < len(matrix[i]); y++ {
			for k := 0; k < len(valueID); k++ {
				if matrix[i][y] == valueID[k] {
					cnt++
				}
			}
		}
		if cnt == len(valueID) {
			totalsupport++
		}
	}
	if totalsupport >= configuration.SupportThreshold && totalsupport != 0 {
		results <- supportcounter{
			counter: totalsupport,
			ID:      ID,
			valueID: valueID,
		}
	}
}

func combination(starter, idmap map[string][]int) map[string][]int {
	combinedID := make(map[string][]int)
	for ID, original := range starter {
		for ID2, value2 := range idmap {
			if ID == ID2 {
				continue
			}
			var founded bool
			for y := 0; y < len(value2); y++ {
				if original[0] == value2[y] {
					founded = true
				}
				if founded {
					break
				}
			}
			if founded {
				continue
			}
			cp := make([]int, len(value2)+1)
			copy(cp, value2)
			cp[len(value2)] = original[0]
			combinedID[ID2+" "+ID] = cp
		}
	}
	return combinedID
}

// Compression permit to compress data to integer rappresentation
func Compression(transaction [][]string) (map[string][]int, [][]int) {
	maxid := 1
	conversionmap := make(map[string][]int)
	matrix := make([][]int, len(transaction))
	for i := 0; i < len(transaction); i++ {
		matrix[i] = make([]int, len(transaction[i]))
		for y := 0; y < len(transaction[i]); y++ {
			_, ok := conversionmap[transaction[i][y]]
			if !ok {
				conversionmap[transaction[i][y]] = []int{maxid}
				maxid++
			}
			matrix[i][y] = conversionmap[transaction[i][y]][0]
		}
	}
	return conversionmap, matrix
}
